package nvml

import (
	"fmt"
	"gpuctl/internal/gpu"
	"strings"
)

var _ gpu.Backend = (*Backend)(nil)

type Backend struct {
	l *NvmlLib
}

func NewBackend() (*Backend, error) {
	l, err := NewNvmlLib()
	if err != nil {
		return nil, err
	}
	return &Backend{l: l}, nil
}

func (d *Backend) Init() error {
	if ret := d.l.Init_v2(); ret != SUCCESS {
		return fmt.Errorf("nvml init failed: %s", d.l.StringFromReturn(ret))
	}
	return nil
}

func (d *Backend) Shutdown() error {
	if d.l.Shutdown != nil {
		d.l.Shutdown()
	}
	return nil
}

func (d *Backend) Name() string {
	return "NVML"
}

func (d *Backend) Version() string {
	var buf [SYSTEM_NVML_VERSION_BUFFER_SIZE]byte
	if ret := d.l.SystemGetNVMLVersion(&buf[0], SYSTEM_NVML_VERSION_BUFFER_SIZE); ret != SUCCESS {
		return "Unknown"
	}
	return string(buf[:len(strings.TrimRight(string(buf[:]), "\x00"))])
}

func (d *Backend) DriverVersion() string {
	var buf [SYSTEM_DRIVER_VERSION_BUFFER_SIZE]byte
	if ret := d.l.SystemGetDriverVersion(&buf[0], SYSTEM_DRIVER_VERSION_BUFFER_SIZE); ret != SUCCESS {
		return "Unknown"
	}
	return string(buf[:len(strings.TrimRight(string(buf[:]), "\x00"))])
}

func (d *Backend) GPUs() ([]gpu.Device, error) {
	var count uint32
	if ret := d.l.DeviceGetCount_v2(&count); ret != SUCCESS {
		return nil, fmt.Errorf("get device count failed: %s", d.l.StringFromReturn(ret))
	}

	var res []gpu.Device
	for i := uint32(0); i < count; i++ {
		var handle DeviceHandle
		if ret := d.l.DeviceGetHandleByIndex_v2(i, &handle); ret != SUCCESS {
			continue
		}

		g := newDevice(handle, d.l)
		if g.index != int(i) {
			panic("Fatal: gpu index mismatch")
		}
		res = append(res, g)
	}
	return res, nil
}
