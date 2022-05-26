package configurer

import "github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"

type ChainConfig struct {
	chainId         string
	validatorConfig []*chain.ValidatorConfig
	// voting period is number of blocks it takes to deposit, 1.2 seconds per validator to vote on the prop, and a buffer.
	votingPeriod float32
	// upgrade proposal height for chain.
	propHeight int
	// Indexes of the validators to skip from running during initialization.
	// This is needed for testing functionality like state-sync where we would
	// like to start a node during tests post-initialization.
	skipRunValidatorIndexes map[int]struct{}
	chain                   *chain.Chain
}

func (cc ChainConfig) GetChain() *chain.Chain {
	return cc.chain
}
