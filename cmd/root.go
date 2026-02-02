package cmd

import (
	"fmt"

	"github.com/zhhc99/gpuctl/internal/gpu"
	"github.com/zhhc99/gpuctl/internal/nvml"

	"github.com/spf13/cobra"
)

var (
	deviceFlag []string

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
			// Don't fail immediately if NVML lib is missing,
			// but marked as nil so commands can handle it.
			// 我不理解.
			return nil
		}

		if err := Backend.Init(); err != nil {
			return fmt.Errorf("nvml init failed: %w", err)
		}

		Devices, err = Backend.GPUs()
		if err != nil {
			return fmt.Errorf("failed to list gpus: %w", err)
		}

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
	rootCmd.PersistentFlags().StringSliceVarP(&deviceFlag, "device", "d", nil, "device selection")
}
