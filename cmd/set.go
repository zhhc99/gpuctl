package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zhhc99/gpuctl/internal/gpu"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set key=value...",
	Short: "Set GPU parameters",
	Long: `Set GPU parameters.

For a list of all available keys, run:
    gpuctl tune --help`,
	Example: `  gpuctl tune set power_limit=200
  gpuctl tune set cogpu=100 -d 0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no settings provided")
		}

		targets, err := resolveDevices()
		if err != nil {
			return err
		}

		updates := make(map[string]int)
		for _, arg := range args {
			parts := strings.Split(arg, "=")
			if len(parts) != 2 {
				return fmt.Errorf("invalid format: %s (expected key=value)", arg)
			}

			key, err := normalizeKey(parts[0])
			if err != nil {
				return err
			}

			val, err := strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("invalid value for %s: %v", parts[0], err)
			}
			updates[key] = val
		}

		for _, dev := range targets {
			fmt.Printf("Device %d (%s):\n", dev.Index(), dev.Name())
			for k, v := range updates {
				if err := applySet(dev, k, v); err != nil {
					fmt.Printf("  [X] %s=%d: %v\n", k, v, err)
				} else {
					fmt.Printf("  [✔] %s=%d\n", k, v)
				}
			}
		}

		return nil
	},
}

func init() {
	tuneCmd.AddCommand(setCmd)
}

func applySet(dev gpu.Device, key string, val int) error {
	switch key {
	case KeyPowerLimit:
		return dev.SetPowerLimit(val)
	case KeyClockOffsetGPU:
		return dev.SetClockOffsetGPU(val)
	case KeyClockOffsetMem:
		return dev.SetClockOffsetMem(val)
	case KeyClockLimitGPU:
		return dev.SetClockLimitGPU(val)
	default:
		return fmt.Errorf("unknown parameter")
	}
}
