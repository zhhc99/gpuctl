package cmd

import (
	"fmt"

	"github.com/zhhc99/gpuctl/internal/gpu"

	"github.com/spf13/cobra"
)

var Version = "dev" // go build -ldflags "-X 'gpuctl/cmd.Version=v1.0.0'"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gpuctl version: %s\n", Version)
		if Backend != nil {
			var info gpu.BackendInfo
			info.Capture(Backend)
			fmt.Printf("Backend: %s (v%s)\n", info.ManagerName, info.ManagerVersion)
			fmt.Printf("Driver: %s\n", info.DriverVersion)
		} else {
			fmt.Println("Backend: None (NVML not initialized)")
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
