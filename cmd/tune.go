package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	KeyPowerLimit     = "power_limit"
	KeyClockOffsetGPU = "clock_offset_gpu"
	KeyClockOffsetMem = "clock_offset_mem"
	KeyClockLimitGPU  = "clock_limit_gpu"
)

var keyAliases = map[string]string{
	"pl":               KeyPowerLimit,
	"power_limit":      KeyPowerLimit,
	"cogpu":            KeyClockOffsetGPU,
	"clock_offset_gpu": KeyClockOffsetGPU,
	"comem":            KeyClockOffsetMem,
	"clock_offset_mem": KeyClockOffsetMem,
	"clgpu":            KeyClockLimitGPU,
	"clock_limit_gpu":  KeyClockLimitGPU,
}

var orderedKeys = []string{
	KeyPowerLimit, KeyClockOffsetGPU, KeyClockOffsetMem, KeyClockLimitGPU,
}

var tuneCmd = &cobra.Command{
	Use:   "tune",
	Short: "Tune GPU parameters",
	Long: `Tune GPU parameters.

Supported parameters for subcommands:
- pl,    power_limit:       Power limit           (Watt)
- cogpu, clock_offset_gpu:  GPU core clock offset (MHz)
- comem, clock_offset_mem:  Memory clock offset   (MHz)
- clgpu, clock_limit_gpu:   GPU core clock limit  (MHz)`,
}

func init() {
	rootCmd.AddCommand(tuneCmd)
}
func resolveTunableKeys(args []string) ([]string, error) {
	if len(args) == 0 {
		return orderedKeys, nil
	}
	var res []string
	for _, arg := range args {
		if k, ok := keyAliases[arg]; ok {
			res = append(res, k)
		} else {
			return nil, fmt.Errorf("invalid parameter: %s", arg)
		}
	}
	return res, nil
}

func normalizeKey(arg string) (string, error) {
	if k, ok := keyAliases[arg]; ok {
		return k, nil
	}
	return "", fmt.Errorf("invalid parameter: %s", arg)
}
