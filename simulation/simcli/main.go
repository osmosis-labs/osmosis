package main

import (
	"cosmossdk.io/client/v2/autocli"

	simulation "github.com/osmosis-labs/osmosis/v13/simulation"
)

func main() {
	err := autocli.RunFromAppConfig(simulation.AppConfig)
	if err != nil {
		panic(err)
	}
}
