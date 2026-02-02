package sysutils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

const ConfigFileName = "config.yaml"

func DefaultConfigPath() (string, error) {
	var home string
	if user, err := GetSessionUser(); err == nil {
		home = user.HomeDir
	} else {
		return "", fmt.Errorf("failed to get session user: %w", err)
	}

	if home == "" {
		return "", fmt.Errorf("could not determine home directory")
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(home, "AppData", "Roaming", "gpuctl", ConfigFileName), nil
	}
	return filepath.Join(home, ".config", "gpuctl", ConfigFileName), nil
}

func SaveFileAsSessionOwner(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		if user, err := GetSessionUser(); err == nil {
			uid, _ := strconv.Atoi(user.Uid)
			gid, _ := strconv.Atoi(user.Gid)
			os.Chown(path, uid, gid)
			os.Chown(dir, uid, gid)
		}
	}
	return nil
}
