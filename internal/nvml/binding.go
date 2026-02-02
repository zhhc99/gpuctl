package nvml

import (
	"bytes"
	"fmt"
	"gpuctl/internal/libloader"
	"unsafe"
)

var (
	VERSION_CLOCK_OFFSET = nvmlStructVersion(unsafe.Sizeof(ClockOffset{}), 1)
	VERSION_FAN_SPEED    = nvmlStructVersion(unsafe.Sizeof(FanSpeedInfo{}), 1)
	VERSION_TEMPERATURE  = nvmlStructVersion(unsafe.Sizeof(Temperature{}), 1)
)

type NvmlLib struct {
	// systems
	Init_v2                    func() Return
	SystemGetDriverVersion     func(buffer *byte, length uint32) Return // NVML_SYSTEM_DRIVER_VERSION_BUFFER_SIZE=80
	SystemGetNVMLVersion       func(buffer *byte, length uint32) Return // NVML_SYSTEM_NVML_VERSION_BUFFER_SIZE=80
	SystemGetCudaDriverVersion func(version *int32) Return
	Shutdown                   func() Return
	ErrorString                func(result Return) uintptr

	// find devices
	DeviceGetCount_v2         func(count *uint32) Return
	DeviceGetHandleByIndex_v2 func(index uint32, device *DeviceHandle) Return
	DeviceGetIndex            func(device DeviceHandle, index *uint32) Return
	DeviceGetUUID             func(device DeviceHandle, buffer *byte, length uint32) Return // NVML_DEVICE_UUID_BUFFER_SIZE=80
	DeviceGetName             func(device DeviceHandle, buffer *byte, length uint32) Return // NVML_DEVICE_NAME_BUFFER_SIZE=64

	// monitor
	DeviceGetUtilizationRates          func(device DeviceHandle, util *Utilization) Return
	DeviceGetMemoryInfo                func(device DeviceHandle, memory *Memory) Return
	DeviceGetClockInfo                 func(device DeviceHandle, clockType ClockType, clock *uint32) Return
	DeviceGetPowerUsage                func(device DeviceHandle, power *uint32) Return
	DeviceGetEnforcedPowerLimit        func(device DeviceHandle, limit *uint32) Return
	DeviceGetTemperature               func(device DeviceHandle, sensor TemperatureSensors, temp *uint32) Return
	DeviceGetTemperatureV              func(device DeviceHandle, info *Temperature) Return
	DeviceGetFanSpeed                  func(device DeviceHandle, speed *uint32) Return
	DeviceGetFanSpeedRPM               func(device DeviceHandle, info *FanSpeedInfo) Return
	DeviceGetSamples                   func(device DeviceHandle, samplingType SamplingType, lastSeen uint64, valType *ValueType, count *uint32, samples *Sample) Return
	DeviceGetCurrentClocksEventReasons func(device DeviceHandle, reasons *uint64) Return

	// power limits
	DeviceGetPowerManagementLimitConstraints func(device DeviceHandle, min *uint32, max *uint32) Return
	DeviceGetPowerManagementDefaultLimit     func(device DeviceHandle, limit *uint32) Return
	DeviceGetPowerManagementLimit            func(device DeviceHandle, limit *uint32) Return
	DeviceSetPowerManagementLimit            func(device DeviceHandle, limit uint32) Return

	// clock offsets
	DeviceGetClockOffsets         func(device DeviceHandle, info *ClockOffset) Return
	DeviceGetGpcClkVfOffset       func(device DeviceHandle, offset *int32) Return          // GetClockOffsetsLegacy
	DeviceGetGpcClkMinMaxVfOffset func(device DeviceHandle, min *int32, max *int32) Return // GetClockOffsetsLegacy
	DeviceGetMemClkVfOffset       func(device DeviceHandle, offset *int32) Return
	DeviceGetMemClkMinMaxVfOffset func(device DeviceHandle, min *int32, max *int32) Return
	DeviceGetMaxClockInfo         func(device DeviceHandle, clockType ClockType, clock *uint32) Return
	DeviceSetClockOffsets         func(device DeviceHandle, info *ClockOffset) Return
	DeviceSetGpcClkVfOffset       func(device DeviceHandle, offset int32) Return // SetClockOffsetsLegacy
	DeviceSetMemClkVfOffset       func(device DeviceHandle, offset int32) Return

	// locked clocks
	DeviceGetSupportedMemoryClocks   func(device DeviceHandle, count *uint32, clocksMHz *uint32) Return
	DeviceGetSupportedGraphicsClocks func(device DeviceHandle, memoryClockMHz uint32, count *uint32, clocksMHz *uint32) Return
	DeviceSetGpuLockedClocks         func(device DeviceHandle, min uint32, max uint32) Return
	DeviceResetGpuLockedClocks       func(device DeviceHandle) Return
}

