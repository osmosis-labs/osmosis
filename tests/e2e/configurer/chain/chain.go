package chain

import "github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"

type Config struct {
	ChainId         string
	ValidatorConfig []*chain.ValidatorConfig
	// voting period is number of blocks it takes to deposit, 1.2 seconds per validator to vote on the prop, and a buffer.
	VotingPeriod float32
	// upgrade proposal height for chain.
	PropHeight int
	// Indexes of the validators to skip from running during initialization.
	// This is needed for testing functionality like state-sync where we would
	// like to start a node during tests post-initialization.
	SkipRunValidatorIndexes map[int]struct{}
	Chain                   *chain.Chain
}

func (cc Config) GetChain() *chain.Chain {
	return cc.Chain
}
