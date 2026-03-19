//go:build windows

package config

import (
	"os"
	"path/filepath"
)

// ConfigPath is the machine-wide configuration file location.
var ConfigPath = filepath.Join(os.Getenv("PROGRAMDATA"), "gpuctl", "config.yaml")
