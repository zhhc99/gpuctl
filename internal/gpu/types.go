package gpu

const Unavailable int = -0x7FFFFFFF

type Snapshot struct {
	Index int
	Name  string
	UUID  string

	UtilizationGPU int
	UtilizationMem int
	ClockGpu       int // MHz
	ClockMem       int // MHz
	MemTotal       int // Byte
	MemUsed        int // Byte
	Power          int // Watt
	Temperature    int // Celsius
	FanPct         int
	FanRPM         int

	PowerLimit     int // Watt
	ClockOffsetGPU int // MHz
	ClockOffsetMem int // MHz
	ClockLimitGPU  int // MHz

	PowerLimitMin         int // Watt
	PowerLimitMax         int // Watt
	PowerLimitDefault     int // Watt
	ClockOffsetGPUMin     int // MHz
	ClockOffsetGPUMax     int // MHz
	ClockOffsetGPUDefault int // MHz
	ClockOffsetMemMin     int // MHz
	ClockOffsetMemMax     int // MHz
	ClockOffsetMemDefault int // MHz
	ClockLimitGPUMin      int // MHz
	ClockLimitGPUMax      int // MHz
	ClockLimitGPUDefault  int // MHz
}

func (d *Snapshot) Capture(dev Device) {
	d.Index = dev.Index()
	if d.Name == "" {
		d.Name = dev.Name()
	}
	if d.UUID == "" {
		d.UUID = dev.UUID()
	}
	d.UtilizationGPU, d.UtilizationMem, _ = dev.Utilization()
	d.ClockGpu, d.ClockMem, _ = dev.Clocks()
	d.MemTotal, _, d.MemUsed, _ = dev.Memory()
	d.Power, _ = dev.Power()
	d.Temperature, _ = dev.Temperature()
	d.FanPct, d.FanRPM, _ = dev.FanSpeed()
	d.PowerLimit, _ = dev.PowerLimit()
	d.ClockOffsetGPU, _ = dev.ClockOffsetGPU()
	d.ClockOffsetMem, _ = dev.ClockOffsetMem()
	d.ClockLimitGPU, _ = dev.ClockLimitGPU()

	d.PowerLimitMin, d.PowerLimitMax, _ = dev.PowerLimitRange()
	d.PowerLimitDefault, _ = dev.PowerLimitDefault()
	d.ClockOffsetGPUMin, d.ClockOffsetGPUMax, _ = dev.ClockOffsetGPURange()
	d.ClockOffsetMemMin, d.ClockOffsetMemMax, _ = dev.ClockOffsetMemRange()
	d.ClockLimitGPUMin, d.ClockLimitGPUMax, _ = dev.ClockLimitGPURange()
	d.ClockOffsetGPUDefault, d.ClockOffsetMemDefault, d.ClockLimitGPUDefault = 0, 0, d.ClockLimitGPUMax
}
