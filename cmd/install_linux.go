//go:build linux

package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	linuxBinPath  = "/usr/local/bin/gpuctl"
	linuxUnitPath = "/etc/systemd/system/gpuctl.service"
	linuxUnitName = "gpuctl.service"
)

func doInstall() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("install requires root privileges — run with sudo")
	}

	// Step 1: install binary.
	if err := installBinary(linuxBinPath); err != nil {
		return err
	}

	// Step 2: write systemd unit.
	unit := fmt.Sprintf(`[Unit]
Description=gpuctl GPU controller
After=multi-user.target

[Service]
Type=simple
ExecStart=%s daemon
ExecReload=/bin/kill -HUP $MAINPID
ExecStop=%s tune reset fan --all
Restart=on-failure

[Install]
WantedBy=multi-user.target
`, linuxBinPath, linuxBinPath)

	fmt.Printf("Writing unit file to %s\n", linuxUnitPath)
	if err := os.WriteFile(linuxUnitPath, []byte(unit), 0644); err != nil {
		return fmt.Errorf("write unit file: %w", err)
	}

	fmt.Println("Enabling and starting service...")
	_ = exec.Command("systemctl", "daemon-reload").Run()
	if out, err := exec.Command("systemctl", "enable", "--now", linuxUnitName).CombinedOutput(); err != nil {
		return fmt.Errorf("enable service: %s", string(out))
	}

	fmt.Println("Done. Service is active.")
	return nil
}

func doUninstall() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("uninstall requires root privileges — run with sudo")
	}

	fmt.Printf("Stopping and disabling %s...\n", linuxUnitName)
	_ = exec.Command("systemctl", "disable", "--now", linuxUnitName).Run()

	fmt.Printf("Removing unit file %s\n", linuxUnitPath)
	if err := os.Remove(linuxUnitPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: could not remove unit file: %v\n", err)
	}
	_ = exec.Command("systemctl", "daemon-reload").Run()
	_ = exec.Command("systemctl", "reset-failed").Run()

	fmt.Printf("Removing binary %s\n", linuxBinPath)
	if err := os.Remove(linuxBinPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: could not remove binary: %v\n", err)
	}

	fmt.Println("Done. gpuctl has been uninstalled.")
	return nil
}
