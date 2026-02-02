package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/zhhc99/gpuctl/internal/sysutils"

	"github.com/spf13/cobra"
)

const (
	ServiceName = "gpuctl"
	AppName     = "gpuctl"
	LinuxBin    = "/usr/local/bin/gpuctl"
	LinuxUnit   = "/etc/systemd/system/gpuctl@.service"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage startup service",
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the service to apply config on login",
	RunE: func(cmd *cobra.Command, args []string) error {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to locate executable: %w", err)
		}
		exe, err = filepath.Abs(exe)
		if err != nil {
			return fmt.Errorf("failed to locate executable: %w", err)
		}
		user, err := sysutils.GetSessionUser()
		if err != nil {
			return fmt.Errorf("failed to determine session user: %w", err)
		}
		userName := user.Username

		if runtime.GOOS == "windows" {
			taskName := fmt.Sprintf("%s@%s", ServiceName, userName)
			cmdStr := fmt.Sprintf("\"%s\" config apply", exe)

			schArgs := []string{
				"/create",
				"/tn", taskName,
				"/tr", cmdStr,
				"/sc", "ONLOGON",
				"/ru", "INTERACTIVE",
				"/rl", "HIGHEST",
				"/f",
			}

			if out, err := exec.Command("schtasks", schArgs...).CombinedOutput(); err != nil {
				return fmt.Errorf("schtasks failed: %s\nOutput: %s", err, string(out))
			}
			fmt.Printf("Service %s installed via Task Scheduler.\n", taskName)
			return nil
		}

		if exe != LinuxBin {
			src, err := os.Open(exe)
			if err != nil {
				return fmt.Errorf("failed to copy binary when opening source binary: %w", err)
			}
			defer src.Close()

			dst, err := os.OpenFile(LinuxBin, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return fmt.Errorf("failed to copy binary when creating target binary: %w", err)
			}
			defer dst.Close()

			if _, err = io.Copy(dst, src); err != nil {
				return fmt.Errorf("failed to copy binary: %w", err)
			}
			fmt.Printf("Binary copied to %s.\n", LinuxBin)
		} else {
			fmt.Printf("Binary found at %s.\n", LinuxBin)
		}

		servicePath := LinuxUnit
		content := fmt.Sprintf(`[Unit]
Description=Apply %s config profiles for %%i
After=user@%%i.service

[Service]
Type=oneshot
Environment="GPUCTL_TARGET_USER=%%i"
ExecStart=%s config apply

[Install]
WantedBy=multi-user.target
`, AppName, LinuxBin)

		if err := os.WriteFile(servicePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write service file: %w", err)
		}
		fmt.Printf("Service file written to %s\n", servicePath)

		instanceName := fmt.Sprintf("%s@%s.service", ServiceName, userName)

		if out, err := exec.Command("systemctl", "enable", instanceName).CombinedOutput(); err != nil {
			return fmt.Errorf("enable failed: %s", string(out))
		}

		fmt.Printf("Service enabled: %s\n", instanceName)
		return nil
	},
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := sysutils.GetSessionUser()
		if err != nil {
			return fmt.Errorf("failed to determine session user: %w", err)
		}
		userName := user.Username

		if runtime.GOOS == "windows" {
			taskName := fmt.Sprintf("%s@%s", ServiceName, userName)
			if out, err := exec.Command("schtasks", "/delete", "/tn", taskName, "/f").CombinedOutput(); err != nil {
				return fmt.Errorf("failed to delete task: %s\nOutput: %s", err, string(out))
			}
			fmt.Printf("Service %s removed.\n", taskName)
			return nil
		}

		instanceName := fmt.Sprintf("%s@%s.service", ServiceName, userName)
		if out, err := exec.Command("systemctl", "disable", instanceName).CombinedOutput(); err != nil {
			return fmt.Errorf("failed to disable service: %s", string(out))
		}
		fmt.Printf("Service disabled: %s\n", instanceName)
		return nil
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := sysutils.GetSessionUser()
		if err != nil {
			return fmt.Errorf("failed to determine session user: %w", err)
		}
		userName := user.Username

		if runtime.GOOS == "windows" {
			taskName := fmt.Sprintf("%s@%s", ServiceName, userName)
			cmd := exec.Command("schtasks", "/query", "/tn", taskName, "/fo", "LIST")
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Println("Service not found or not installed.")
				return nil
			}
			fmt.Println(string(out))
			return nil
		}

		instanceName := fmt.Sprintf("%s@%s.service", ServiceName, userName)
		out, _ := exec.Command("systemctl", "status", instanceName).CombinedOutput()
		fmt.Println(string(out))
		return nil
	},
}

func init() {
	serviceCmd.AddCommand(serviceInstallCmd)
	serviceCmd.AddCommand(serviceUninstallCmd)
	serviceCmd.AddCommand(serviceStatusCmd)
	rootCmd.AddCommand(serviceCmd)
}
