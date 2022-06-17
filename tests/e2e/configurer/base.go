package configurer

import (
	"bytes"
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

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/containers"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
)

// baseConfigurer is the base implementation for the
// other 2 types of configurers. It is not meant to be used
// on its own. Instead, it is meant to be embedded
// by composition into more concrete configurers.
type baseConfigurer struct {
	chainConfigs     []*ChainConfig
	containerManager *containers.Manager
	setupTests       setupFn
	t                *testing.T
}

func (bc *baseConfigurer) GetChainConfig(chainIndex int) ChainConfig {
	return *bc.chainConfigs[chainIndex]
}

func (bc *baseConfigurer) RunValidators() error {
	for i, chainConfig := range bc.chainConfigs {
		if err := bc.runValidators(chainConfig, bc.containerManager.OsmosisRepository, bc.containerManager.OsmosisTag, i*10); err != nil {
			return err
		}
	}
	return nil
}

func (bc *baseConfigurer) runValidators(chainConfig *ChainConfig, dockerRepository, dockerTag string, portOffset int) error {
	chain := chainConfig.chain
	bc.t.Logf("starting %s validator containers...", chain.ChainMeta.Id)
	bc.containerManager.ValResources[chain.ChainMeta.Id] = make([]*dockertest.Resource, len(chain.Validators)-len(chainConfig.skipRunValidatorIndexes))
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	for i, val := range chain.Validators {
		// Skip some validators from running during set up.
		// This is needed for testing functionality like
		// state-sunc where we might want to start some validators during tests.
		if _, ok := chainConfig.skipRunValidatorIndexes[i]; ok {
			bc.t.Logf("skipping %s validator with index %d from running...", val.Name, i)
			continue
		}

		runOpts := &dockertest.RunOptions{
			Name:      val.Name,
			NetworkID: bc.containerManager.Network.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/osmosis/.osmosisd", val.ConfigDir),
				fmt.Sprintf("%s/scripts:/osmosis", pwd),
			},
			Repository: dockerRepository,
			Tag:        dockerTag,
			Cmd: []string{
				"start",
			},
		}

		// expose the first validator for debugging and communication
		if val.Index == 0 {
			runOpts.PortBindings = map[docker.Port][]docker.PortBinding{
				"1317/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 1317+portOffset)}},
				"6060/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6060+portOffset)}},
				"6061/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6061+portOffset)}},
				"6062/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6062+portOffset)}},
				"6063/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6063+portOffset)}},
				"6064/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6064+portOffset)}},
				"6065/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6065+portOffset)}},
				"9090/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 9090+portOffset)}},
				"26656/tcp": {{HostIP: "", HostPort: fmt.Sprintf("%d", 26656+portOffset)}},
				"26657/tcp": {{HostIP: "", HostPort: fmt.Sprintf("%d", 26657+portOffset)}},
			}
		}

		resource, err := bc.containerManager.Pool.RunWithOptions(runOpts, noRestart)
		if err != nil {
			return err
		}

		bc.containerManager.ValResources[chain.ChainMeta.Id][i] = resource
		bc.t.Logf("started %s validator container: %s", resource.Container.Name[1:], resource.Container.ID)
	}

	rpcClient, err := rpchttp.New("tcp://localhost:26657", "/websocket")
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
			if err := bc.runIBCRelayer(bc.chainConfigs[i].chain, bc.chainConfigs[j].chain); err != nil {
				return err
			}
		}
	}
	return nil
}

func (bc *baseConfigurer) runIBCRelayer(chainA *chain.Chain, chainB *chain.Chain) error {
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

	osmoAValMnemonic := chainA.Validators[0].Mnemonic
	osmoBValMnemonic := chainB.Validators[0].Mnemonic

	hermesResource, err := bc.containerManager.RunHermesResource(chainA.ChainMeta.Id, osmoAValMnemonic, chainB.ChainMeta.Id, osmoBValMnemonic, hermesCfgPath)
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
	return bc.connectIBCChains(chainA, chainB)
}

func (bc *baseConfigurer) connectIBCChains(chainA *chain.Chain, chainB *chain.Chain) error {
	bc.t.Logf("connecting %s and %s chains via IBC", chainA.ChainMeta.Id, chainB.ChainMeta.Id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := bc.containerManager.Pool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    bc.containerManager.GetHermesContainerID(),
		User:         "root",
		Cmd: []string{
			"hermes",
			"create",
			"channel",
			chainA.ChainMeta.Id,
			chainB.ChainMeta.Id,
			"--port-a=transfer",
			"--port-b=transfer",
		},
	})
	if err != nil {
		return err
	}

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = bc.containerManager.Pool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	if err != nil {
		bc.t.Logf("failed connect chains; stdout: %s, stderr: %s", outBuf.String(), errBuf.String())
		return err
	}

	if err != nil {
		bc.t.Logf("failed connect chains; stdout: %s, stderr: %s", outBuf.String(), errBuf.String())
		return err
	}

	require.Containsf(
		bc.t,
		errBuf.String(),
		"successfully opened init channel",
		"failed to connect chains via IBC: %s", errBuf.String(),
	)

	bc.t.Logf("connected %s and %s chains via IBC", chainA.ChainMeta.Id, chainB.ChainMeta.Id)
	return nil
}

func (bc *baseConfigurer) ClearResources() error {
	bc.t.Log("tearing down e2e integration test suite...")

	if err := bc.containerManager.ClearResources(); err != nil {
		return err
	}

	for _, chainConfig := range bc.chainConfigs {
		os.RemoveAll(chainConfig.chain.ChainMeta.DataDir)
	}
	return nil
}

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}
