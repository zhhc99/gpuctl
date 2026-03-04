package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/zhhc99/gpuctl/internal/stress"
)

var (
	stressMode     string
	stressVRAM     string
	stressDuration string
)

var stressCmd = &cobra.Command{
	Use:   "stress",
	Short: "Stress the GPU with sustained compute workloads",
	Long: `Stress the GPU with sustained compute workloads.

Modes:
  alu    FMA loop -- saturates shader cores, maximizes core power draw
  mem    streaming read/write -- saturates VRAM bandwidth
  mixed  alternates alu and mem (default)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !stress.ValidMode(stressMode) {
			return fmt.Errorf("invalid mode %q: use alu, mem, or mixed", stressMode)
		}

		var timeout time.Duration
		if stressDuration != "" {
			var err error
			timeout, err = time.ParseDuration(stressDuration)
			if err != nil {
				return fmt.Errorf("invalid duration %q: %w", stressDuration, err)
			}
		}

		r, err := stress.NewRunner(stressVRAM, stress.Mode(stressMode))
		if err != nil {
			return err
		}

		if err := r.Init(); err != nil {
			return err
		}
		defer r.Close()

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		d := stress.NewDisplay(r.Stats())

		errCh := make(chan error, 1)
		go func() { errCh <- r.Run(ctx) }()

		d.Start()
		err = <-errCh
		d.Stop()

		return err
	},
}

func init() {
	stressCmd.Flags().StringVarP(&stressMode, "mode", "m", "mixed", "stress mode: alu, mem, mixed")
	stressCmd.Flags().StringVar(&stressVRAM, "vram", "1g", "VRAM to allocate, e.g. 512m, 2g")
	stressCmd.Flags().StringVar(&stressDuration, "duration", "", "run duration, e.g. 60s, 5m (default: until Ctrl+C)")
	rootCmd.AddCommand(stressCmd)
}
