package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/zhhc99/gpuctl/internal/config"
	"github.com/zhhc99/gpuctl/internal/gpu"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var daemonCmd = &cobra.Command{
	Use:    "daemon",
	Short:  "Run as a background service",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDaemon()
	},
}

func runFanLoop(quit <-chan struct{}, reload <-chan struct{}, cfg *config.Config) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-quit:
			return
		case <-reload:
			if c, err := loadConfigFromDisk(); err != nil {
				fmt.Fprintf(os.Stderr, "daemon: reload error: %v\n", err)
			} else {
				cfg = c
			}
		case <-ticker.C:
			adjustFans(cfg)
		}
	}
}

func loadConfigFromDisk() (*config.Config, error) {
	data, err := os.ReadFile(config.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", config.ConfigPath, err)
	}
	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func applyConfig(cfg *config.Config) {
	if cfg == nil {
		return
	}
	for i, dev := range Devices {
		profile, ok := cfg.Settings[fmt.Sprintf("%s | %s", dev.Name(), dev.PCIBusID())]
		if !ok {
			continue
		}
		fmt.Printf("Applying config to Device %d (%s)...\n", i, dev.Name())
		applyProfileToDevice(dev, profile)
	}
}

func applyProfileToDevice(dev gpu.Device, profile config.Profile) {
	if profile.PowerLimit != nil {
		if err := dev.SetPowerLimit(*profile.PowerLimit); err != nil {
			fmt.Printf("  [X] PowerLimit (%dW): %v\n", *profile.PowerLimit, err)
		} else {
			fmt.Printf("  [✔] PowerLimit (%dW)\n", *profile.PowerLimit)
		}
	} else {
		fmt.Printf("  [✔] PowerLimit skipped.\n")
	}

	if profile.ClockOffsetGPU != nil {
		if err := dev.SetClockOffsetGPU(*profile.ClockOffsetGPU); err != nil {
			fmt.Printf("  [X] ClockOffsetGPU (%dMHz): %v\n", *profile.ClockOffsetGPU, err)
		} else {
			fmt.Printf("  [✔] ClockOffsetGPU (%dMHz)\n", *profile.ClockOffsetGPU)
		}
	} else {
		fmt.Printf("  [✔] ClockOffsetGPU skipped.\n")
	}

	if profile.ClockOffsetMem != nil {
		if err := dev.SetClockOffsetMem(*profile.ClockOffsetMem); err != nil {
			fmt.Printf("  [X] ClockOffsetMem (%dMHz): %v\n", *profile.ClockOffsetMem, err)
		} else {
			fmt.Printf("  [✔] ClockOffsetMem (%dMHz)\n", *profile.ClockOffsetMem)
		}
	} else {
		fmt.Printf("  [✔] ClockOffsetMem skipped.\n")
	}

	if profile.ClockLimitGPU != nil {
		if err := dev.SetClockLimitGPU(*profile.ClockLimitGPU); err != nil {
			fmt.Printf("  [X] ClockLimitGPU (%dMHz): %v\n", *profile.ClockLimitGPU, err)
		} else {
			fmt.Printf("  [✔] ClockLimitGPU (%dMHz)\n", *profile.ClockLimitGPU)
		}
	} else {
		fmt.Printf("  [✔] ClockLimitGPU skipped.\n")
	}

	if profile.FanControl != nil && *profile.FanControl && len(profile.FanCurve) > 0 {
		if err := config.ValidateFanCurve(profile.FanCurve); err != nil {
			fmt.Printf("  [!] FanCurve: %v — daemon will not manage fans.\n", err)
		} else {
			fmt.Printf("  [✔] FanCurve (%d points) — managed by daemon.\n", len(profile.FanCurve))
		}
	} else if profile.FanControl != nil && !*profile.FanControl {
		if err := dev.ResetFanSpeed(); err != nil {
			fmt.Printf("  [X] FanCurve (reset): %v\n", err)
		} else {
			fmt.Printf("  [✔] FanCurve (reset to vbios)\n")
		}
	} else {
		fmt.Printf("  [✔] FanCurve skipped.\n")
	}
}

func adjustFans(cfg *config.Config) {
	if cfg == nil {
		return
	}
	for _, dev := range Devices {
		profile, ok := cfg.Settings[fmt.Sprintf("%s | %s", dev.Name(), dev.PCIBusID())]
		if !ok || profile.FanControl == nil || !*profile.FanControl || len(profile.FanCurve) == 0 {
			continue
		}
		if config.ValidateFanCurve(profile.FanCurve) != nil {
			continue
		}
		temp, err := dev.Temperature()
		if err != nil {
			continue
		}
		_ = dev.SetFanSpeed(config.InterpolateFan(profile.FanCurve, temp))
	}
}

func resetFans(cfg *config.Config) {
	if cfg == nil {
		return
	}
	for _, dev := range Devices {
		profile, ok := cfg.Settings[fmt.Sprintf("%s | %s", dev.Name(), dev.PCIBusID())]
		if ok && profile.FanControl != nil && *profile.FanControl {
			_ = dev.ResetFanSpeed()
		}
	}
}
