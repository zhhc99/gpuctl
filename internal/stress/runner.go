package stress

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
	_ "github.com/gogpu/wgpu/hal/allbackends"
)

type Mode string

const (
	ModeALU   Mode = "alu"
	ModeMem   Mode = "mem"
	ModeMixed Mode = "mixed"
)

func ValidMode(s string) bool {
	m := Mode(s)
	return m == ModeALU || m == ModeMem || m == ModeMixed
}

type Stats struct {
	Name   string
	Mode   Mode
	Buffer uint64

	total    atomic.Int64
	start    time.Time
	lastRead int64
	lastTime time.Time
}

func (s *Stats) inc()                   { s.total.Add(1) }
func (s *Stats) Total() int64           { return s.total.Load() }
func (s *Stats) Elapsed() time.Duration { return time.Since(s.start) }

// Rate 返回上次调用到现在的瞬时 dispatch/s. 仅由 display goroutine 调用.
func (s *Stats) Rate() int64 {
	now := time.Now()
	cur := s.total.Load()
	dt := now.Sub(s.lastTime).Seconds()
	if dt < 0.01 {
		return 0
	}
	rate := int64(float64(cur-s.lastRead) / dt)
	s.lastRead = cur
	s.lastTime = now
	return rate
}

func (s *Stats) AvgRate() int64 {
	e := s.Elapsed().Seconds()
	if e < 0.01 {
		return 0
	}
	return int64(float64(s.total.Load()) / e)
}

type Runner struct {
	vramBytes uint64
	stats     Stats

	instance *wgpu.Instance
	adapter  *wgpu.Adapter
	device   *wgpu.Device
}

func NewRunner(vramSpec string, mode Mode) (*Runner, error) {
	b, err := parseVRAMBytes(vramSpec)
	if err != nil {
		return nil, err
	}
	return &Runner{
		vramBytes: b,
		stats:     Stats{Mode: mode, Buffer: b},
	}, nil
}

// Init creates the wgpu device and resolves the GPU name.
// Must be called before Run(). Stats().Name is valid after this returns.
func (r *Runner) Init() error {
	var err error
	r.instance, err = wgpu.CreateInstance(nil)
	if err != nil {
		return fmt.Errorf("wgpu instance: %w", err)
	}

	r.adapter, err = r.instance.RequestAdapter(nil)
	if err != nil {
		r.instance.Release()
		return fmt.Errorf("gpu adapter: %w", err)
	}

	r.stats.Name = r.adapter.Info().Name

	r.device, err = r.adapter.RequestDevice(nil)
	if err != nil {
		r.adapter.Release()
		r.instance.Release()
		return fmt.Errorf("gpu device: %w", err)
	}

	now := time.Now()
	r.stats.start = now
	r.stats.lastTime = now
	return nil
}

func (r *Runner) Close() {
	if r.device != nil {
		r.device.Release()
	}
	if r.adapter != nil {
		r.adapter.Release()
	}
	if r.instance != nil {
		r.instance.Release()
	}
}

func (r *Runner) Stats() *Stats { return &r.stats }

func (r *Runner) Run(ctx context.Context) error {
	return r.dispatchLoop(ctx, r.device)
}

// maxBufSize 是驱动常见的单次 buffer 分配上限.
const maxBufSize = 256 * 1024 * 1024

