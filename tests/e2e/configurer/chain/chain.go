package chain

import (
	chaininit "github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

type Config struct {
	chaininit.ChainMeta

	ValidatorInitConfigs []*chaininit.ValidatorConfig
	// voting period is number of blocks it takes to deposit, 1.2 seconds per validator to vote on the prop, and a buffer.
	VotingPeriod float32
	// upgrade proposal height for chain.
	PropHeight           int
	LatestProposalNumber int
	LatestLockNumber     int
	ValidatorConfigs     []*ValidatorConfig
}

func New(id string, initValidatorConfigs []*chaininit.ValidatorConfig) *Config {
	return &Config{
		ChainMeta: chaininit.ChainMeta{
			Id: id,
		},
		ValidatorInitConfigs: initValidatorConfigs,
	}
}
