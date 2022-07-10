package chain

import "github.com/osmosis-labs/osmosis/v7/tests/e2e/initialization"

type ValidatorConfig struct {
	initialization.Node

	OperatorAddress string
}
