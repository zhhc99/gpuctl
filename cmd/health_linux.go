//go:build linux

package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

func doHealth() error {
	enabled := serviceIsEnabled()
	active := serviceIsActive()

	switch {
	case enabled && active:
		fmt.Println("Enabled & Active.")
		fmt.Println("For details, run: systemctl status gpuctl.service")
	case enabled && !active:
		fmt.Fprintln(os.Stderr, "Enabled but not running.")
		os.Exit(1)
	case !enabled && active:
		fmt.Println("Running (not enabled on boot).")
	default:
		fmt.Fprintln(os.Stderr, "Not installed. Run: gpuctl install")
		os.Exit(1)
	}
	return nil
}

func serviceIsEnabled() bool {
	return exec.Command("systemctl", "is-enabled", "--quiet", "gpuctl.service").Run() == nil
}

func serviceIsActive() bool {
	return exec.Command("systemctl", "is-active", "--quiet", "gpuctl.service").Run() == nil
}
