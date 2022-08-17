package configurer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v11/tests/e2e/configurer/chain"
	"github.com/osmosis-labs/osmosis/v11/tests/e2e/containers"
	"github.com/osmosis-labs/osmosis/v11/tests/e2e/initialization"
	"github.com/osmosis-labs/osmosis/v11/tests/e2e/util"
)

// baseConfigurer is the base implementation for the
// other 2 types of configurers. It is not meant to be used
// on its own. Instead, it is meant to be embedded
// by composition into more concrete configurers.
type baseConfigurer struct {
	chainConfigs     []*chain.Config
	containerManager *containers.Manager
	setupTests       setupFn
	syncUntilHeight  int64 // the height until which to wait for validators to sync when first started.
	t                *testing.T
}

// defaultSyncUntilHeight arbitrary small height to make sure the chain is making progress.
const defaultSyncUntilHeight = 3

func (bc *baseConfigurer) ClearResources() error {
	bc.t.Log("tearing down e2e integration test suite...")

	if err := bc.containerManager.ClearResources(); err != nil {
		return err
	}

	for _, chainConfig := range bc.chainConfigs {
		os.RemoveAll(chainConfig.DataDir)
	}
	return nil
}

func (bc *baseConfigurer) GetChainConfig(chainIndex int) *chain.Config {
	return bc.chainConfigs[chainIndex]
}

func (bc *baseConfigurer) RunValidators() error {
	for _, chainConfig := range bc.chainConfigs {
		if err := bc.runValidators(chainConfig); err != nil {
			return err
		}
	}
	return nil
}

func (bc *baseConfigurer) runValidators(chainConfig *chain.Config) error {
	bc.t.Logf("starting %s validator containers...", chainConfig.Id)
	for _, node := range chainConfig.NodeConfigs {
		if err := node.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (bc *baseConfigurer) RunIBC() error {
	// Run a relayer between every possible pair of chains.
	for i := 0; i < len(bc.chainConfigs); i++ {
		for j := i + 1; j < len(bc.chainConfigs); j++ {
			if err := bc.runIBCRelayer(bc.chainConfigs[i], bc.chainConfigs[j]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (bc *baseConfigurer) runIBCRelayer(chainConfigA *chain.Config, chainConfigB *chain.Config) error {
	bc.t.Log("starting Hermes relayer container...")

	tmpDir, err := os.MkdirTemp("", "osmosis-e2e-testnet-hermes-")
	if err != nil {
		return err
	}

	hermesCfgPath := path.Join(tmpDir, "hermes")

	if err := os.MkdirAll(hermesCfgPath, 0o755); err != nil {
		return err
	}

	_, err = util.CopyFile(
		filepath.Join("./scripts/", "hermes_bootstrap.sh"),
		filepath.Join(hermesCfgPath, "hermes_bootstrap.sh"),
	)
	if err != nil {
		return err
	}

	relayerNodeA := chainConfigA.NodeConfigs[0]
	relayerNodeB := chainConfigB.NodeConfigs[0]

	hermesResource, err := bc.containerManager.RunHermesResource(
		chainConfigA.Id,
		relayerNodeA.Name,
		relayerNodeA.Mnemonic,
		chainConfigB.Id,
		relayerNodeB.Name,
		relayerNodeB.Mnemonic,
		hermesCfgPath)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("http://%s/state", hermesResource.GetHostPort("3031/tcp"))

	require.Eventually(bc.t, func() bool {
		resp, err := http.Get(endpoint)
		if err != nil {
			return false
		}

		defer resp.Body.Close()

		bz, err := io.ReadAll(resp.Body)
		if err != nil {
			return false
		}

		var respBody map[string]interface{}
		if err := json.Unmarshal(bz, &respBody); err != nil {
			return false
		}

		status, ok := respBody["status"].(string)
		require.True(bc.t, ok)
		result, ok := respBody["result"].(map[string]interface{})
		require.True(bc.t, ok)

		chains, ok := result["chains"].([]interface{})
		require.True(bc.t, ok)

		return status == "success" && len(chains) == 2
	},
		5*time.Minute,
		time.Second,
		"hermes relayer not healthy")

	bc.t.Logf("started Hermes relayer container: %s", hermesResource.Container.ID)

	// XXX: Give time to both networks to start, otherwise we might see gRPC
	// transport errors.
	time.Sleep(10 * time.Second)

	// create the client, connection and channel between the two Osmosis chains
	return bc.connectIBCChains(chainConfigA, chainConfigB)
}

func (bc *baseConfigurer) connectIBCChains(chainA *chain.Config, chainB *chain.Config) error {
	bc.t.Logf("connecting %s and %s chains via IBC", chainA.ChainMeta.Id, chainB.ChainMeta.Id)
	cmd := []string{"hermes", "create", "channel", chainA.ChainMeta.Id, chainB.ChainMeta.Id, "--port-a=transfer", "--port-b=transfer"}
	_, _, err := bc.containerManager.ExecHermesCmd(bc.t, cmd, "successfully opened init channel")
	if err != nil {
		return err
	}
	bc.t.Logf("connected %s and %s chains via IBC", chainA.ChainMeta.Id, chainB.ChainMeta.Id)
	return nil
}

func (bc *baseConfigurer) initializeChainConfigFromInitChain(initializedChain *initialization.Chain, chainConfig *chain.Config) {
	chainConfig.ChainMeta = initializedChain.ChainMeta
	chainConfig.NodeConfigs = make([]*chain.NodeConfig, 0, len(initializedChain.Nodes))
	setupTime := time.Now()
	for i, validator := range initializedChain.Nodes {
		conf := chain.NewNodeConfig(bc.t, validator, chainConfig.ValidatorInitConfigs[i], chainConfig.Id, bc.containerManager).WithSetupTime(setupTime)
		chainConfig.NodeConfigs = append(chainConfig.NodeConfigs, conf)
	}
}
