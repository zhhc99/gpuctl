package cmd

import (
	"fmt"
	"gpuctl/internal/gpu"

	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset [params...]",
	Short: "Reset GPU parameters to default",
	Long: `Reset GPU parameters to default.

For a list of all available keys and their descriptions, run:
    gpuctl tune --help`,
	Example: `  gpuctl tune reset (reset all)
  gpuctl tune reset pl`,
	RunE: func(cmd *cobra.Command, args []string) error {
		keys, err := resolveTunableKeys(args)
		if err != nil {
			return err
		}

		targets, err := resolveDevices()
		if err != nil {
			return err
		}

		for _, dev := range targets {
			fmt.Printf("Device %d (%s):\n", dev.Index(), dev.Name())
			for _, key := range keys {
				if err := applyReset(dev, key); err != nil {
					fmt.Printf("  [X] Reset %s: %v\n", key, err)
				} else {
					fmt.Printf("  [✔] Reset %s\n", key)
				}
			}
		}
		return nil
	},
}

func init() {
	tuneCmd.AddCommand(resetCmd)
}

func applyReset(dev gpu.Device, key string) error {
	switch key {
	case KeyPowerLimit:
		return dev.ResetPowerLimit()
	case KeyClockOffsetGPU:
		return dev.ResetClockOffsetGPU()
	case KeyClockOffsetMem:
		return dev.ResetClockOffsetMem()
	case KeyClockLimitGPU:
		return dev.ResetClockLimitGPU()
	default:
		return fmt.Errorf("unknown parameter: %s", key)
	}
}
