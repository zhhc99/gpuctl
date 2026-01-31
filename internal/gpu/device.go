package gpu

type Device interface {
	Index() int
	Name() string
	UUID() string

	Utilization() (gpu, mem int, err error)
	Clocks() (gpu, mem int, err error)
	Memory() (total, free, used int, err error) // byte
	Power() (watt int, err error)
	Temperature() (celsius int, err error)
	FanSpeed() (percent, rpm int, err error)
	PowerLimit() (watt int, err error)
	ClockOffsetGPU() (mhz int, err error)
	ClockOffsetMem() (mhz int, err error)
	ClockLimitGPU() (mhz int, err error)

	PowerLimitRange() (min, max int, err error)
	PowerLimitDefault() (watt int, err error)
	ClockOffsetGPURange() (min, max int, err error)
	ClockOffsetMemRange() (min, max int, err error)
	ClockLimitGPURange() (min, max int, err error)

	SetPowerLimit(watts int) error
	SetClockOffsetGPU(mhz int) error
	SetClockOffsetMem(mhz int) error
	SetClockLimitGPU(mhz int) error
	ResetPowerLimit() error
	ResetClockOffsetGPU() error
	ResetClockOffsetMem() error
	ResetClockLimitGPU() error

	IsPowerLimitSetterSupported() bool
}
