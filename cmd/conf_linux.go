//go:build linux

package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

func openConfigEditor(path string) error {
	if os.Getenv("EDITOR") == "" {
		fmt.Printf("EDITOR is not set. Please edit the file manually:\n  %s\n", path)
		return nil
	}
	// sudoedit lets a non-root user safely edit a root-owned file.
	c := exec.Command("sudo", "-e", path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
