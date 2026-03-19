//go:build windows

package cmd

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/svc"
)

type gpuSvc struct{}

func (g *gpuSvc) Execute(_ []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	s <- svc.Status{State: svc.StartPending}

	cfg, err := loadConfigFromDisk()
	if err != nil {
		fmt.Fprintf(os.Stderr, "daemon: config error: %v\n", err)
	} else {
		applyConfig(cfg)
	}

	quit := make(chan struct{})
	reload := make(chan struct{}, 1)
	go runFanLoop(quit, reload, cfg)

	s <- svc.Status{
		State:   svc.Running,
		Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptParamChange,
	}

	for c := range r {
		switch c.Cmd {
		case svc.Stop, svc.Shutdown:
			s <- svc.Status{State: svc.StopPending}
			close(quit)
			if latest, err := loadConfigFromDisk(); err == nil {
				resetFans(latest)
			}
			return false, 0
		case svc.ParamChange:
			select {
			case reload <- struct{}{}:
			default:
			}
		}
	}
	return false, 0
}

func runDaemon() error {
	return svc.Run("gpuctl", &gpuSvc{})
}
