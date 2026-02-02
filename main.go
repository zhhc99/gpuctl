package main

import (
	"os"

	"github.com/zhhc99/gpuctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
