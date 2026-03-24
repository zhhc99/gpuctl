//go:build linux

package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/zhhc99/gpuctl/internal/locale"
)

func checkPrivileged() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("%s", locale.T("err.need_root"))
	}
	return nil
}

func openConfigEditor(path string) error {
	if os.Geteuid() == 0 {
		// Already root (e.g. user ran sudo gpuctl conf edit).
		// sudoedit requires a non-root caller to locate the user's $EDITOR.
		fmt.Fprintf(os.Stderr, locale.T("msg.edit_as_root")+"\n", path)
		return nil
	}
	c := exec.Command("sudo", "-e", path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
