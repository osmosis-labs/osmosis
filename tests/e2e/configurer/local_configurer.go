package configurer

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	dockerImages "github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/docker"
)

type LocalConfigurer struct {
	BaseConfigurer
}

var _ Configurer = (*LocalConfigurer)(nil)

func NewLocalConfigurer(t *testing.T, chainConfigs []*ChainConfig, setupTests setupFn, dockerImages *dockerImages.ImageConfig, dockerPool *dockertest.Pool, dockerNetwork *dockertest.Network) Configurer {
	return &LocalConfigurer{
		BaseConfigurer: BaseConfigurer{
			chainConfigs:  chainConfigs,
			dockerImages:  dockerImages,
			dockerPool:    dockerPool,
			dockerNetwork: dockerNetwork,
			setupTests:    setupTests,
			valResources:  make(map[string][]*dockertest.Resource),
			t:             t,
		},
	}
}

func (lc *LocalConfigurer) ConfigureChains() error {
	for _, chainConfig := range lc.chainConfigs {
		if err := lc.ConfigureChain(chainConfig); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LocalConfigurer) ConfigureChain(chainConfig *ChainConfig) error {
	lc.t.Logf("starting local e2e infrastructure for chain-id: %s", chainConfig.chainId)
	tmpDir, err := ioutil.TempDir("", "osmosis-e2e-testnet-")
	if err != nil {
		return err
	}
	lc.t.Logf("temp directory for chain-id %v: %v", chainConfig.chainId, tmpDir)
	initializedChain, err := chain.Init(chainConfig.chainId, tmpDir, chainConfig.validatorConfig, time.Duration(chainConfig.votingPeriod))
	if err != nil {
		return err
	}
	chainConfig.chain = initializedChain
	return nil
}

func (lc *LocalConfigurer) RunSetup() error {
	return lc.setupTests(lc)
}
