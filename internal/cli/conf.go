package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/zhhc99/gpuctl/internal/config"
	"github.com/zhhc99/gpuctl/internal/gpu"
	"github.com/zhhc99/gpuctl/internal/ipc"
	"github.com/zhhc99/gpuctl/internal/locale"
)

func runConf(args []string) error {
	if len(args) == 0 {
		fmt.Println(locale.T("help.conf"))
		return nil
	}
	sub := args[0]
	switch sub {
	case "init":
		return runConfInit()
	case "edit":
		return runConfEdit()
	default:
		fmt.Fprintf(os.Stderr, locale.T("err.unknown_subcmd"), sub, "conf", "conf")
		return fmt.Errorf("exit 1")
	}
}

func runConfInit() error {
	if err := checkPrivileged(); err != nil {
		return err
	}

	resp, err := ipc.PostList(ipc.ListRequest{All: true})
	if err != nil {
		return err
	}
	if resp.Err != "" {
		return fmt.Errorf(locale.T("err.daemon"), resp.Err)
	}
	if len(resp.Devices) == 0 {
		return fmt.Errorf("%s", locale.T("err.no_gpus"))
	}

	path := config.ConfigPath
	if _, err := os.Stat(path); err == nil {
		fmt.Printf(locale.T("msg.config_exists")+"\n", path)
		if !confirm(locale.T("msg.overwrite_prompt"), true) {
			return fmt.Errorf("%s", locale.T("msg.aborted"))
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf(locale.T("err.create_config_dir"), err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf(locale.T("err.create_config"), err)
	}
	defer f.Close()

	if err := configTemplate().Execute(f, resp.Devices); err != nil {
		return fmt.Errorf(locale.T("err.write_config"), err)
	}
	fmt.Printf(locale.T("msg.config_written")+"\n", path)
	fmt.Println(locale.T("msg.config_hint"))
	return nil
}

func runConfEdit() error {
	path := config.ConfigPath
	fmt.Printf(locale.T("msg.opening_file")+"\n", path)
	return openConfigEditor(path)
}

func confirm(prompt string, defaultYes bool) bool {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return defaultYes
	}
	switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return defaultYes
	}
}

func configTemplate() *template.Template {
	return template.Must(template.New("config").Funcs(template.FuncMap{
		"buskey": func(s gpu.Snapshot) string {
			return fmt.Sprintf("%s | %s", s.Name, s.BusID)
		},
	}).Parse(strings.Join([]string{
		"settings: # " + locale.T("msg.conf_tpl_ignore"),
		"{{- range .}}",
		`  {{buskey . | printf "%q"}}:`,
		"    power_limit: ~      # " + locale.T("msg.conf_tpl_power_limit"),
		"    clock_offset_gpu: ~ # " + locale.T("msg.conf_tpl_clock_offset_gpu"),
		"    clock_offset_mem: ~ # " + locale.T("msg.conf_tpl_clock_offset_mem"),
		"    clock_limit_gpu: ~  # " + locale.T("msg.conf_tpl_clock_limit_gpu"),
		"    fan_control: ~      # " + locale.T("msg.conf_tpl_fan_control"),
		"    fan_curve:          # " + locale.T("msg.conf_tpl_fan_curve"),
		"      - temp: 40",
		"        fan: 30",
		"      - temp: 50",
		"        fan: 30",
		"      - temp: 60",
		"        fan: 45",
		"      - temp: 90",
		"        fan: 100",
		"{{- end}}",
		"",
	}, "\n")))
}
