package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/zhhc99/gpuctl/internal/config"
	"github.com/zhhc99/gpuctl/internal/gpu"
	"github.com/spf13/cobra"
)

var confCmd = &cobra.Command{
	Use:   "conf",
	Short: "Configuration management",
}

var confInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize (or overwrite) the configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if Backend == nil || len(Devices) == 0 {
			return fmt.Errorf("no GPUs detected — cannot initialize config")
		}
		path := config.ConfigPath
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
		defer f.Close()
		if err := configTmpl.Execute(f, Devices); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
		fmt.Printf("Config written to %s\n", path)
		fmt.Println("Edit with 'gpuctl conf edit', then apply with 'gpuctl load'.")
		return nil
	},
}

var confEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open the configuration file in an editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.ConfigPath
		fmt.Printf("Opening %s\n", path)
		return openConfigEditor(path)
	},
}

func init() {
	confCmd.AddCommand(confInitCmd)
	confCmd.AddCommand(confEditCmd)
}

// yaml.Marshal cannot produce comments or the preset fan curve layout,
// so we use a template instead.
var configTmpl = template.Must(template.New("config").Funcs(template.FuncMap{
	"buskey": func(dev gpu.Device) string {
		return fmt.Sprintf("%s | %s", dev.Name(), dev.PCIBusID())
	},
}).Parse(`settings: # set value to null or ~ to ignore
{{- range .}}
  {{buskey . | printf "%q"}}:
    power_limit: ~      # Watt
    clock_offset_gpu: ~ # MHz
    clock_offset_mem: ~ # MHz
    clock_limit_gpu: ~  # MHz
    fan_control: ~      # true: use fan curve; false: reset to vbios
    fan_curve:          # temp(°C) -> fan(%)
      - temp: 40
        fan: 30
      - temp: 50
        fan: 30
      - temp: 60
        fan: 45
      - temp: 90
        fan: 100
{{- end}}
`))
