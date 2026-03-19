package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/zhhc99/gpuctl/internal/gpu"
	"github.com/zhhc99/gpuctl/internal/nvml"
)

var (
	deviceIndices []int
	allDevices    bool

	Backend gpu.Backend
	Devices []gpu.Device
)

var rootCmd = &cobra.Command{
	Use:   "gpuctl",
	Short: "GPU monitoring and control utility",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		Backend, err = nvml.NewBackend()
		if err != nil {
			// NVML not available; commands that don't need GPU still work.
			return nil
		}
		if err := Backend.Init(); err != nil {
			return fmt.Errorf("nvml init failed: %w", err)
		}
		devs, err := Backend.GPUs()
		if err != nil {
			return fmt.Errorf("failed to list gpus: %w", err)
		}
		sort.Slice(devs, func(i, j int) bool {
			return devs[i].PCIBusID() < devs[j].PCIBusID()
		})
		Devices = devs
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if Backend != nil {
			Backend.Shutdown()
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.EnableCommandSorting = false
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(tuneCmd)
	rootCmd.AddCommand(confCmd)
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(daemonCmd)
}
