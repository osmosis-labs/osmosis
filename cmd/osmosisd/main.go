package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	osmosis "github.com/osmosis-labs/osmosis/v9/app"
	"github.com/osmosis-labs/osmosis/v9/app/params"
	"github.com/osmosis-labs/osmosis/v9/cmd/osmosisd/cmd"
)

func main() {
	params.SetAddressPrefixes()
	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, osmosis.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
