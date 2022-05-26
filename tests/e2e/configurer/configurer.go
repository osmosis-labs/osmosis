package configurer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	dockerImages "github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/docker"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
)

type Configurer interface {
	ConfigureChains() error

	ClearResources() error

	GetChainConfig(chainIndex int) ChainConfig

	RunSetup() error

	RunValidators() error

	RunIBC() error

	SendIBC(srcChain *chain.Chain, dstChain *chain.Chain, recipient string, token sdk.Coin)

	CreatePool(chainId string, valIdx int, poolFile string)
}

type BaseConfigurer struct {
	chainConfigs   []*ChainConfig
	dockerImages   *dockerImages.ImageConfig
	dockerPool     *dockertest.Pool
	dockerNetwork  *dockertest.Network
	valResources   map[string][]*dockertest.Resource
	hermesResource *dockertest.Resource
	setupTests     setupFn
	t              *testing.T
}

type status struct {
	LatestHeight string `json:"latest_block_height"`
}

type syncInfo struct {
	SyncInfo status `json:"SyncInfo"`
}

const (
	// osmosis version being upgraded to (folder must exist here https://github.com/osmosis-labs/osmosis/tree/main/app/upgrades)
	UpgradeVersion = "v9"
	// estimated number of blocks it takes to submit for a proposal
	PropSubmitBlocks float32 = 10
	// estimated number of blocks it takes to deposit for a proposal
	PropDepositBlocks float32 = 10
	// number of blocks it takes to vote for a single validator to vote for a proposal
	PropVoteBlocks float32 = 1.2
	// number of blocks used as a calculation buffer
	PropBufferBlocks float32 = 5
	// max retries for json unmarshalling
	MaxRetries = 60
)

