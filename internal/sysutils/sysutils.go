package sysutils

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

const ServiceName = "gtu"

// GetUserName 获取真实用户名 (尊重 SUDO/PKEXEC 环境变量)
// TODO: verify
func GetUserName() string {
	if runtime.GOOS == "windows" {
		u, _ := user.Current()
		// Windows 用户名通常包含域名，如 "HOST\User"，取后半部分
		// TODO:????
		parts := strings.Split(u.Username, "\\")
		return parts[len(parts)-1]
	}

	if env := os.Getenv("GTU_TARGET_USER"); env != "" {
		return env
	}
	if uid := os.Getenv("SUDO_UID"); uid != "" {
		if u, err := user.LookupId(uid); err == nil {
			return u.Username
		}
	}
	if uid := os.Getenv("PKEXEC_UID"); uid != "" {
		if u, err := user.LookupId(uid); err == nil {
			return u.Username
		}
	}
	if env := os.Getenv("USER"); env != "" {
		return env
	}
	return "root"
}

// GetUserConfigPath returns ~/.config/gtu or %APPDATA%/gtu
func GetUserConfigPath() (string, error) {
	userName := GetUserName()
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("%%APPDATA%% env not found")
		}
		return filepath.Join(appData, ServiceName), nil
	}

	u, err := user.Lookup(userName)
	if err != nil {
		return "", err
	}
	return filepath.Join(u.HomeDir, ".config", ServiceName), nil
}

// ExecPrivileged 运行时提权重新启动自身
// TODO: WHY NEED THIS???
func ExecPrivileged(args ...string) error {
	exe, _ := os.Executable()
	userName := GetUserName()

	if runtime.GOOS == "windows" {
		// TODO: check
		psArgs := []string{
			"Start-Process",
			fmt.Sprintf("'%s'", exe),
			"-ArgumentList", fmt.Sprintf("'%s'", strings.Join(args, " ")),
			"-Verb", "RunAs",
			"-WindowStyle", "Hidden",
			"-Wait",
		}
		return exec.Command("powershell", "-Command", strings.Join(psArgs, " ")).Run()
	}

	// Linux: pkexec 重新运行，并通过 env 保持当前用户上下文
	// TODO: bro i use the fucking PKEXEC why do i need to
	//       set the fucking GTU_TARGET_USER???
	// alright then I set anyway, no side effects
	cmdArgs := []string{"env", fmt.Sprintf("GTU_TARGET_USER=%s", userName), exe}
	cmdArgs = append(cmdArgs, args...)
	return exec.Command("pkexec", cmdArgs...).Run()
}
