package config

import "fmt"

type FanCurvePoint struct {
	Temp int `yaml:"temp"`
	Fan  int `yaml:"fan"`
}

type Profile struct {
	PowerLimit     *int            `yaml:"power_limit,omitempty"`
	ClockOffsetGPU *int            `yaml:"clock_offset_gpu,omitempty"`
	ClockOffsetMem *int            `yaml:"clock_offset_mem,omitempty"`
	ClockLimitGPU  *int            `yaml:"clock_limit_gpu,omitempty"`
	FanControl     *bool           `yaml:"fan_control,omitempty"`
	FanCurve       []FanCurvePoint `yaml:"fan_curve,omitempty"`
}

type Config struct {
	Settings map[string]Profile `yaml:"settings"` // key: "gpu name | pci bus id"
}

func ValidateFanCurve(curve []FanCurvePoint) error {
	if len(curve) < 2 {
		return fmt.Errorf("fan curve must have at least 2 points")
	}
	for i, p := range curve {
		if p.Temp < 0 || p.Temp > 100 {
			return fmt.Errorf("point %d: temp %d out of range [0, 100]", i, p.Temp)
		}
		if p.Fan < 0 || p.Fan > 100 {
			return fmt.Errorf("point %d: fan %d out of range [0, 100]", i, p.Fan)
		}
		if i > 0 {
			if p.Temp <= curve[i-1].Temp {
				return fmt.Errorf("point %d: temp not strictly increasing", i)
			}
			if p.Fan < curve[i-1].Fan {
				return fmt.Errorf("point %d: fan must not decrease", i)
			}
		}
	}
	return nil
}

func InterpolateFan(curve []FanCurvePoint, temp int) int {
	if temp <= curve[0].Temp {
		return curve[0].Fan
	}
	if temp >= curve[len(curve)-1].Temp {
		return curve[len(curve)-1].Fan
	}
	for i := 1; i < len(curve); i++ {
		if temp <= curve[i].Temp {
			t0, t1 := curve[i-1].Temp, curve[i].Temp
			f0, f1 := curve[i-1].Fan, curve[i].Fan
			return f0 + (f1-f0)*(temp-t0)/(t1-t0)
		}
	}
	return curve[len(curve)-1].Fan
}
