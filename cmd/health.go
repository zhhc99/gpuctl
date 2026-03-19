package cmd

import (
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Show daemon health status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doHealth()
	},
}

