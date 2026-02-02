package config

type Profile struct {
	PowerLimit     *int `yaml:"power_limit"`
	ClockOffsetGPU *int `yaml:"clock_offset_gpu"`
	ClockOffsetMem *int `yaml:"clock_offset_mem"`
	ClockLimitGPU  *int `yaml:"clock_limit_gpu"`
}

type Config struct {
	Settings map[string]Profile `yaml:"settings"` // key: uuid
}
