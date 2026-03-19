package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/zhhc99/gpuctl/internal/gpu"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List GPUs and their metrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tTEMP\tFAN\tPOWER\tUTIL\tCLOCK\tMEMORY")

		if len(deviceIndices) > 0 && !allDevices {
			for _, i := range deviceIndices {
				if i < 0 || i >= len(Devices) {
					return fmt.Errorf("device index %d out of range (0-%d)", i, len(Devices)-1)
				}
				printMetricRow(w, i, Devices[i])
			}
		} else {
			for i, dev := range Devices {
				printMetricRow(w, i, dev)
			}
		}
		w.Flush()
		return nil
	},
}

func init() {
	listCmd.Flags().IntSliceVarP(&deviceIndices, "device", "d", nil, "device indices (comma-separated)")
	listCmd.Flags().BoolVarP(&allDevices, "all", "a", false, "select all devices")
}

const na = "N/A"

func printMetricRow(w *tabwriter.Writer, i int, d gpu.Device) {
	fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		i, d.Name(),
		metricTemp(d), metricFan(d), metricPower(d),
		metricUtil(d), metricClock(d), metricMemory(d),
	)
}

func metricTemp(d gpu.Device) string {
	if v, err := d.Temperature(); err == nil {
		return fmt.Sprintf("%d°C", v)
	}
	return na
}

func metricFan(d gpu.Device) string {
	if pct, rpm, err := d.FanSpeed(); err == nil {
		if rpm != gpu.Unavailable {
			return fmt.Sprintf("%d%%/%drpm", pct, rpm)
		}
		return fmt.Sprintf("%d%%", pct)
	}
	return na
}

func metricPower(d gpu.Device) string {
	if w, err := d.Power(); err == nil {
		return fmt.Sprintf("%dW", w)
	}
	return na
}

func metricUtil(d gpu.Device) string {
	if g, m, err := d.Utilization(); err == nil {
		return fmt.Sprintf("G:%d%% M:%d%%", g, m)
	}
	return na
}

func metricClock(d gpu.Device) string {
	if g, m, err := d.Clocks(); err == nil {
		return fmt.Sprintf("G:%d M:%d", g, m)
	}
	return na
}

func metricMemory(d gpu.Device) string {
	if total, _, used, err := d.Memory(); err == nil {
		return fmt.Sprintf("%.1f/%.1fGB", float64(used)/(1<<30), float64(total)/(1<<30))
	}
	return na
}
