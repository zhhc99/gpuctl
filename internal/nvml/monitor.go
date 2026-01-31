package nvml

import (
	"fmt"
	"gpuctl/internal/gpu"
	"unsafe"
)

func (g *Device) Utilization() (int, int, error) {
	var util Utilization
	if ret := g.lib.DeviceGetUtilizationRates(g.handle, &util); ret != SUCCESS {
		return gpu.Unavailable, gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return int(util.Gpu), int(util.Memory), nil
}

func (g *Device) Clocks() (int, int, error) {
	var gclk, mclk uint32
	if ret := g.lib.DeviceGetClockInfo(g.handle, CLOCK_GRAPHICS, &gclk); ret != SUCCESS {
		return gpu.Unavailable, gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	if ret := g.lib.DeviceGetClockInfo(g.handle, CLOCK_MEM, &mclk); ret != SUCCESS {
		return int(gclk), gpu.Unavailable, nil
	}
	return int(gclk), int(mclk), nil
}

func (g *Device) Memory() (int, int, int, error) {
	var mem Memory
	if ret := g.lib.DeviceGetMemoryInfo(g.handle, &mem); ret != SUCCESS {
		return gpu.Unavailable, gpu.Unavailable, gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return int(mem.Total), int(mem.Free), int(mem.Used), nil
}

func (g *Device) Power() (int, error) {
	var mw uint32
	if ret := g.lib.DeviceGetPowerUsage(g.handle, &mw); ret == SUCCESS {
		return int(mw / 1000), nil
	}
	return g.getPowerViaSample()
}

func (g *Device) Temperature() (int, error) {
	// fallback
	if g.lib.DeviceGetTemperatureV == nil {
		var temp uint32
		if ret := g.lib.DeviceGetTemperature(g.handle, TEMPERATURE_GPU, &temp); ret != SUCCESS {
			return gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
		}
		return int(temp), nil
	}

	var temp Temperature
	temp.Version = VERSION_TEMPERATURE
	temp.SensorType = TEMPERATURE_GPU
	if ret := g.lib.DeviceGetTemperatureV(g.handle, &temp); ret != SUCCESS {
		return gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return int(temp.Temperature), nil

}

func (g *Device) FanSpeed() (int, int, error) {
	var percent uint32
	if ret := g.lib.DeviceGetFanSpeed(g.handle, &percent); ret != SUCCESS {
		return gpu.Unavailable, gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
	}

	// error only if fan speed percent is not available
	if g.lib.DeviceGetFanSpeedRPM != nil {
		var fan FanSpeedInfo
		fan.Version = VERSION_FAN_SPEED
		if g.lib.DeviceGetFanSpeedRPM(g.handle, &fan) == SUCCESS {
			return int(percent), int(fan.Speed), nil
		}
	}
	return int(percent), gpu.Unavailable, nil
}

func (g *Device) PowerLimit() (int, error) {
	var mw uint32
	if ret := g.lib.DeviceGetEnforcedPowerLimit(g.handle, &mw); ret != SUCCESS {
		return gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return int(mw / 1000), nil
}

func (g *Device) PowerLimitDefault() (int, error) {
	var mw uint32
	if ret := g.lib.DeviceGetPowerManagementDefaultLimit(g.handle, &mw); ret != SUCCESS {
		return gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return int(mw / 1000), nil
}

func (g *Device) ClockOffsetGPU() (int, error) {
	// fallback
	if g.lib.DeviceGetClockOffsets == nil {
		var co int32
		if ret := g.lib.DeviceGetGpcClkVfOffset(g.handle, &co); ret != SUCCESS {
			return gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
		}
		return int(co), nil
	}

	var co ClockOffset
	co.Version = VERSION_CLOCK_OFFSET
	co.Type = CLOCK_GRAPHICS
	co.Pstate = PSTATE_0
	if ret := g.lib.DeviceGetClockOffsets(g.handle, &co); ret != SUCCESS {
		return gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return int(co.ClockOffsetMHz), nil
}

func (g *Device) ClockOffsetMem() (int, error) {
	// fallback
	if g.lib.DeviceGetClockOffsets == nil {
		var co int32
		if ret := g.lib.DeviceGetMemClkVfOffset(g.handle, &co); ret != SUCCESS {
			return gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
		}
		return int(co), nil
	}

	var co ClockOffset
	co.Version = VERSION_CLOCK_OFFSET
	co.Type = CLOCK_MEM
	co.Pstate = PSTATE_0
	if ret := g.lib.DeviceGetClockOffsets(g.handle, &co); ret != SUCCESS {
		return gpu.Unavailable, fmt.Errorf(g.lib.StringFromReturn(ret))
	}
	return int(co.ClockOffsetMHz), nil
}

func (g *Device) ClockLimitGPU() (int, error) {
	return gpu.Unavailable, fmt.Errorf("not supported by nvidia yet")
}

func (g *Device) getPowerViaSample() (int, error) {
	var (
		sample      Sample
		sampleType  ValueType
		sampleCount uint32 = 1
	)

	ret := g.lib.DeviceGetSamples(g.handle, TOTAL_POWER_SAMPLES, 0, &sampleType, &sampleCount, &sample)
	if ret != SUCCESS || sampleCount == 0 {
		return gpu.Unavailable, fmt.Errorf("sampling failed: %s", g.lib.StringFromReturn(ret))
	}

	var mw float64
	switch sampleType {
	case VALUE_TYPE_DOUBLE:
		mw = *(*float64)(unsafe.Pointer(&sample.SampleValue.Data))
	case VALUE_TYPE_UNSIGNED_INT:
		mw = float64(*(*uint32)(unsafe.Pointer(&sample.SampleValue.Data)))
	case VALUE_TYPE_UNSIGNED_LONG, VALUE_TYPE_UNSIGNED_LONG_LONG:
		mw = float64(*(*uint64)(unsafe.Pointer(&sample.SampleValue.Data)))
	default:
		mw = float64(sample.SampleValue.asUInt32())
	}

	return int(mw / 1000), nil
}

func (v Value) asUInt32() uint32 {
	return *(*uint32)(unsafe.Pointer(&v.Data))
}
