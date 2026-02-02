//go:build linux

package sysutils

import (
	"os"
	"os/user"
	"strings"
)

func GetSessionUser() (*user.User, error) {
	// 环境变量 (systemd)
	if target := os.Getenv("GPUCTL_TARGET_USER"); target != "" {
		return user.Lookup(target)
	}

	// 会话中提权 (sudo, pkexec, ...)
	if uid, ok := readLoginUID(); ok {
		return user.LookupId(uid)
	}

	// fallback
	return user.Current()
}

func readLoginUID() (string, bool) {
	data, err := os.ReadFile("/proc/self/loginuid")
	if err != nil {
		return "", false
	}
	uid := strings.TrimSpace(string(data))
	if uid == "" || uid == "4294967295" {
		return "", false // 4294... = (uint32)-1 也表示空值
	}
	return uid, true
}
