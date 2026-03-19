package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:               "install",
	Short:             "Install gpuctl and register the startup service",
	PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
	PersistentPostRun: func(*cobra.Command, []string) {},
	RunE: func(cmd *cobra.Command, args []string) error {
		return doInstall()
	},
}

var uninstallCmd = &cobra.Command{
	Use:               "uninstall",
	Short:             "Remove gpuctl and unregister the startup service",
	PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
	PersistentPostRun: func(*cobra.Command, []string) {},
	RunE: func(cmd *cobra.Command, args []string) error {
		return doUninstall()
	},
}

func installBinary(dst string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	if exe == dst {
		fmt.Printf("Binary already at %s — skipping copy.\n", dst)
		return nil
	}

	if _, err := os.Stat(dst); err == nil {
		fmt.Printf("File already exists at %s.\n", dst)
		if !confirm("Overwrite? [Y/n] (default=Y): ", true) {
			return fmt.Errorf("aborted")
		}
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	src, err := os.Open(exe)
	if err != nil {
		return fmt.Errorf("open source binary: %w", err)
	}
	defer src.Close()

	tmp := dst + ".tmp"
	out, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create target file: %w", err)
	}

	if _, err := io.Copy(out, src); err != nil {
		out.Close()
		os.Remove(tmp)
		return fmt.Errorf("copy binary: %w", err)
	}
	out.Close()

	if err := os.Rename(tmp, dst); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("install binary: %w", err)
	}

	fmt.Printf("Binary installed to %s\n", dst)
	return nil
}

func confirm(prompt string, defaultYes bool) bool {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return defaultYes
	}
	switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return defaultYes
	}
}
