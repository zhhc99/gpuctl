package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/zhhc99/gpuctl/internal/gpu"

	"github.com/spf13/cobra"
)

var specCmd = &cobra.Command{
	Use:   "spec [params...]",
	Short: "Show tuning ranges and defaults",
	Example: `  gpuctl tune spec
  gpuctl tune spec pl -d 0`,
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
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "PARAMETER\tCURRENT\tDEFAULT\tMIN\tMAX")

			for _, key := range keys {
				printSpecRow(w, dev, key)
			}
			w.Flush()
			fmt.Println()
		}
		return nil
	},
}

func init() {
	tuneCmd.AddCommand(specCmd)
}

func printSpecRow(w *tabwriter.Writer, dev gpu.Device, key string) {
	const NA = "N/A"
	var (
		name           string
		cur, def       string
		minStr, maxStr string
		unit           string
	)

	switch key {
	case KeyPowerLimit:
		name, unit = "Power Limit", "W"
		cur = formatVal(dev.PowerLimit())
		def = formatVal(dev.PowerLimitDefault())
		minStr, maxStr = formatRange(dev.PowerLimitRange())
	case KeyClockOffsetGPU:
		name, unit = "Clock Offset GPU", "MHz"
		cur = formatSigned(dev.ClockOffsetGPU())
		def = "0"
		minStr, maxStr = formatRange(dev.ClockOffsetGPURange())
	case KeyClockOffsetMem:
		name, unit = "Clock Offset Mem", "MHz"
		cur = formatSigned(dev.ClockOffsetMem())
		def = "0"
		minStr, maxStr = formatRange(dev.ClockOffsetMemRange())
	case KeyClockLimitGPU:
		name, unit = "Clock Limit GPU", "MHz"
		cur = formatVal(dev.ClockLimitGPU())
		if _, max, err := dev.ClockLimitGPURange(); err == nil {
			def = fmt.Sprintf("%d", max)
			maxStr = fmt.Sprintf("%d", max)
			if min, _, err := dev.ClockLimitGPURange(); err == nil {
				minStr = fmt.Sprintf("%d", min)
			} else {
				minStr = NA
			}
		} else {
			def, minStr, maxStr = NA, NA, NA
		}
	}

	if cur != NA {
		cur += unit
	}
	if def != NA && def != "0" {
		def += unit
	}
	if minStr != NA {
		minStr += unit
	}
	if maxStr != NA {
		maxStr += unit
	}

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, cur, def, minStr, maxStr)
}

func formatVal(v int, err error) string {
	if err != nil {
		return "N/A"
	}
	return fmt.Sprintf("%d", v)
}

func formatSigned(v int, err error) string {
	if err != nil {
		return "N/A"
	}
	return fmt.Sprintf("%+d", v)
}

func formatRange(min, max int, err error) (string, string) {
	if err != nil {
		return "N/A", "N/A"
	}
	return fmt.Sprintf("%d", min), fmt.Sprintf("%d", max)
}
