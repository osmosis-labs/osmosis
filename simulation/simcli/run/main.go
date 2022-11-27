package main

import (
	"cosmossdk.io/client/v2/autocli"


	simcli "github.com/osmosis-labs/osmosis/v13/simulation/simcli"
)

func main() {
	err := autocli.RunFromAppConfig(simcli.AppConfig)
	if err != nil {
		panic(err)
	}
}
