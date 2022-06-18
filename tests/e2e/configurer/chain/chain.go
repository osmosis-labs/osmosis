package chain

import "github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"

type Config struct {
	ChainId         string
	ValidatorConfig []*chain.ValidatorConfig
	// voting period is number of blocks it takes to deposit, 1.2 seconds per validator to vote on the prop, and a buffer.
	VotingPeriod float32
	// upgrade proposal height for chain.
	PropHeight int
	Chain      *chain.Chain
}

func (cc Config) GetChain() *chain.Chain {
	return cc.Chain
}
