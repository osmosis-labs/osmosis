package main

import (
	"os"
	"runtime"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	osmosis "github.com/osmosis-labs/osmosis/v25/app"
	"github.com/osmosis-labs/osmosis/v25/app/params"
	"github.com/osmosis-labs/osmosis/v25/cmd/osmosisd/cmd"
)

func main() {
	runtime.SetBlockProfileRate(1)

	params.SetAddressPrefixes()
	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "OSMOSISD", osmosis.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
