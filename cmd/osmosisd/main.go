package main

import (
	"os"

	"github.com/c-osmosis/osmosis/cmd/osmosisd/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()
	if err := cmd.Execute(rootCmd); err != nil {
		os.Exit(1)
	}
}
