//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

func openConfigEditor(path string) error {
	if !isElevated() {
		fmt.Fprintf(os.Stderr, "Editing the config requires administrator privileges.\n")
		fmt.Fprintf(os.Stderr, "Please open the following file manually as administrator:\n  %s\n", path)
		return nil
	}
	return exec.Command("notepad", path).Run()
}
