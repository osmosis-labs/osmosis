package configurer

import (
	"io/ioutil"
	"testing"
	"time"

	chaininit "github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/containers"
)

type CurrentBranchConfigurer struct {
	baseConfigurer
}

var _ Configurer = (*CurrentBranchConfigurer)(nil)

func NewCurrentBranchConfigurer(t *testing.T, chainConfigs []*chain.Config, setupTests setupFn, containerManager *containers.Manager) Configurer {
	return &CurrentBranchConfigurer{
		baseConfigurer: baseConfigurer{
			chainConfigs:     chainConfigs,
			containerManager: containerManager,
			setupTests:       setupTests,
			t:                t,
		},
	}
}

func (cb *CurrentBranchConfigurer) ConfigureChains() error {
	for _, chainConfig := range cb.chainConfigs {
		if err := cb.ConfigureChain(chainConfig); err != nil {
			return err
		}
	}
	return nil
}

func (cb *CurrentBranchConfigurer) ConfigureChain(chainConfig *chain.Config) error {
	cb.t.Logf("starting e2e infrastructure from current branch for chain-id: %s", chainConfig.Id)
	tmpDir, err := ioutil.TempDir("", "osmosis-e2e-testnet-")
	if err != nil {
		return err
	}
	cb.t.Logf("temp directory for chain-id %v: %v", chainConfig.Id, tmpDir)
	initializedChain, err := chaininit.Init(chainConfig.Id, tmpDir, chainConfig.ValidatorInitConfigs, time.Duration(chainConfig.VotingPeriod))
	if err != nil {
		return err
	}
	cb.initializeChainConfigFromInitChain(initializedChain, chainConfig)
	return nil
}

func (cb *CurrentBranchConfigurer) RunSetup() error {
	return cb.setupTests(cb)
}
