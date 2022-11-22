package main

import (
	autocli "github.com/osmosis-labs/osmosis/v13/client/v2/autocli"
	simcli "github.com/osmosis-labs/osmosis/v13/simulation/simcli"
)

func main() {
	err := autocli.RunFromAppConfig(simcli.AppConfig)
	if err != nil {
		panic(err)
	}
}
