package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/zhhc99/gpuctl/internal/gpu"
)

const (
	keyPowerLimit     = "power_limit"
	keyClockOffsetGPU = "clock_offset_gpu"
	keyClockOffsetMem = "clock_offset_mem"
	keyClockLimitGPU  = "clock_limit_gpu"
	keyFan            = "fan"
)

var keyAliases = map[string]string{
	"pl":               keyPowerLimit,
	"power_limit":      keyPowerLimit,
	"cogpu":            keyClockOffsetGPU,
	"clock_offset_gpu": keyClockOffsetGPU,
	"comem":            keyClockOffsetMem,
	"clock_offset_mem": keyClockOffsetMem,
	"clgpu":            keyClockLimitGPU,
	"clock_limit_gpu":  keyClockLimitGPU,
	"fan":              keyFan,
}

var orderedKeys = []string{
	keyPowerLimit, keyClockOffsetGPU, keyClockOffsetMem, keyClockLimitGPU, keyFan,
}

var tuneCmd = &cobra.Command{
	Use:   "tune",
	Short: "Set/Reset GPU parameters (not persistent)",
	Long: `Set/Reset GPU parameters (not persistent).

Supported parameters:
  pl:    Power Limit (W)
  cogpu: Clock Offset of GPU (MHz)
  comem: Clock Offset of Memory (MHz)
  clgpu: Clock Limit of GPU (MHz)
  fan:   Fan Speed (%)`,
}

var tuneGetCmd = &cobra.Command{
	Use:   "get [params...]",
	Short: "Show tuning ranges and current values",
	RunE: func(cmd *cobra.Command, args []string) error {
		keys, err := resolveKeys(args)
		if err != nil {
			return err
		}
		for _, i := range makeIndices(true) {
			dev := Devices[i]
			fmt.Printf("Device %d (%s):\n", i, dev.Name())
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "  PARAMETER\tCURRENT\tDEFAULT\tMIN\tMAX")
			for _, k := range keys {
				printSpecRow(w, dev, k)
			}
			w.Flush()
			fmt.Println()
		}
		return nil
	},
}

var tuneSetCmd = &cobra.Command{
	Use:   "set key=value...",
	Short: "Set GPU parameters",
	Example: `  gpuctl tune set pl=100 -d 0
  gpuctl tune set cogpu=210 clgpu=2520 -a
  gpuctl tune set fan=30 -d 0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no parameters provided")
		}
		updates, err := parseKeyValues(args)
		if err != nil {
			return err
		}
		indices, err := tuneSetIndices()
		if err != nil {
			return err
		}
		for _, i := range indices {
			fmt.Printf("Device %d (%s):\n", i, Devices[i].Name())
			for _, k := range orderedKeys {
				v, ok := updates[k]
				if !ok {
					continue
				}
				if err := applySet(Devices[i], k, v); err != nil {
					fmt.Printf("  [X] %s=%d: %v\n", k, v, err)
				} else {
					fmt.Printf("  [✔] %s=%d\n", k, v)
				}
			}
		}
		return nil
	},
}

var tuneResetCmd = &cobra.Command{
	Use:   "reset [params...]",
	Short: "Reset GPU parameters to default",
	Example: `  gpuctl tune reset -a
  gpuctl tune reset pl -d 0
  gpuctl tune reset fan -d 0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		keys, err := resolveKeys(args)
		if err != nil {
			return err
		}
		for _, i := range makeIndices(true) {
			fmt.Printf("Device %d (%s):\n", i, Devices[i].Name())
			for _, k := range keys {
				if err := applyReset(Devices[i], k); err != nil {
					fmt.Printf("  [X] reset %s: %v\n", k, err)
				} else {
					fmt.Printf("  [✔] reset %s\n", k)
				}
			}
		}
		return nil
	},
}

func init() {
	tuneCmd.PersistentFlags().IntSliceVarP(&deviceIndices, "device", "d", nil, "device indices (comma-separated)")
	tuneCmd.PersistentFlags().BoolVarP(&allDevices, "all", "a", false, "select all devices")
	tuneCmd.AddCommand(tuneGetCmd, tuneSetCmd, tuneResetCmd)
}

// makeIndices returns target indices: explicit -d, all if -a or defaultAll, else all devices.
func makeIndices(defaultAll bool) []int {
	if len(deviceIndices) > 0 && !allDevices {
		return deviceIndices
	}
	idx := make([]int, len(Devices))
	for i := range Devices {
		idx[i] = i
	}
	return idx
}

