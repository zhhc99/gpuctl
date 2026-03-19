//go:build windows

package cmd

import (
	"fmt"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func notifyDaemon() {
	m, err := mgr.Connect()
	if err != nil {
		fmt.Println("Daemon not notified.")
		return
	}
	defer m.Disconnect()
	s, err := m.OpenService("gpuctl")
	if err != nil {
		fmt.Println("Daemon not notified (not installed?).")
		return
	}
	defer s.Close()
	if _, err := s.Control(svc.ParamChange); err != nil {
		fmt.Println("Daemon not notified.")
		return
	}
	fmt.Println("Daemon notified.")
}
