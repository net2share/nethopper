package main

import (
	"os"

	"github.com/nethopper/nethopper/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
