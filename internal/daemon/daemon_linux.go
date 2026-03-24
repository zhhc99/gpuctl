//go:build linux

package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/zhhc99/gpuctl/internal/config"
	"github.com/zhhc99/gpuctl/internal/ipc"
	"github.com/zhhc99/gpuctl/internal/locale"
)

func runDaemon() error {
	if ipc.IsRunning() {
		return fmt.Errorf("%s", locale.T("err.service_already_running"))
	}

	if os.Geteuid() != 0 {
		return fmt.Errorf("%s", locale.T("err.need_root"))
	}

	signal.Ignore(syscall.SIGHUP)

	backendErr := initBackend()
	if backendErr != nil {
		fmt.Fprintf(os.Stderr, "daemon: backend error: %v\n", backendErr)
	} else {
		defer backend.Shutdown()
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "daemon: config error: %v\n", err)
	} else if backendErr == nil {
		printDeviceResults(applyConfig(cfg))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfgCh := make(chan *config.Config, 1)

	srv, err := ipc.NewServer(&daemonHandler{cfgCh: cfgCh, backendErr: backendErr})
	if err != nil {
		if errors.Is(err, ipc.ErrAlreadyRunning) {
			return fmt.Errorf("%s", locale.T("err.service_already_running"))
		}
		return fmt.Errorf(locale.T("err.service_start_failed"), err)
	}
	go srv.Serve()
	defer srv.Close()

	runFanLoop(ctx.Done(), cfgCh, cfg)

	if backendErr == nil {
		if c, err := config.Load(); err == nil {
			resetFans(c)
		}
	}
	return nil
}
