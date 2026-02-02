package cmd

import (
	"fmt"
	"gpuctl/internal/config"
	"gpuctl/internal/sysutils"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
}

var configWhereCmd = &cobra.Command{
	Use:   "where",
	Short: "Show configuration file location",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := sysutils.DefaultConfigPath()
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize an empty configuration file for detected GPUs",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := sysutils.DefaultConfigPath()
		if err != nil {
			return err
		}

		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config file already exists at %s", path)
		}

		fmt.Printf("Initializing config at %s...\n", path)

		cfg := config.Config{
			Settings: make(map[string]config.Profile),
		}

		for _, dev := range Devices {
			cfg.Settings[dev.UUID()] = config.Profile{}
		}

		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		if err := sysutils.SaveFileAsSessionOwner(path, data); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("Config initialized. Use 'gpuctl config edit' to modify.")
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := sysutils.DefaultConfigPath()
		if err != nil {
			return err
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Println("Config does not exist. Running init first...")
			if err := configInitCmd.RunE(cmd, args); err != nil {
				return err
			}
		}

		if runtime.GOOS == "windows" {
			return exec.Command("notepad", path).Run()
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			return fmt.Errorf("EDITOR environment variable not set. Config: %s", path)
		}

		c := exec.Command(editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

var configApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply configuration to GPUs",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := sysutils.DefaultConfigPath()
		if err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("config file not found at %s", path)
			}
			return err
		}

		var cfg config.Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}

		for _, dev := range Devices {
			profile, ok := cfg.Settings[dev.UUID()]
			if !ok {
				continue
			}

			fmt.Printf("Applying config to Device %d (%s)...\n", dev.Index(), dev.Name())

			if profile.PowerLimit != nil {
				if err := dev.SetPowerLimit(*profile.PowerLimit); err != nil {
					fmt.Printf("  [X] PowerLimit (%dW): %v\n", *profile.PowerLimit, err)
				} else {
					fmt.Printf("  [✔] PowerLimit (%dW)\n", *profile.PowerLimit)
				}
			} else {
				fmt.Printf("  [✔] PowerLimit skipped.\n")
			}

			if profile.ClockOffsetGPU != nil {
				if err := dev.SetClockOffsetGPU(*profile.ClockOffsetGPU); err != nil {
					fmt.Printf("  [X] ClockOffsetGPU (%dMHz): %v\n", *profile.ClockOffsetGPU, err)
				} else {
					fmt.Printf("  [✔] ClockOffsetGPU (%dMHz)\n", *profile.ClockOffsetGPU)
				}
			} else {
				fmt.Printf("  [✔] ClockOffsetGPU skipped.\n")
			}

			if profile.ClockOffsetMem != nil {
				if err := dev.SetClockOffsetMem(*profile.ClockOffsetMem); err != nil {
					fmt.Printf("  [X] ClockOffsetMem (%dMHz): %v\n", *profile.ClockOffsetMem, err)
				} else {
					fmt.Printf("  [✔] ClockOffsetMem (%dMHz)\n", *profile.ClockOffsetMem)
				}
			} else {
				fmt.Printf("  [✔] ClockOffsetMem skipped.\n")
			}

			if profile.ClockLimitGPU != nil {
				if err := dev.SetClockLimitGPU(*profile.ClockLimitGPU); err != nil {
					fmt.Printf("  [X] ClockLimitGPU (%dMHz): %v\n", *profile.ClockLimitGPU, err)
				} else {
					fmt.Printf("  [✔] ClockLimitGPU (%dMHz)\n", *profile.ClockLimitGPU)
				}
			} else {
				fmt.Printf("  [✔] ClockLimitGPU skipped.\n")
			}
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configWhereCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configApplyCmd)
	rootCmd.AddCommand(configCmd)
}
