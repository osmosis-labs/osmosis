package main

import (
	"os"

	"github.com/osmosis-labs/osmosis/v4/app/params"
	"github.com/osmosis-labs/osmosis/v4/cmd/osmosisd/cmd"
)

func main() {
	params.SetAddressPrefixes()
	rootCmd, _ := cmd.NewRootCmd()
	if err := cmd.Execute(rootCmd); err != nil {
		os.Exit(1)
	}
}
