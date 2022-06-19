package chain

import (
	"testing"

	chaininit "github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/containers"
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

	t                *testing.T
	containerManager *containers.Manager
}

type status struct {
	LatestHeight string `json:"latest_block_height"`
}

type syncInfo struct {
	SyncInfo status `json:"SyncInfo"`
}

func New(t *testing.T, containerManager *containers.Manager, id string, initValidatorConfigs []*chaininit.ValidatorConfig) *Config {
	return &Config{
		ChainMeta: chaininit.ChainMeta{
			Id: id,
		},
		ValidatorInitConfigs: initValidatorConfigs,
		t:                    t,
		containerManager:     containerManager,
	}
}
