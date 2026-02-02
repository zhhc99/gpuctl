package nvml

import (
	"errors"
	"fmt"
	"gpuctl/internal/gpu"
	"strings"
)

var _ gpu.Device = (*Device)(nil)

type Device struct {
	handle DeviceHandle
	lib    *NvmlLib
	index  int
	name   string
	uuid   string
}

func newDevice(handle DeviceHandle, lib *NvmlLib) *Device {
	g := &Device{handle: handle, lib: lib}

	g.fetchIndex()
	g.fetchName()
	g.fetchUUID()

	return g
}

func (g *Device) Index() int   { return g.index }
func (g *Device) Name() string { return g.name }
func (g *Device) UUID() string { return g.uuid }

func (g *Device) fetchIndex() {
	var index uint32
	if ret := g.lib.DeviceGetIndex(g.handle, &index); ret != SUCCESS {
		panic("Fatal: failed to fetch gpu index. check if your gpu is still connected.")
	}
	g.index = int(index)
}

func (g *Device) fetchName() {
	var buf [DEVICE_NAME_BUFFER_SIZE]byte
	if ret := g.lib.DeviceGetName(g.handle, &buf[0], DEVICE_NAME_BUFFER_SIZE); ret != SUCCESS {
		g.name = "Unknown Nvidia GPU"
	}
	g.name = strings.TrimRight(string(buf[:]), "\x00")
}

func (g *Device) fetchUUID() {
	var buf [DEVICE_UUID_BUFFER_SIZE]byte
	if ret := g.lib.DeviceGetUUID(g.handle, &buf[0], DEVICE_UUID_BUFFER_SIZE); ret != SUCCESS {
		g.uuid = "Unknown UUID"
	}
	g.uuid = strings.TrimRight(string(buf[:]), "\x00")
}

func (g *Device) getSupportedMemClocks() ([]int, error) {
	var count uint32
	ret := g.lib.DeviceGetSupportedMemoryClocks(g.handle, &count, nil)
	if ret != SUCCESS && ret != ERROR_INSUFFICIENT_SIZE {
		return nil, fmt.Errorf("failed to get supported mem clock count: %s", g.lib.StringFromReturn(ret))
	}

	if count == 0 {
		return []int{}, nil
	}

	clocks := make([]uint32, count)
	ret = g.lib.DeviceGetSupportedMemoryClocks(g.handle, &count, &clocks[0])
	if ret != SUCCESS {
		return nil, fmt.Errorf("failed to fetch supported mem clocks: %s", g.lib.StringFromReturn(ret))
	}

	res := make([]int, count)
	for i, v := range clocks {
		res[i] = int(v)
	}
	return res, nil
}

func (g *Device) getSupportedGpuClocks(memClockMHz int) ([]int, error) {
	var count uint32
	ret := g.lib.DeviceGetSupportedGraphicsClocks(g.handle, uint32(memClockMHz), &count, nil)
	if ret != SUCCESS && ret != ERROR_INSUFFICIENT_SIZE {
		return nil, fmt.Errorf("failed to get gpu clock count: %s", g.lib.StringFromReturn(ret))
	}

	if count == 0 {
		return []int{}, nil
	}

	clocks := make([]uint32, count)
	ret = g.lib.DeviceGetSupportedGraphicsClocks(g.handle, uint32(memClockMHz), &count, &clocks[0])
	if ret != SUCCESS {
		return nil, fmt.Errorf("failed to fetch gpu clocks: %s", g.lib.StringFromReturn(ret))
	}

	res := make([]int, count)
	for i, v := range clocks {
		res[i] = int(v)
	}
	return res, nil
}

func (g *Device) getClockLimitGPURangeV1() (int, int, error) {
	// 将 min 视作 0 会更简洁, 但这不对

	var max uint32
	if ret := g.lib.DeviceGetMaxClockInfo(g.handle, CLOCK_GRAPHICS, &max); ret != SUCCESS {
		return gpu.Unavailable, gpu.Unavailable, fmt.Errorf("failed to get max clock: %s", g.lib.StringFromReturn(ret))
	}
	co, err := g.ClockOffsetGPU()
	if err != nil {
		return gpu.Unavailable, gpu.Unavailable, fmt.Errorf("failed to get current co: %w", err)
	}
	return 0, int(max) - co, nil
}

func (g *Device) getClockLimitGPURangeV2() (int, int, error) {
	// see also: nvidia-smi -q -d SUPPORTED_CLOCKS

	memClocks, err := g.getSupportedMemClocks()
	if err != nil {
		return gpu.Unavailable, gpu.Unavailable, fmt.Errorf("failed to get supported mem clocks: %w", err)
	}
	if len(memClocks) == 0 {
		return gpu.Unavailable, gpu.Unavailable, errors.New("no supported mem clocks")
	}

	minMem := memClocks[0]
	maxMem := memClocks[0]
	for _, m := range memClocks {
		if m < minMem {
			minMem = m
		}
		if m > maxMem {
			maxMem = m
		}
	}

	minGpuClocks, err := g.getSupportedGpuClocks(minMem)
	if err != nil {
		return gpu.Unavailable, gpu.Unavailable, fmt.Errorf("failed to get gpu clocks for min mem %d: %v", minMem, err)
	}
	if len(minGpuClocks) == 0 {
		return gpu.Unavailable, gpu.Unavailable, fmt.Errorf("no gpu clocks found for min mem %d", minMem)
	}
	maxGpuClocks, err := g.getSupportedGpuClocks(maxMem)
	if err != nil {
		return gpu.Unavailable, gpu.Unavailable, fmt.Errorf("failed to get gpu clocks for max mem %d: %v", maxMem, err)
	}
	if len(maxGpuClocks) == 0 {
		return gpu.Unavailable, gpu.Unavailable, fmt.Errorf("no gpu clocks found for max mem %d", maxMem)
	}

	minGpu := minGpuClocks[0]
	for _, g := range minGpuClocks {
		if g < minGpu {
			minGpu = g
		}
	}
	maxGpu := maxGpuClocks[0]
	for _, g := range maxGpuClocks {
		if g > maxGpu {
			maxGpu = g
		}
	}

	return minGpu, maxGpu, nil
}
