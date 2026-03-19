package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhhc99/gpuctl/internal/config"
)

var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Apply configuration and notify the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigFromDisk()
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("config file not found at %s — run 'gpuctl conf init' first", config.ConfigPath)
			}
			return err
		}
		applyConfig(cfg)
		notifyDaemon()
		return nil
	},
}
