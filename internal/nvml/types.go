package nvml

type Return int32
type DeviceHandle uintptr
type Value struct{ Data [8]byte } // nvmlValue_t, a union
type Sample struct {
	TimeStamp   uint64
	SampleValue Value
}
type ClockOffset struct {
	Version                                              uint32
	Type                                                 ClockType
	Pstate                                               Pstates
	ClockOffsetMHz, MinClockOffsetMHz, MaxClockOffsetMHz int32
}
type FanSpeedInfo struct{ Version, Fan, Speed uint32 }
type Temperature struct {
	Version     uint32
	SensorType  TemperatureSensors
	Temperature uint32
}
type Utilization struct{ Gpu, Memory uint32 }
type Memory struct{ Total, Free, Used uint64 }

type ClockType int32          // nvmlClockType_t
type SamplingType int32       // nvmlSamplingType_t
type TemperatureSensors int32 // nvmlTemperatureSensors_t
type ValueType int32          // nvmlValueType_t
type Pstates int32            // nvmlPstates_t

const (
	CLOCK_GRAPHICS ClockType = 0
	CLOCK_SM       ClockType = 1
	CLOCK_MEM      ClockType = 2
	CLOCK_VIDEO    ClockType = 3
)
const (
	TOTAL_POWER_SAMPLES        SamplingType = 0
	GPU_UTILIZATION_SAMPLES    SamplingType = 1
	MEMORY_UTILIZATION_SAMPLES SamplingType = 2
	ENC_UTILIZATION_SAMPLES    SamplingType = 3
	DEC_UTILIZATION_SAMPLES    SamplingType = 4
	PROCESSOR_CLK_SAMPLES      SamplingType = 5
	MEMORY_CLK_SAMPLES         SamplingType = 6
)
const (
	TEMPERATURE_GPU TemperatureSensors = 0
)
const (
	VALUE_TYPE_DOUBLE             ValueType = 0
	VALUE_TYPE_UNSIGNED_INT       ValueType = 1
	VALUE_TYPE_UNSIGNED_LONG      ValueType = 2
	VALUE_TYPE_UNSIGNED_LONG_LONG ValueType = 3
	VALUE_TYPE_SIGNED_LONG_LONG   ValueType = 4
	VALUE_TYPE_SIGNED_INT         ValueType = 5
	VALUE_TYPE_UNSIGNED_SHORT     ValueType = 6
)
const (
	PSTATE_0 Pstates = iota
	PSTATE_1
	PSTATE_2
	PSTATE_3
	PSTATE_4
	PSTATE_5
	PSTATE_6
	PSTATE_7
	PSTATE_8
	PSTATE_9
	PSTATE_10
	PSTATE_11
	PSTATE_12
	PSTATE_13
	PSTATE_14
	PSTATE_15
)

const (
	SYSTEM_DRIVER_VERSION_BUFFER_SIZE = 80
	SYSTEM_NVML_VERSION_BUFFER_SIZE   = 80
	DEVICE_UUID_BUFFER_SIZE           = 80
	DEVICE_NAME_BUFFER_SIZE           = 64
)
const (
	SUCCESS                         Return = 0
	ERROR_UNINITIALIZED             Return = 1
	ERROR_INVALID_ARGUMENT          Return = 2
	ERROR_NOT_SUPPORTED             Return = 3
	ERROR_NO_PERMISSION             Return = 4
	ERROR_ALREADY_INITIALIZED       Return = 5
	ERROR_NOT_FOUND                 Return = 6
	ERROR_INSUFFICIENT_SIZE         Return = 7
	ERROR_INSUFFICIENT_POWER        Return = 8
	ERROR_DRIVER_NOT_LOADED         Return = 9
	ERROR_TIMEOUT                   Return = 10
	ERROR_IRQ_ISSUE                 Return = 11
	ERROR_LIBRARY_NOT_FOUND         Return = 12
	ERROR_FUNCTION_NOT_FOUND        Return = 13
	ERROR_CORRUPTED_INFOROM         Return = 14
	ERROR_GPU_IS_LOST               Return = 15
	ERROR_RESET_REQUIRED            Return = 16
	ERROR_OPERATING_SYSTEM          Return = 17
	ERROR_LIB_RM_VERSION_MISMATCH   Return = 18
	ERROR_IN_USE                    Return = 19
	ERROR_MEMORY                    Return = 20
	ERROR_NO_DATA                   Return = 21
	ERROR_VGPU_ECC_NOT_SUPPORTED    Return = 22
	ERROR_INSUFFICIENT_RESOURCES    Return = 23
	ERROR_FREQ_NOT_SUPPORTED        Return = 24
	ERROR_ARGUMENT_VERSION_MISMATCH Return = 25
	ERROR_DEPRECATED                Return = 26
	ERROR_NOT_READY                 Return = 27
	ERROR_GPU_NOT_FOUND             Return = 28
	ERROR_INVALID_STATE             Return = 29
	ERROR_RESET_TYPE_NOT_SUPPORTED  Return = 30
	ERROR_UNKNOWN                   Return = 999
)
