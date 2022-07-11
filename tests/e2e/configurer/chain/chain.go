package chain

import (
	"testing"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/containers"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/initialization"
)

type Config struct {
	initialization.ChainMeta

	ValidatorInitConfigs []*initialization.NodeConfig
	// voting period is number of blocks it takes to deposit, 1.2 seconds per validator to vote on the prop, and a buffer.
	VotingPeriod float32
	// upgrade proposal height for chain.
	PropHeight           int
	LatestProposalNumber int
	LatestLockNumber     int
	NodeConfigs          []*ValidatorConfig

	t                *testing.T
	containerManager *containers.Manager
}

type status struct {
	LatestHeight string `json:"latest_block_height"`
}

type syncInfo struct {
	SyncInfo status `json:"SyncInfo"`
}

func New(t *testing.T, containerManager *containers.Manager, id string, initValidatorConfigs []*initialization.NodeConfig) *Config {
	return &Config{
		ChainMeta: initialization.ChainMeta{
			Id: id,
		},
		ValidatorInitConfigs: initValidatorConfigs,
		t:                    t,
		containerManager:     containerManager,
	}
}
