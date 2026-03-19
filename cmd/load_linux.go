//go:build linux

package cmd

import (
	"fmt"
	"os/exec"
)

func notifyDaemon() {
	err := exec.Command("systemctl", "kill", "--signal=HUP", "--kill-whom=main", "gpuctl.service").Run()
	if err != nil {
		fmt.Println("Daemon not notified (not running or not installed).")
	} else {
		fmt.Println("Daemon notified.")
	}
}