func (r *Runner) dispatchLoop(ctx context.Context, device *wgpu.Device) error {
	bgl, err := device.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label: "stress-bgl",
		Entries: []wgpu.BindGroupLayoutEntry{{
			Binding:    0,
			Visibility: wgpu.ShaderStageCompute,
			Buffer:     &gputypes.BufferBindingLayout{Type: gputypes.BufferBindingTypeStorage},
		}},
	})
	if err != nil {
		return fmt.Errorf("bind group layout: %w", err)
	}
	defer bgl.Release()

	pl, err := device.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		Label:            "stress-pl",
		BindGroupLayouts: []*wgpu.BindGroupLayout{bgl},
	})
	if err != nil {
		return fmt.Errorf("pipeline layout: %w", err)
	}
	defer pl.Release()

	pipelines, err := r.buildPipelines(device, pl)
	if err != nil {
		return err
	}
	for _, p := range pipelines {
		defer p.Release()
	}

	// 按 maxBufSize 分片, 满足驱动单次分配上限.
	sliceSize := uint64(maxBufSize)
	if sliceSize > r.vramBytes {
		sliceSize = r.vramBytes
	}
	nSlices := (r.vramBytes + sliceSize - 1) / sliceSize

	bufs := make([]*wgpu.Buffer, nSlices)
	bgs := make([]*wgpu.BindGroup, nSlices)
	for i := range nSlices {
		sz := sliceSize
		if uint64(i)*sliceSize+sz > r.vramBytes {
			sz = r.vramBytes - uint64(i)*sliceSize
		}
		// CopySrc 是同步屏障所需: 每次 dispatch 后从该 buffer copy 4 字节到 syncBuf 触发 GPU 等待.
		bufs[i], err = device.CreateBuffer(&wgpu.BufferDescriptor{
			Label: fmt.Sprintf("stress-buf-%d", i),
			Size:  sz,
			Usage: wgpu.BufferUsageStorage | wgpu.BufferUsageCopySrc,
		})
		if err != nil {
			for j := range i {
				bufs[j].Release()
			}
			return fmt.Errorf("vram alloc [%d]: %w", i, err)
		}
		bgs[i], err = device.CreateBindGroup(&wgpu.BindGroupDescriptor{
			Label:  fmt.Sprintf("stress-bg-%d", i),
			Layout: bgl,
			Entries: []wgpu.BindGroupEntry{{
				Binding: 0,
				Buffer:  bufs[i],
				Size:    sz,
			}},
		})
		if err != nil {
			for j := range i + 1 {
				bufs[j].Release()
			}
			for j := range i {
				bgs[j].Release()
			}
			return fmt.Errorf("bind group [%d]: %w", i, err)
		}
	}
	defer func() {
		for _, b := range bufs {
			b.Release()
		}
		for _, b := range bgs {
			b.Release()
		}
	}()

	// syncBuf 是 GPU 同步屏障: dispatch 结束后 copy 4 字节到此, ReadBuffer 阻塞直到 GPU 完成.
	// 这将 in-flight 队列深度限制为 1, Ctrl+C 后立即响应.
	syncBuf, err := device.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "stress-sync",
		Size:  4,
		Usage: wgpu.BufferUsageCopyDst | wgpu.BufferUsageMapRead,
	})
	if err != nil {
		return fmt.Errorf("sync buffer: %w", err)
	}
	defer syncBuf.Release()
	syncTmp := make([]byte, 4)

	wgs := uint32(sliceSize / 4 / 256)
	if wgs > 65535 {
		wgs = 65535
	}
	if wgs == 0 {
		wgs = 1
	}

	queue := device.Queue()
	np := len(pipelines)
	nb := int(nSlices)
	for i := 0; ctx.Err() == nil; i++ {
		enc, err := device.CreateCommandEncoder(nil)
		if err != nil {
			return fmt.Errorf("command encoder: %w", err)
		}

		pass, err := enc.BeginComputePass(nil)
		if err != nil {
			return fmt.Errorf("compute pass: %w", err)
		}
		pass.SetPipeline(pipelines[i%np])
		pass.SetBindGroup(0, bgs[i%nb], nil)
		pass.Dispatch(wgs, 1, 1)
		if err := pass.End(); err != nil {
			return fmt.Errorf("pass end: %w", err)
		}

		// dispatch 结束后 copy 4 字节到 syncBuf, 建立同步点.
		enc.CopyBufferToBuffer(bufs[i%nb], 0, syncBuf, 0, 4)

		cmd, err := enc.Finish()
		if err != nil {
			return fmt.Errorf("encoder finish: %w", err)
		}
		if err := queue.Submit(cmd); err != nil {
			return fmt.Errorf("queue submit: %w", err)
		}

		// 阻塞直到本次 dispatch 在 GPU 上完成, 队列深度始终为 1.
		if err := queue.ReadBuffer(syncBuf, 0, syncTmp); err != nil {
			return fmt.Errorf("sync readback: %w", err)
		}

		r.stats.inc()
	}
	return nil
}

func (r *Runner) buildPipelines(device *wgpu.Device, pl *wgpu.PipelineLayout) ([]*wgpu.ComputePipeline, error) {
	build := func(label, src string) (*wgpu.ComputePipeline, error) {
		mod, err := device.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
			Label: label,
			WGSL:  src,
		})
		if err != nil {
			return nil, fmt.Errorf("shader %s: %w", label, err)
		}
		defer mod.Release()

		return device.CreateComputePipeline(&wgpu.ComputePipelineDescriptor{
			Label:      label,
			Layout:     pl,
			Module:     mod,
			EntryPoint: "main",
		})
	}

	switch r.stats.Mode {
	case ModeALU:
		p, err := build("alu", shaderALU)
		if err != nil {
			return nil, err
		}
		return []*wgpu.ComputePipeline{p}, nil
	case ModeMem:
		p, err := build("mem", shaderMem)
		if err != nil {
			return nil, err
		}
		return []*wgpu.ComputePipeline{p}, nil
	default: // mixed
		pa, err := build("alu", shaderALU)
		if err != nil {
			return nil, err
		}
		pm, err := build("mem", shaderMem)
		if err != nil {
			pa.Release()
			return nil, err
		}
		return []*wgpu.ComputePipeline{pa, pm}, nil
	}
}
