package sysutils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// CheckServiceStatus 检查自动启动是否已启用
// TODO: check
func CheckServiceStatus() bool {
	userName := GetUserName()
	if runtime.GOOS == "windows" {
		taskName := fmt.Sprintf("%s@%s", ServiceName, userName)
		return exec.Command("schtasks", "/query", "/tn", taskName).Run() == nil
	}
	serviceName := fmt.Sprintf("%s@%s.service", ServiceName, userName)
	return exec.Command("systemctl", "is-enabled", serviceName).Run() == nil
}

func InstallService() error {
	exe, _ := os.Executable()
	userName := GetUserName()

	if runtime.GOOS == "windows" {
		taskName := fmt.Sprintf("%s@%s", ServiceName, userName)
		// /sc ONLOGON: 登录时运行; /rl HIGHEST: 最高权限; /IT: 交互式
		// TODO: check
		args := []string{"/create", "/tn", taskName, "/tr", fmt.Sprintf("\"%s\" --apply-only", exe),
			"/sc", "ONLOGON", "/ru", "INTERACTIVE", "/rl", "HIGHEST", "/f"}
		return exec.Command("schtasks", args...).Run()
	}

	// Linux: 创建模板化的 systemd service
	// TODO: check
	servicePath := fmt.Sprintf("/etc/systemd/system/%s@.service", ServiceName)
	content := fmt.Sprintf(`[Unit]
Description=Apply GTU profiles for %%i
After=user@%%i.service

[Service]
Type=oneshot
ExecStart=/bin/sh -c 'GTU_TARGET_USER=%%i "%s" --apply-only'

[Install]
WantedBy=multi-user.target
`, exe)

	if err := os.WriteFile(servicePath, []byte(content), 0644); err != nil {
		return err
	}

	serviceName := fmt.Sprintf("%s@%s.service", ServiceName, userName)
	exec.Command("systemctl", "daemon-reload").Run()
	return exec.Command("systemctl", "enable", serviceName).Run()
}

// UninstallService 卸载服务
// TODO: check
func UninstallService() error {
	userName := GetUserName()
	if runtime.GOOS == "windows" {
		taskName := fmt.Sprintf("%s@%s", ServiceName, userName)
		return exec.Command("schtasks", "/delete", "/tn", taskName, "/f").Run()
	}
	serviceName := fmt.Sprintf("%s@%s.service", ServiceName, userName)
	exec.Command("systemctl", "disable", serviceName).Run()
	return nil
}
