package nvml

import (
	"errors"
	"fmt"
	"gpuctl/internal/gpu"
)

func (g *Device) PowerLimitRange() (int, int, error) {
	var min, max uint32
	if ret := g.lib.DeviceGetPowerManagementLimitConstraints(g.handle, &min, &max); ret != SUCCESS {
		return gpu.Unavailable, gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return int(min / 1000), int(max / 1000), nil
}

func (g *Device) ClockOffsetGPURange() (int, int, error) {
	// fallback
	if g.lib.DeviceGetClockOffsets == nil {
		var min, max int32
		if ret := g.lib.DeviceGetGpcClkMinMaxVfOffset(g.handle, &min, &max); ret != SUCCESS {
			return gpu.Unavailable, gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
		}
		return int(min), int(max), nil
	}

	var co ClockOffset
	co.Version = VERSION_CLOCK_OFFSET
	co.Type = CLOCK_GRAPHICS
	co.Pstate = PSTATE_0
	if ret := g.lib.DeviceGetClockOffsets(g.handle, &co); ret != SUCCESS {
		return gpu.Unavailable, gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return int(co.MinClockOffsetMHz), int(co.MaxClockOffsetMHz), nil
}

func (g *Device) ClockOffsetMemRange() (int, int, error) {
	// fallback
	if g.lib.DeviceGetClockOffsets == nil {
		var min, max int32
		if ret := g.lib.DeviceGetMemClkMinMaxVfOffset(g.handle, &min, &max); ret != SUCCESS {
			return gpu.Unavailable, gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
		}
		return int(min), int(max), nil
	}

	var co ClockOffset
	co.Version = VERSION_CLOCK_OFFSET
	co.Type = CLOCK_MEM
	co.Pstate = PSTATE_0
	if ret := g.lib.DeviceGetClockOffsets(g.handle, &co); ret != SUCCESS {
		return gpu.Unavailable, gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return int(co.MinClockOffsetMHz), int(co.MaxClockOffsetMHz), nil
}

func (g *Device) ClockLimitGPURange() (int, int, error) {
	return g.getClockLimitGPURangeV2()
}

func (g *Device) IsPowerLimitSetterSupported() bool {
	var limit uint32

	// 部分 laptop GPUs 自行控制功耗墙: getter 总是失败, setter 成功但无效
	return g.lib.DeviceGetPowerManagementLimit != nil &&
		g.lib.DeviceSetPowerManagementLimit != nil &&
		g.lib.DeviceGetPowerManagementLimit(g.handle, &limit) == SUCCESS
}

func (g *Device) SetPowerLimit(watt int) error {
	if !g.IsPowerLimitSetterSupported() {
		return errors.New("controlled by vbios/hardware")
	}
	if ret := g.lib.DeviceSetPowerManagementLimit(g.handle, uint32(watt*1000)); ret != SUCCESS {
		return fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return nil
}

func (g *Device) SetClockOffsetGPU(mhz int) error {
	// fallback
	if g.lib.DeviceSetClockOffsets == nil {
		if ret := g.lib.DeviceSetGpcClkVfOffset(g.handle, int32(mhz)); ret != SUCCESS {
			return fmt.Errorf(g.lib.StringFromReturn(ret))
		}
		return nil
	}

	var co ClockOffset
	co.Version = VERSION_CLOCK_OFFSET
	co.Type = CLOCK_GRAPHICS
	co.Pstate = PSTATE_0
	co.ClockOffsetMHz = int32(mhz)
	if ret := g.lib.DeviceSetClockOffsets(g.handle, &co); ret != SUCCESS {
		return fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return nil
}

func (g *Device) SetClockOffsetMem(mhz int) error {
	// fallback
	if g.lib.DeviceSetClockOffsets == nil {
		if ret := g.lib.DeviceSetMemClkVfOffset(g.handle, int32(mhz)); ret != SUCCESS {
			return fmt.Errorf(g.lib.StringFromReturn(ret))
		}
		return nil
	}

	var co ClockOffset
	co.Version = VERSION_CLOCK_OFFSET
	co.Type = CLOCK_MEM
	co.Pstate = PSTATE_0
	co.ClockOffsetMHz = int32(mhz)
	if ret := g.lib.DeviceSetClockOffsets(g.handle, &co); ret != SUCCESS {
		return fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return nil
}

func (g *Device) SetClockLimitGPU(mhz int) error {
	if ret := g.lib.DeviceSetGpuLockedClocks(g.handle, 0, uint32(mhz)); ret != SUCCESS {
		return fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return nil
}

func (g *Device) ResetPowerLimit() error {
	if !g.IsPowerLimitSetterSupported() {
		return errors.New("controlled by vbios/hardware")
	}
	defaultPl, err := g.PowerLimitDefault()
	if err != nil {
		return fmt.Errorf("failed to get default pl: %w", err)
	}
	return g.SetPowerLimit(defaultPl)
}

func (g *Device) ResetClockOffsetGPU() error {
	return g.SetClockOffsetGPU(0)
}

func (g *Device) ResetClockOffsetMem() error {
	return g.SetClockOffsetMem(0)
}

func (g *Device) ResetClockLimitGPU() error {
	if ret := g.lib.DeviceResetGpuLockedClocks(g.handle); ret != SUCCESS {
		return fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return nil
}
