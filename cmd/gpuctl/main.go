package main

import (
	"fmt"
	"os"

	"github.com/zhhc99/gpuctl/internal/cli"
	"github.com/zhhc99/gpuctl/internal/locale"
)

func main() {
	locale.Init()
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
