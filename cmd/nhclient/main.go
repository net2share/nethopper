package main

import (
	"fmt"
	"os"

	"github.com/net2share/nethopper/internal/cmd"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	root := cmd.NewClientRoot(version, buildTime)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
