package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/zhhc99/gpuctl/internal/ipc"
	"github.com/zhhc99/gpuctl/internal/locale"
)

var keyAliases = map[string]string{
	"pl":               "power_limit",
	"power_limit":      "power_limit",
	"cogpu":            "clock_offset_gpu",
	"clock_offset_gpu": "clock_offset_gpu",
	"comem":            "clock_offset_mem",
	"clock_offset_mem": "clock_offset_mem",
	"clgpu":            "clock_limit_gpu",
	"clock_limit_gpu":  "clock_limit_gpu",
	"fan":              "fan",
}

func normalizeKey(s string) (string, error) {
	if k, ok := keyAliases[s]; ok {
		return k, nil
	}
	return "", fmt.Errorf(locale.T("err.unknown_param"), s)
}

func runTune(args []string) error {
	if len(args) == 0 {
		fmt.Println(locale.T("help.tune"))
		return nil
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "get":
		return runTuneGet(rest)
	case "set":
		return runTuneSet(rest)
	case "reset":
		return runTuneReset(rest)
	default:
		fmt.Fprintf(os.Stderr, locale.T("err.unknown_subcmd"), sub, "tune", "tune")
		return fmt.Errorf("exit 1")
	}
}

func runTuneGet(args []string) error {
	_, indices, all, err := parseDeviceFlags(args)
	if err != nil {
		return err
	}
	resp, err := ipc.PostTuneGet(ipc.TuneGetRequest{Indices: indices, All: all})
	if err != nil {
		return err
	}
	if resp.Err != "" {
		return fmt.Errorf(locale.T("err.daemon"), resp.Err)
	}
	for _, ds := range resp.Devices {
		fmt.Printf(locale.T("msg.device_header")+"\n", ds.Index, ds.Name)
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, locale.T("msg.tune_get_header"))
		for _, row := range ds.Rows {
			fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n",
				row.Name, row.Current, row.Default, row.Min, row.Max)
		}
		w.Flush()
		fmt.Println()
	}
	return nil
}

func runTuneSet(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%s", locale.T("err.no_params"))
	}
	kvArgs, indices, all, err := parseDeviceFlags(args)
	if err != nil {
		return err
	}
	var kvTokens []string
	for _, a := range kvArgs {
		if strings.Contains(a, "=") {
			kvTokens = append(kvTokens, a)
		} else {
			return fmt.Errorf(locale.T("err.invalid_kv"), a)
		}
	}
	if len(kvTokens) == 0 {
		return fmt.Errorf("%s", locale.T("err.no_params"))
	}
	updates, err := parseKeyValues(kvTokens)
	if err != nil {
		return err
	}
	resp, err := ipc.PostTuneSet(ipc.TuneSetRequest{Indices: indices, All: all, Updates: updates})
	if err != nil {
		return err
	}
	if resp.Err != "" {
		return fmt.Errorf(locale.T("err.daemon"), resp.Err)
	}
	printTuneResults(resp.Devices)
	return nil
}

func runTuneReset(args []string) error {
	paramArgs, indices, all, err := parseDeviceFlags(args)
	if err != nil {
		return err
	}
	var keys []string
	for _, arg := range paramArgs {
		k, err := normalizeKey(arg)
		if err != nil {
			return err
		}
		keys = append(keys, k)
	}
	resp, err := ipc.PostTuneReset(ipc.TuneResetRequest{Indices: indices, All: all, Keys: keys})
	if err != nil {
		return err
	}
	if resp.Err != "" {
		return fmt.Errorf(locale.T("err.daemon"), resp.Err)
	}
	printTuneResults(resp.Devices)
	return nil
}

func printTuneResults(devices []ipc.DeviceResult) {
	for _, dr := range devices {
		fmt.Printf(locale.T("msg.device_header")+"\n", dr.Index, dr.Name)
		for _, f := range dr.Fields {
			if f.Err != "" {
				fmt.Printf(locale.T("msg.field_err")+"\n", f.Field, f.Err)
			} else {
				fmt.Printf(locale.T("msg.field_ok")+"\n", f.Field)
			}
		}
	}
}

func parseKeyValues(args []string) (map[string]int, error) {
	updates := make(map[string]int)
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf(locale.T("err.invalid_kv"), arg)
		}
		key, err := normalizeKey(parts[0])
		if err != nil {
			return nil, err
		}
		var val int
		if _, err := fmt.Sscanf(parts[1], "%d", &val); err != nil {
			return nil, fmt.Errorf(locale.T("err.invalid_param_val"), parts[0], err)
		}
		updates[key] = val
	}
	return updates, nil
}
