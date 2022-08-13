package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	osmosis "github.com/osmosis-labs/osmosis/v11/app"
	"github.com/osmosis-labs/osmosis/v11/app/params"
	"github.com/osmosis-labs/osmosis/v11/cmd/osmosisd/cmd"
)

func main() {
	params.SetAddressPrefixes()
	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, osmosis.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
