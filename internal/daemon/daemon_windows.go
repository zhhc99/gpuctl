//go:build windows

package daemon

import (
	"errors"
	"fmt"
	"os"

	"github.com/zhhc99/gpuctl/internal/config"
	"github.com/zhhc99/gpuctl/internal/ipc"
	"github.com/zhhc99/gpuctl/internal/locale"
	"golang.org/x/sys/windows/svc"
)

type gpuSvc struct{}

func (g *gpuSvc) Execute(_ []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	s <- svc.Status{State: svc.StartPending}

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

	cfgCh := make(chan *config.Config, 1)

	srv, err := ipc.NewServer(&daemonHandler{cfgCh: cfgCh, backendErr: backendErr})
	if err != nil {
		var msg string
		if errors.Is(err, ipc.ErrAlreadyRunning) {
			msg = locale.T("err.service_already_running")
		} else {
			msg = fmt.Sprintf(locale.T("err.service_start_failed"), err)
		}
		s <- svc.Status{State: svc.Stopped}
		fmt.Fprintln(os.Stderr, msg)
		return false, 1
	}
	go srv.Serve()
	defer srv.Close()

	quit := make(chan struct{})
	go runFanLoop(quit, cfgCh, cfg)

	s <- svc.Status{
		State:   svc.Running,
		Accepts: svc.AcceptStop | svc.AcceptShutdown,
	}

	for c := range r {
		switch c.Cmd {
		case svc.Stop, svc.Shutdown:
			s <- svc.Status{State: svc.StopPending}
			close(quit)
			if backendErr == nil {
				if latest, err := config.Load(); err == nil {
					resetFans(latest)
				}
			}
			return false, 0
		}
	}
	return false, 0
}

func runDaemon() error {
	return svc.Run("gpuctl", &gpuSvc{})
}