func NewNvmlLib() (*NvmlLib, error) {
	lib, err := libloader.Load(libloader.NVML)
	if err != nil {
		return nil, err
	}

	nvml := &NvmlLib{}

	libloader.Bind(lib, &nvml.Init_v2, "nvmlInit_v2")
	libloader.Bind(lib, &nvml.SystemGetDriverVersion, "nvmlSystemGetDriverVersion")
	libloader.Bind(lib, &nvml.SystemGetNVMLVersion, "nvmlSystemGetNVMLVersion")
	libloader.Bind(lib, &nvml.SystemGetCudaDriverVersion, "nvmlSystemGetCudaDriverVersion")
	libloader.Bind(lib, &nvml.Shutdown, "nvmlShutdown")
	libloader.Bind(lib, &nvml.ErrorString, "nvmlErrorString")

	libloader.Bind(lib, &nvml.DeviceGetCount_v2, "nvmlDeviceGetCount_v2")
	libloader.Bind(lib, &nvml.DeviceGetHandleByIndex_v2, "nvmlDeviceGetHandleByIndex_v2")
	libloader.Bind(lib, &nvml.DeviceGetIndex, "nvmlDeviceGetIndex")
	libloader.Bind(lib, &nvml.DeviceGetUUID, "nvmlDeviceGetUUID")
	libloader.Bind(lib, &nvml.DeviceGetName, "nvmlDeviceGetName")

	libloader.Bind(lib, &nvml.DeviceGetUtilizationRates, "nvmlDeviceGetUtilizationRates")
	libloader.Bind(lib, &nvml.DeviceGetMemoryInfo, "nvmlDeviceGetMemoryInfo")
	libloader.Bind(lib, &nvml.DeviceGetClockInfo, "nvmlDeviceGetClockInfo")
	libloader.Bind(lib, &nvml.DeviceGetPowerUsage, "nvmlDeviceGetPowerUsage")
	libloader.Bind(lib, &nvml.DeviceGetEnforcedPowerLimit, "nvmlDeviceGetEnforcedPowerLimit")
	libloader.Bind(lib, &nvml.DeviceGetTemperature, "nvmlDeviceGetTemperature")
	libloader.Bind(lib, &nvml.DeviceGetTemperatureV, "nvmlDeviceGetTemperatureV")
	libloader.Bind(lib, &nvml.DeviceGetFanSpeed, "nvmlDeviceGetFanSpeed")
	libloader.Bind(lib, &nvml.DeviceGetFanSpeedRPM, "nvmlDeviceGetFanSpeedRPM")
	libloader.Bind(lib, &nvml.DeviceGetSamples, "nvmlDeviceGetSamples")
	libloader.Bind(lib, &nvml.DeviceGetCurrentClocksEventReasons, "nvmlDeviceGetCurrentClocksEventReasons")

	libloader.Bind(lib, &nvml.DeviceGetPowerManagementLimitConstraints, "nvmlDeviceGetPowerManagementLimitConstraints")
	libloader.Bind(lib, &nvml.DeviceGetPowerManagementDefaultLimit, "nvmlDeviceGetPowerManagementDefaultLimit")
	libloader.Bind(lib, &nvml.DeviceGetPowerManagementLimit, "nvmlDeviceGetPowerManagementLimit")
	libloader.Bind(lib, &nvml.DeviceSetPowerManagementLimit, "nvmlDeviceSetPowerManagementLimit")

	libloader.Bind(lib, &nvml.DeviceGetClockOffsets, "nvmlDeviceGetClockOffsets")
	libloader.Bind(lib, &nvml.DeviceGetGpcClkVfOffset, "nvmlDeviceGetGpcClkVfOffset")
	libloader.Bind(lib, &nvml.DeviceGetGpcClkMinMaxVfOffset, "nvmlDeviceGetGpcClkMinMaxVfOffset")
	libloader.Bind(lib, &nvml.DeviceGetMemClkVfOffset, "nvmlDeviceGetMemClkVfOffset")
	libloader.Bind(lib, &nvml.DeviceGetMemClkMinMaxVfOffset, "nvmlDeviceGetMemClkMinMaxVfOffset")
	libloader.Bind(lib, &nvml.DeviceGetMaxClockInfo, "nvmlDeviceGetMaxClockInfo")
	libloader.Bind(lib, &nvml.DeviceSetClockOffsets, "nvmlDeviceSetClockOffsets")
	libloader.Bind(lib, &nvml.DeviceSetGpcClkVfOffset, "nvmlDeviceSetGpcClkVfOffset")
	libloader.Bind(lib, &nvml.DeviceSetMemClkVfOffset, "nvmlDeviceSetMemClkVfOffset")

	libloader.Bind(lib, &nvml.DeviceGetSupportedMemoryClocks, "nvmlDeviceGetSupportedMemoryClocks")
	libloader.Bind(lib, &nvml.DeviceGetSupportedGraphicsClocks, "nvmlDeviceGetSupportedGraphicsClocks")
	libloader.Bind(lib, &nvml.DeviceSetGpuLockedClocks, "nvmlDeviceSetGpuLockedClocks")
	libloader.Bind(lib, &nvml.DeviceResetGpuLockedClocks, "nvmlDeviceResetGpuLockedClocks")

	return nvml, nil
}

func (s *NvmlLib) StringFromReturn(r Return) string {
	if s.ErrorString == nil {
		return fmt.Sprintf("NVML Error %d", r)
	}
	ptr := s.ErrorString(r)
	if ptr == 0 {
		return "Unknown NVML Error"
	}

	p := (*byte)(unsafe.Pointer(ptr))

	const maxLen = 1024

	str := unsafe.Slice(p, maxLen)
	n := bytes.IndexByte(str, 0)
	if n == -1 {
		n = maxLen
	}

	return string(str[:n])
}

// nvmlStructVersion implements NVML_STRUCT_VERSION
func nvmlStructVersion(size uintptr, ver uint32) uint32 {
	return uint32(size) | (ver << 24)
}
