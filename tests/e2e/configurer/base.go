package configurer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	chaininit "github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/containers"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
)

// baseConfigurer is the base implementation for the
// other 2 types of configurers. It is not meant to be used
// on its own. Instead, it is meant to be embedded
// by composition into more concrete configurers.
type baseConfigurer struct {
	chainConfigs     []*chain.Config
	containerManager *containers.Manager
	setupTests       setupFn
	t                *testing.T
}

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
	for i, chainConfig := range bc.chainConfigs {
		if err := bc.runValidators(chainConfig, i*10); err != nil {
			return err
		}
	}
	return nil
}

func (bc *baseConfigurer) runValidators(chainConfig *chain.Config, portOffset int) error {
	bc.t.Logf("starting %s validator containers...", chainConfig.Id)

	for _, val := range chainConfig.ValidatorConfigs {
		resource, err := bc.containerManager.RunValidatorResource(chainConfig.Id, val.Name, val.ConfigDir)
		if err != nil {
			return err
		}
		bc.t.Logf("started %s validator container: %s", resource.Container.Name[1:], resource.Container.ID)
	}

	validatorHostPort, err := bc.containerManager.GetValidatorHostPort(chainConfig.Id, 0, "26657/tcp")
	if err != nil {
		return err
	}

	rpcClient, err := rpchttp.New(fmt.Sprintf("tcp://%s", validatorHostPort), "/websocket")
	if err != nil {
		return err
	}

	require.Eventually(
		bc.t,
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			status, err := rpcClient.Status(ctx)
			if err != nil {
				return false
			}

			// let the node produce a few blocks
			if status.SyncInfo.CatchingUp || status.SyncInfo.LatestBlockHeight < 3 {
				return false
			}

			return true
		},
		5*time.Minute,
		time.Second,
		"Osmosis node failed to produce blocks",
	)
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

	tmpDir, err := ioutil.TempDir("", "osmosis-e2e-testnet-hermes-")
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

	hermesResource, err := bc.containerManager.RunHermesResource(
		chainConfigA.Id,
		chainConfigA.ValidatorConfigs[0].Mnemonic,
		chainConfigB.Id, chainConfigB.ValidatorConfigs[0].Mnemonic,
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

	bc.t.Logf("started Hermes relayer container: %s", bc.containerManager.GetHermesContainerID())

	// XXX: Give time to both networks to start, otherwise we might see gRPC
	// transport errors.
	time.Sleep(10 * time.Second)

	// create the client, connection and channel between the two Osmosis chains
	return bc.connectIBCChains(chainConfigA, chainConfigB)
}

func (bc *baseConfigurer) connectIBCChains(chainA *chain.Config, chainB *chain.Config) error {
	bc.t.Logf("connecting %s and %s chains via IBC", chainA.ChainMeta.Id, chainB.ChainMeta.Id)
	cmd := []string{"hermes", "create", "channel", chainA.ChainMeta.Id, chainB.ChainMeta.Id, "--port-a=transfer", "--port-b=transfer"}
	_, _, err := bc.containerManager.ExecCmd(bc.t, "", 0, cmd, "successfully opened init channel")
	if err != nil {
		return err
	}
	bc.t.Logf("connected %s and %s chains via IBC", chainA.ChainMeta.Id, chainB.ChainMeta.Id)
	return nil
}

func (bc *baseConfigurer) initializeChainConfigFromInitChain(initializedChain *chaininit.Chain, chainConfig *chain.Config) {
	chainConfig.ChainMeta = initializedChain.ChainMeta
	chainConfig.ValidatorConfigs = make([]*chain.ValidatorConfig, 0, len(initializedChain.Validators))
	for _, validator := range initializedChain.Validators {
		chainConfig.ValidatorConfigs = append(chainConfig.ValidatorConfigs, &chain.ValidatorConfig{
			Validator: *validator,
		})
	}
}