// tuneSetIndices requires explicit -d or -a when multiple GPUs are present.
func tuneSetIndices() ([]int, error) {
	if allDevices {
		return makeIndices(true), nil
	}
	if len(deviceIndices) > 0 {
		for _, i := range deviceIndices {
			if i < 0 || i >= len(Devices) {
				return nil, fmt.Errorf("device index %d out of range (0-%d)", i, len(Devices)-1)
			}
		}
		return deviceIndices, nil
	}
	if len(Devices) > 1 {
		return nil, fmt.Errorf("multiple GPUs detected — use -d or -a")
	}
	return []int{0}, nil
}

func printSpecRow(w *tabwriter.Writer, dev gpu.Device, key string) {
	fv := func(v int, err error) string {
		if err != nil || v == gpu.Unavailable {
			return na
		}
		return strconv.Itoa(v)
	}
	fs := func(v int, err error) string {
		if err != nil || v == gpu.Unavailable {
			return na
		}
		return fmt.Sprintf("%+d", v)
	}

	var name, unit, cur, def, lo, hi string
	switch key {
	case keyPowerLimit:
		name, unit = "PowerLimit", "W"
		cur = fv(dev.PowerLimit())
		def = fv(dev.PowerLimitDefault())
		mn, mx, err := dev.PowerLimitRange()
		lo, hi = fv(mn, err), fv(mx, err)
	case keyClockOffsetGPU:
		name, unit = "ClockOffsetGPU", "MHz"
		cur = fs(dev.ClockOffsetGPU())
		def = "+0"
		mn, mx, err := dev.ClockOffsetGPURange()
		lo, hi = fs(mn, err), fs(mx, err)
	case keyClockOffsetMem:
		name, unit = "ClockOffsetMem", "MHz"
		cur = fs(dev.ClockOffsetMem())
		def = "+0"
		mn, mx, err := dev.ClockOffsetMemRange()
		lo, hi = fs(mn, err), fs(mx, err)
	case keyClockLimitGPU:
		name, unit = "ClockLimitGPU", "MHz"
		cur = fv(dev.ClockLimitGPU())
		mn, mx, err := dev.ClockLimitGPURange()
		lo, hi = fv(mn, err), fv(mx, err)
		def = hi
	case keyFan:
		name, unit = "Fan", "%"
		pct, _, err := dev.FanSpeed()
		cur = fv(pct, err)
		def, lo, hi = "auto", "0", "100"
	}

	addUnit := func(s string) string {
		if s == na {
			return s
		}
		return s + unit
	}
	fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n",
		name, addUnit(cur), addUnit(def), addUnit(lo), addUnit(hi))
}

func parseKeyValues(args []string) (map[string]int, error) {
	updates := make(map[string]int)
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format %q (expected key=value)", arg)
		}
		key, err := normalizeKey(parts[0])
		if err != nil {
			return nil, err
		}
		val, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s: %v", parts[0], err)
		}
		updates[key] = val
	}
	return updates, nil
}

func normalizeKey(s string) (string, error) {
	if k, ok := keyAliases[s]; ok {
		return k, nil
	}
	return "", fmt.Errorf("unknown parameter %q", s)
}

func resolveKeys(args []string) ([]string, error) {
	if len(args) == 0 {
		return orderedKeys, nil
	}
	keys := make([]string, 0, len(args))
	for _, arg := range args {
		k, err := normalizeKey(arg)
		if err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func applySet(dev gpu.Device, key string, val int) error {
	switch key {
	case keyPowerLimit:
		return dev.SetPowerLimit(val)
	case keyClockOffsetGPU:
		return dev.SetClockOffsetGPU(val)
	case keyClockOffsetMem:
		return dev.SetClockOffsetMem(val)
	case keyClockLimitGPU:
		return dev.SetClockLimitGPU(val)
	case keyFan:
		return dev.SetFanSpeed(val)
	default:
		return fmt.Errorf("unknown parameter")
	}
}

func applyReset(dev gpu.Device, key string) error {
	switch key {
	case keyPowerLimit:
		return dev.ResetPowerLimit()
	case keyClockOffsetGPU:
		return dev.ResetClockOffsetGPU()
	case keyClockOffsetMem:
		return dev.ResetClockOffsetMem()
	case keyClockLimitGPU:
		return dev.ResetClockLimitGPU()
	case keyFan:
		return dev.ResetFanSpeed()
	default:
		return fmt.Errorf("unknown parameter %q", key)
	}
}