var (
	// whatever number of validator configs get posted here are how many validators that will spawn on chain A and B respectively
	validatorConfigsChainA = []*chain.ValidatorConfig{
		{
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "custom",
			PruningKeepRecent:  "10000",
			PruningInterval:    "13",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "everything",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   0,
			SnapshotKeepRecent: 0,
		},
	}
	validatorConfigsChainB = []*chain.ValidatorConfig{
		{
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "custom",
			PruningKeepRecent:  "10000",
			PruningInterval:    "13",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
	}
)

func New(t *testing.T, isIBCEnabled, isUpgradeEnabled bool) (Configurer, error) {
	dockerImages := dockerImages.NewImageConfig(isUpgradeEnabled)
	dkrPool, err := dockertest.NewPool("")
	if err != nil {
		return nil, err
	}
	dockerNetwork, err := dkrPool.CreateNetwork("osmosis-testnet")
	if err != nil {
		return nil, err
	}

	if isIBCEnabled && isUpgradeEnabled {
		// skip none - configure two chains via Docker
		// to utilize the older version of osmosis to upgrade from
		return NewUpgradeConfigurer(t,
			[]*ChainConfig{
				{
					chainId:         chain.ChainAID,
					validatorConfig: validatorConfigsChainA,
					skipRunValidatorIndexes: map[int]struct{}{
						3: {}, // skip validator at index 3
					},
				},
				{
					chainId:         chain.ChainBID,
					validatorConfig: validatorConfigsChainB,
				},
			},
			withUpgrade(withIBC(baseSetup)), // base set up with IBC and upgrade
			dockerImages,
			dkrPool,
			dockerNetwork,
		), nil
	} else if isIBCEnabled {
		// configure two chains locally
		return NewLocalConfigurer(t,
			[]*ChainConfig{
				{
					chainId:         chain.ChainAID,
					validatorConfig: validatorConfigsChainA,
					skipRunValidatorIndexes: map[int]struct{}{
						3: {}, // skip validator at index 3
					},
				},
				{
					chainId:         chain.ChainBID,
					validatorConfig: validatorConfigsChainB,
				},
			},
			withIBC(baseSetup), // base set up with IBC
			dockerImages,
			dkrPool,
			dockerNetwork,
		), nil
	} else if isUpgradeEnabled {
		// invalid - IBC tests must be enabled for upgrade
		// to function
		return nil, errors.New("IBC tests must be enabled for upgrade to work")
	} else {
		// configure one chain locally
		return NewLocalConfigurer(t,
			[]*ChainConfig{
				{
					chainId:         chain.ChainAID,
					validatorConfig: validatorConfigsChainA,
				},
			},
			baseSetup, // base set up only
			dockerImages,
			dkrPool,
			dockerNetwork,
		), nil
	}
}

func (bc *BaseConfigurer) GetChainConfig(chainIndex int) ChainConfig {
	return *bc.chainConfigs[chainIndex]
}

func (bc *BaseConfigurer) RunValidators() error {
	for i, chainConfig := range bc.chainConfigs {
		if err := bc.runValidators(chainConfig, bc.dockerImages.OsmosisRepository, bc.dockerImages.OsmosisTag, i*10); err != nil {
			return err
		}
	}
	return nil
}

func (bc *BaseConfigurer) runValidators(chainConfig *ChainConfig, dockerRepository, dockerTag string, portOffset int) error {
	chain := chainConfig.chain
	bc.t.Logf("starting %s validator containers...", chain.ChainMeta.Id)
	bc.valResources[chain.ChainMeta.Id] = make([]*dockertest.Resource, len(chain.Validators)-len(chainConfig.skipRunValidatorIndexes))
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
			NetworkID: bc.dockerNetwork.Network.ID,
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

		resource, err := bc.dockerPool.RunWithOptions(runOpts, noRestart)
		if err != nil {
			return err
		}

		bc.valResources[chain.ChainMeta.Id][i] = resource
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

func (bc *BaseConfigurer) RunIBC() error {
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

func (bc *BaseConfigurer) runIBCRelayer(chainA *chain.Chain, chainB *chain.Chain) error {
	bc.t.Log("starting Hermes relayer container...")

	tmpDir, err := ioutil.TempDir("", "osmosis-e2e-testnet-hermes-")
	if err != nil {
		return err
	}

	osmoAVal := chainA.Validators[0]
	osmoBVal := chainB.Validators[0]
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

	bc.hermesResource, err = bc.dockerPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s-%s-relayer", chainA.ChainMeta.Id, chainB.ChainMeta.Id),
			Repository: bc.dockerImages.RelayerRepository,
			Tag:        bc.dockerImages.RelayerTag,
			NetworkID:  bc.dockerNetwork.Network.ID,
			Cmd: []string{
				"start",
			},
			User: "root:root",
			Mounts: []string{
				fmt.Sprintf("%s/:/root/hermes", hermesCfgPath),
			},
			ExposedPorts: []string{
				"3031",
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"3031/tcp": {{HostIP: "", HostPort: "3031"}},
			},
			Env: []string{
				fmt.Sprintf("OSMO_A_E2E_CHAIN_ID=%s", chainA.ChainMeta.Id),
				fmt.Sprintf("OSMO_B_E2E_CHAIN_ID=%s", chainB.ChainMeta.Id),
				fmt.Sprintf("OSMO_A_E2E_VAL_MNEMONIC=%s", osmoAVal.Mnemonic),
				fmt.Sprintf("OSMO_B_E2E_VAL_MNEMONIC=%s", osmoBVal.Mnemonic),
				fmt.Sprintf("OSMO_A_E2E_VAL_HOST=%s", bc.valResources[chainA.ChainMeta.Id][0].Container.Name[1:]),
				fmt.Sprintf("OSMO_B_E2E_VAL_HOST=%s", bc.valResources[chainB.ChainMeta.Id][0].Container.Name[1:]),
			},
			Entrypoint: []string{
				"sh",
				"-c",
				"chmod +x /root/hermes/hermes_bootstrap.sh && /root/hermes/hermes_bootstrap.sh",
			},
		},
		noRestart,
	)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("http://%s/state", bc.hermesResource.GetHostPort("3031/tcp"))

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

	bc.t.Logf("started Hermes relayer container: %s", bc.hermesResource.Container.ID)

	// XXX: Give time to both networks to start, otherwise we might see gRPC
	// transport errors.
	time.Sleep(10 * time.Second)

	// create the client, connection and channel between the two Osmosis chains
	return bc.connectIBCChains(chainA, chainB)
}

func (bc *BaseConfigurer) connectIBCChains(chainA *chain.Chain, chainB *chain.Chain) error {
	bc.t.Logf("connecting %s and %s chains via IBC", chainA.ChainMeta.Id, chainB.ChainMeta.Id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := bc.dockerPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    bc.hermesResource.Container.ID,
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

	err = bc.dockerPool.Client.StartExec(exec.ID, docker.StartExecOptions{
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

func (bc *BaseConfigurer) ClearResources() error {
	bc.t.Log("tearing down e2e integration test suite...")

	require.NoError(bc.t, bc.dockerPool.Purge(bc.hermesResource))

	for _, vr := range bc.valResources {
		for _, r := range vr {
			require.NoError(bc.t, bc.dockerPool.Purge(r))
		}
	}

	require.NoError(bc.t, bc.dockerPool.RemoveNetwork(bc.dockerNetwork))

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
