//go:build linux

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func runDaemon() error {
	cfg, err := loadConfigFromDisk()
	if err != nil {
		fmt.Fprintf(os.Stderr, "daemon: config error: %v\n", err)
	} else {
		applyConfig(cfg)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	reload := make(chan struct{}, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	go func() {
		for range sigs {
			select {
			case reload <- struct{}{}:
			default:
			}
		}
	}()

	runFanLoop(ctx.Done(), reload, cfg)

	if c, err := loadConfigFromDisk(); err == nil {
		resetFans(c)
	}
	return nil
}
