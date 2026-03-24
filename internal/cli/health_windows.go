//go:build windows

package cli

import (
	"fmt"
	"os"

	"github.com/zhhc99/gpuctl/internal/ipc"
	"github.com/zhhc99/gpuctl/internal/locale"
	"golang.org/x/sys/windows/svc/mgr"
)

func doHealth() error {
	online := ipc.IsRunning()

	if !online {
		fmt.Fprintln(os.Stderr, locale.T("msg.health_not_running"))
		os.Exit(1)
	}

	fmt.Println(locale.T("msg.health_running"))

	if m, err := mgr.Connect(); err == nil {
		defer m.Disconnect()
		if s, err := m.OpenService("gpuctl"); err == nil {
			defer s.Close()
			if cfg, err := s.Config(); err == nil {
				if cfg.StartType == mgr.StartAutomatic {
					fmt.Println(locale.T("msg.health_boot_enabled"))
				} else {
					fmt.Println(locale.T("msg.health_boot_not_enabled"))
				}
			}
			fmt.Println(locale.T("msg.health_details_windows"))
		}
	}

	if resp, err := ipc.PostVersion(); err == nil && resp.Err == "" {
		fmt.Printf(locale.T("msg.version_backend")+"\n",
			resp.BackendName, resp.BackendVersion, resp.DriverVersion)
	}
	return nil
}
