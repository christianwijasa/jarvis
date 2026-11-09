package main

import (
	"os"

	"github.com/christianwijasabfi/atlas/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
