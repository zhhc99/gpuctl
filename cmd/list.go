package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all detected GPUs",
	Run: func(cmd *cobra.Command, args []string) {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tUUID\tBACKEND")

		backendName := "Unknown"
		if Backend != nil {
			backendName = Backend.Name()
		}

		for i, d := range Devices {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", i, d.Name(), d.UUID(), backendName)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
