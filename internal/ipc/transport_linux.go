//go:build linux

package ipc

import (
	"errors"
	"fmt"
	"net"
	"os"
)

const SockPath = "/run/gpuctl/gpud.sock"

// ErrAlreadyRunning is returned by listen() when another instance is active.
var ErrAlreadyRunning = errors.New("service is already running")

func listen() (net.Listener, error) {
	if err := os.MkdirAll("/run/gpuctl", 0755); err != nil {
		return nil, err
	}
	// Check for an active instance before clobbering the socket.
	if conn, err := net.Dial("unix", SockPath); err == nil {
		conn.Close()
		return nil, ErrAlreadyRunning
	}
	_ = os.Remove(SockPath)
	l, err := net.Listen("unix", SockPath)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(SockPath, 0666); err != nil {
		_ = l.Close()
		return nil, err
	}
	return l, nil
}

func dial() (net.Conn, error) {
	return net.Dial("unix", SockPath)
}

func isServicePresent() bool {
	_, err := os.Stat(SockPath)
	return err == nil
}

func serviceNotRunningErr() error {
	return fmt.Errorf("socket not found at %s", SockPath)
}
