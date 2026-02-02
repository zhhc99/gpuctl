package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/zhhc99/gpuctl/internal/gpu"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [params...]",
	Short: "Get GPU parameters (temp, power, util, clock, fan, memory)",
	RunE: func(cmd *cobra.Command, args []string) error {
		targets, err := resolveDevices()
		if err != nil {
			return err
		}

		allKeys := []string{"temp", "fan", "power", "util", "clock", "memory"}
		reqs := args

		if len(reqs) == 0 {
			reqs = allKeys
		} else {
			validMap := make(map[string]bool)
			for _, k := range allKeys {
				validMap[k] = true
			}
			for _, r := range reqs {
				if !validMap[r] {
					return fmt.Errorf("invalid parameter: %s", r)
				}
			}
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		// 表头
		headers := []string{"ID", "NAME"}
		for _, r := range reqs {
			headers = append(headers, strings.ToUpper(r))
		}
		fmt.Fprintln(w, strings.Join(headers, "\t"))

		// 行
		for _, dev := range targets {
			row := []string{
				fmt.Sprintf("%d", dev.Index()),
				dev.Name(),
			}

			for _, r := range reqs {
				val := fetchValue(dev, r)
				row = append(row, val)
			}
			fmt.Fprintln(w, strings.Join(row, "\t"))
		}
		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}

func fetchValue(d gpu.Device, key string) string {
	const NA = "N/A"
	switch key {
	case "temp":
		if v, err := d.Temperature(); err == nil {
			return fmt.Sprintf("%dC", v)
		}
	case "fan":
		if pct, rpm, err := d.FanSpeed(); err == nil {
			if rpm != gpu.Unavailable {
				return fmt.Sprintf("%d%%/%drpm", pct, rpm)
			}
			return fmt.Sprintf("%d%%", pct)
		}
	case "power":
		if w, err := d.Power(); err == nil {
			return fmt.Sprintf("%dW", w)
		}
	case "util":
		if g, m, err := d.Utilization(); err == nil {
			return fmt.Sprintf("G:%d%% M:%d%%", g, m)
		}
	case "clock":
		if g, m, err := d.Clocks(); err == nil {
			return fmt.Sprintf("G:%d M:%d", g, m)
		}
	case "memory":
		if t, f, _, err := d.Memory(); err == nil {
			return fmt.Sprintf("%d/%dMB", (t-f)/1024/1024, t/1024/1024)
		}
	}
	return NA
}
