//go:build windows

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

var windowsBinDir = filepath.Join(os.Getenv("PROGRAMFILES"), "gpuctl")
var windowsBinPath = filepath.Join(windowsBinDir, "gpuctl.exe")

func doInstall() error {
	if !isElevated() {
		return fmt.Errorf("install requires administrator privileges")
	}
	if err := installBinary(windowsBinPath); err != nil {
		return err
	}
	addToSystemPath(windowsBinDir)

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect to service manager: %w", err)
	}
	defer m.Disconnect()

	// If the service already exists, update its binary path and ensure it's running.
	if s, err := m.OpenService("gpuctl"); err == nil {
		defer s.Close()
		fmt.Println("Service already registered — updating binary path...")
		cfg, err := s.Config()
		if err != nil {
			return fmt.Errorf("query service config: %w", err)
		}
		cfg.BinaryPathName = windowsBinPath + " daemon"
		if err := s.UpdateConfig(cfg); err != nil {
			return fmt.Errorf("update service config: %w", err)
		}
		status, _ := s.Query()
		if status.State != svc.Running {
			fmt.Println("Starting service...")
			if err := s.Start(); err != nil {
				return fmt.Errorf("start service: %w", err)
			}
		}
		fmt.Println("Done. Service is running.")
		return nil
	}

	fmt.Println("Registering Windows service...")
	s, err := m.CreateService("gpuctl", windowsBinPath, mgr.Config{
		DisplayName: "gpuctl",
		Description: "gpuctl daemon - apply profiles & control fans",
		StartType:   mgr.StartAutomatic,
	}, "daemon")
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	defer s.Close()

	fmt.Println("Starting service...")
	if err := s.Start(); err != nil {
		return fmt.Errorf("start service: %w", err)
	}
	fmt.Println("Done. Service is running.")
	return nil
}

func doUninstall() error {
	if !isElevated() {
		return fmt.Errorf("uninstall requires administrator privileges")
	}

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService("gpuctl")
	if err == nil {
		fmt.Println("Stopping service...")
		_, _ = s.Control(svc.Stop)
		fmt.Println("Unregistering service...")
		_ = s.Delete()
		s.Close()
	}

	fmt.Printf("Removing binary %s\n", windowsBinPath)
	if err := os.Remove(windowsBinPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: could not remove binary: %v\n", err)
	}
	removeFromSystemPath(windowsBinDir)

	fmt.Println("Done. gpuctl has been uninstalled.")
	return nil
}

func isElevated() bool {
	f, err := os.Open(`\\.\PHYSICALDRIVE0`)
	if err == nil {
		f.Close()
		return true
	}
	return false
}

func addToSystemPath(dir string) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		fmt.Printf("Warning: could not open system PATH: %v\n", err)
		return
	}
	defer k.Close()

	cur, _, err := k.GetStringValue("Path")
	if err != nil {
		return
	}
	for _, p := range strings.Split(cur, ";") {
		if strings.EqualFold(strings.TrimRight(p, `\`), strings.TrimRight(dir, `\`)) {
			return // already present
		}
	}
	if err := k.SetStringValue("Path", cur+";"+dir); err != nil {
		fmt.Printf("Warning: could not update system PATH: %v\n", err)
		return
	}
	fmt.Printf("Added %s to system PATH.\n", dir)
}

func removeFromSystemPath(dir string) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return
	}
	defer k.Close()

	cur, _, err := k.GetStringValue("Path")
	if err != nil {
		return
	}
	parts := strings.Split(cur, ";")
	filtered := parts[:0]
	for _, p := range parts {
		if !strings.EqualFold(strings.TrimRight(p, `\`), strings.TrimRight(dir, `\`)) {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) == len(parts) {
		return // not found, nothing to do
	}
	if err := k.SetStringValue("Path", strings.Join(filtered, ";")); err != nil {
		fmt.Printf("Warning: could not update system PATH: %v\n", err)
		return
	}
	fmt.Printf("Removed %s from system PATH.\n", dir)
}
