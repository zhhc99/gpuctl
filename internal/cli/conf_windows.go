//go:build windows

package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/zhhc99/gpuctl/internal/locale"
)

func checkPrivileged() error {
	if !isElevated() {
		return fmt.Errorf("%s", locale.T("err.need_admin"))
	}
	return nil
}

func isElevated() bool {
	f, err := os.Open(`\\.\PHYSICALDRIVE0`)
	if err == nil {
		f.Close()
		return true
	}
	return false
}

func openConfigEditor(path string) error {
	if !isElevated() {
		fmt.Fprintf(os.Stderr, locale.T("msg.need_admin_edit")+"\n", path)
		return nil
	}
	return exec.Command("notepad", path).Run()
}
