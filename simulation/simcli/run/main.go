package main

import (
	"cosmossdk.io/client/v2/autocli"

	_ "github.com/cosmos/cosmos-sdk/runtime"

	simcli "github.com/osmosis-labs/osmosis/v13/simulation/simcli"
	_ "github.com/osmosis-labs/osmosis/v13/x/lockup"
)

func main() {
	err := autocli.RunFromAppConfig(simcli.AppConfig)
	if err != nil {
		panic(err)
	}
}
