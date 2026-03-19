//go:build windows

package cmd

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func doHealth() error {
	m, err := mgr.Connect()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to service manager:", err)
		os.Exit(1)
	}
	defer m.Disconnect()

	s, err := m.OpenService("gpuctl")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Not installed. Run: gpuctl install")
		os.Exit(1)
	}
	defer s.Close()

	cfg, err := s.Config()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to query service config:", err)
		os.Exit(1)
	}
	status, err := s.Query()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to query service status:", err)
		os.Exit(1)
	}

	enabled := cfg.StartType == mgr.StartAutomatic
	active := status.State == svc.Running

	switch {
	case enabled && active:
		fmt.Println("Enabled & Running.")
		fmt.Println("For details, run: sc query gpuctl")
	case enabled && !active:
		fmt.Fprintln(os.Stderr, "Enabled but not running.")
		os.Exit(1)
	case !enabled && active:
		fmt.Println("Running (not set to start automatically).")
	default:
		fmt.Fprintln(os.Stderr, "Not installed or disabled. Run: gpuctl install")
		os.Exit(1)
	}
	return nil
}
