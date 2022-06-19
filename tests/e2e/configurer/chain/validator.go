package chain

import chaininit "github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"

type ValidatorConfig struct {
	chaininit.Validator

	OperatorAddress string
}
