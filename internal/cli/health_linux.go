//go:build linux

package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/zhhc99/gpuctl/internal/ipc"
	"github.com/zhhc99/gpuctl/internal/locale"
)

func doHealth() error {
	online := ipc.IsRunning()
	enabled := serviceIsEnabled()

	if !online {
		fmt.Fprintln(os.Stderr, locale.T("msg.health_not_running"))
		os.Exit(1)
	}

	fmt.Println(locale.T("msg.health_running"))
	if enabled {
		fmt.Println(locale.T("msg.health_boot_enabled"))
	} else {
		fmt.Println(locale.T("msg.health_boot_not_enabled"))
	}
	fmt.Println(locale.T("msg.health_details_linux"))

	if resp, err := ipc.PostVersion(); err == nil && resp.Err == "" {
		fmt.Printf(locale.T("msg.version_backend")+"\n",
			resp.BackendName, resp.BackendVersion, resp.DriverVersion)
	}
	return nil
}

func serviceIsEnabled() bool {
	return exec.Command("systemctl", "is-enabled", "--quiet", "gpuctl.service").Run() == nil
}
