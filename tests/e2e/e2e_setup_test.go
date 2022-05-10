package e2e

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
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"

	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
)

var (
	// variable used to switch between chain A and B prop height in for loop
	propHeight int
	// upgrade proposal height for chain A
	propHeightA int
	// upgrade proposal height for chain B
	propHeightB int
	// max retries for json unmarshalling
	maxRetries = 60
	// whatever number of validator configs get posted here are how many validators that will spawn on chain A and B respectively
	validatorConfigsChainA = []*chain.ValidatorConfig{
		{
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{Pruning: "nothing", PruningKeepRecent: "0", PruningInterval: "0", SnapshotInterval: 1500, SnapshotKeepRecent: 2},
		{Pruning: "custom", PruningKeepRecent: "10000", PruningInterval: "13", SnapshotInterval: 1500, SnapshotKeepRecent: 2},
	}
	validatorConfigsChainB = []*chain.ValidatorConfig{
		{
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{Pruning: "nothing", PruningKeepRecent: "0", PruningInterval: "0", SnapshotInterval: 1500, SnapshotKeepRecent: 2},
		{Pruning: "custom", PruningKeepRecent: "10000", PruningInterval: "13", SnapshotInterval: 1500, SnapshotKeepRecent: 2},
	}
)

type IntegrationTestSuite struct {
	suite.Suite

	tmpDirs        []string
	chains         []*chain.Chain
	dkrPool        *dockertest.Pool
	dkrNet         *dockertest.Network
	hermesResource *dockertest.Resource
	initResource   *dockertest.Resource
	valResources   map[string][]*dockertest.Resource
}

type status struct {
	LatestHeight string `json:"latest_block_height"`
}

type syncInfo struct {
	SyncInfo status `json:"SyncInfo"`
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")

	s.chains = make([]*chain.Chain, 0, 2)

	// The e2e test flow is as follows:
	//
	// 1. Configure two chains - chan A and chain B.
	//   * For each chain, set up two validators
	//   * Initialize configs and genesis for all validators.
	// 2. Start both networks.
	// 3. Run IBC relayer betweeen the two chains.
	// 4. Execute various e2e tests, including IBC.
	s.configureDockerResources(chain.ChainAID, chain.ChainBID)
	s.configureChain(chain.ChainAID, validatorConfigsChainA)
	s.configureChain(chain.ChainBID, validatorConfigsChainB)

	s.runValidators(s.chains[0], 0)
	s.runValidators(s.chains[1], 10)
	s.runIBCRelayer()
	// pre upgrade state creation
	s.sendIBC(s.chains[0], s.chains[1], s.chains[1].Validators[0].PublicAddress, chain.OsmoToken)
	s.sendIBC(s.chains[1], s.chains[0], s.chains[0].Validators[0].PublicAddress, chain.OsmoToken)
	s.sendIBC(s.chains[0], s.chains[1], s.chains[1].Validators[0].PublicAddress, chain.StakeToken)
	s.sendIBC(s.chains[1], s.chains[0], s.chains[0].Validators[0].PublicAddress, chain.StakeToken)
	s.createPool(s.chains[0], "pool1A.json")
	s.createPool(s.chains[1], "pool1B.json")
	// initialize and run the upgrade
	s.upgrade()
	// post upgrade tests
	s.sendIBC(s.chains[0], s.chains[1], s.chains[1].Validators[0].PublicAddress, chain.OsmoToken)
	s.sendIBC(s.chains[1], s.chains[0], s.chains[0].Validators[0].PublicAddress, chain.OsmoToken)
	s.sendIBC(s.chains[0], s.chains[1], s.chains[1].Validators[0].PublicAddress, chain.StakeToken)
	s.sendIBC(s.chains[1], s.chains[0], s.chains[0].Validators[0].PublicAddress, chain.StakeToken)
	s.createPool(s.chains[0], "pool2A.json")
	s.createPool(s.chains[1], "pool2B.json")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if str := os.Getenv("OSMOSIS_E2E_SKIP_CLEANUP"); len(str) > 0 {
		skipCleanup, err := strconv.ParseBool(str)
		s.Require().NoError(err)

		if skipCleanup {
			return
		}
	}

	s.T().Log("tearing down e2e integration test suite...")

	s.Require().NoError(s.dkrPool.Purge(s.hermesResource))

	for _, vr := range s.valResources {
		for _, r := range vr {
			s.Require().NoError(s.dkrPool.Purge(r))
		}
	}

	s.Require().NoError(s.dkrPool.RemoveNetwork(s.dkrNet))

	for _, chain := range s.chains {
		os.RemoveAll(chain.ChainMeta.DataDir)
	}

	for _, td := range s.tmpDirs {
		os.RemoveAll(td)
	}
}

func (s *IntegrationTestSuite) runValidators(c *chain.Chain, portOffset int) {
	s.T().Logf("starting %s validator containers...", c.ChainMeta.Id)
	s.valResources[c.ChainMeta.Id] = make([]*dockertest.Resource, len(c.Validators))
	pwd, err := os.Getwd()
	s.Require().NoError(err)
	for i, val := range c.Validators {
		runOpts := &dockertest.RunOptions{
			Name:      val.Name,
			NetworkID: s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/osmosis/.osmosisd", val.ConfigDir),
				fmt.Sprintf("%s/scripts:/osmosis", pwd),
			},
			Repository: "osmolabs/osmosis-dev",
			Tag:        "v7.2.1-debug",
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

		resource, err := s.dkrPool.RunWithOptions(runOpts, noRestart)
		s.Require().NoError(err)

		s.valResources[c.ChainMeta.Id][i] = resource
		s.T().Logf("started %s validator container: %s", resource.Container.Name[1:], resource.Container.ID)
	}

	rpcClient, err := rpchttp.New("tcp://localhost:26657", "/websocket")
	s.Require().NoError(err)

	s.Require().Eventually(
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
}

func (s *IntegrationTestSuite) runIBCRelayer() {
	s.T().Log("starting Hermes relayer container...")

	tmpDir, err := ioutil.TempDir("", "osmosis-e2e-testnet-hermes-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	osmoAVal := s.chains[0].Validators[0]
	osmoBVal := s.chains[1].Validators[0]
	hermesCfgPath := path.Join(tmpDir, "hermes")

	s.Require().NoError(os.MkdirAll(hermesCfgPath, 0o755))
	_, err = util.CopyFile(
		filepath.Join("./scripts/", "hermes_bootstrap.sh"),
		filepath.Join(hermesCfgPath, "hermes_bootstrap.sh"),
	)
	s.Require().NoError(err)

	s.hermesResource, err = s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s-%s-relayer", s.chains[0].ChainMeta.Id, s.chains[1].ChainMeta.Id),
			Repository: "osmolabs/hermes",
			Tag:        "0.13.0",
			NetworkID:  s.dkrNet.Network.ID,
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
				fmt.Sprintf("OSMO_A_E2E_CHAIN_ID=%s", s.chains[0].ChainMeta.Id),
				fmt.Sprintf("OSMO_B_E2E_CHAIN_ID=%s", s.chains[1].ChainMeta.Id),
				fmt.Sprintf("OSMO_A_E2E_VAL_MNEMONIC=%s", osmoAVal.Mnemonic),
				fmt.Sprintf("OSMO_B_E2E_VAL_MNEMONIC=%s", osmoBVal.Mnemonic),
				fmt.Sprintf("OSMO_A_E2E_VAL_HOST=%s", s.valResources[s.chains[0].ChainMeta.Id][0].Container.Name[1:]),
				fmt.Sprintf("OSMO_B_E2E_VAL_HOST=%s", s.valResources[s.chains[1].ChainMeta.Id][0].Container.Name[1:]),
			},
			Entrypoint: []string{
				"sh",
				"-c",
				"chmod +x /root/hermes/hermes_bootstrap.sh && /root/hermes/hermes_bootstrap.sh",
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	endpoint := fmt.Sprintf("http://%s/state", s.hermesResource.GetHostPort("3031/tcp"))
	s.Require().Eventually(
		func() bool {
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

			status := respBody["status"].(string)
			result := respBody["result"].(map[string]interface{})

			return status == "success" && len(result["chains"].([]interface{})) == 2
		},
		5*time.Minute,
		time.Second,
		"hermes relayer not healthy",
	)

	s.T().Logf("started Hermes relayer container: %s", s.hermesResource.Container.ID)

	// XXX: Give time to both networks to start, otherwise we might see gRPC
	// transport errors.
	time.Sleep(10 * time.Second)

	// create the client, connection and channel between the two Osmosis chains
	s.connectIBCChains()
}

func (s *IntegrationTestSuite) configureChain(chainId string, validatorConfigs []*chain.ValidatorConfig) {

	s.T().Logf("starting e2e infrastructure for chain-id: %s", chainId)
	tmpDir, err := ioutil.TempDir("", "osmosis-e2e-testnet-")

	s.T().Logf("temp directory for chain-id %v: %v", chainId, tmpDir)
	s.Require().NoError(err)

	b, err := json.Marshal(validatorConfigs)
	s.Require().NoError(err)

	s.initResource, err = s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s", chainId),
			Repository: "osmolabs/osmosis-init",
			Tag:        "v7.3.0",
			NetworkID:  s.dkrNet.Network.ID,
			Cmd: []string{
				fmt.Sprintf("--data-dir=%s", tmpDir),
				fmt.Sprintf("--chain-id=%s", chainId),
				fmt.Sprintf("--config=%s", b),
			},
			User: "root:root",
			Mounts: []string{
				fmt.Sprintf("%s:%s", tmpDir, tmpDir),
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	var newChain chain.Chain

	fileName := fmt.Sprintf("%v/%v-encode", tmpDir, chainId)
	s.T().Logf("serialized init file for chain-id %v: %v", chainId, fileName)

	// loop through the reading and unmarshaling of the init file a total of maxRetries or until error is nil
	// without this, test attempts to unmarshal file before docker container is finished writing
	for i := 0; i < maxRetries; i++ {
		encJson, _ := os.ReadFile(fileName)
		err = json.Unmarshal(encJson, &newChain)
		if err == nil {
			break
		}

		if i == maxRetries-1 {
			s.Require().NoError(err)
		}

		if i > 0 {
			time.Sleep(1 * time.Second)
		}
	}
	s.chains = append(s.chains, &newChain)
	s.Require().NoError(s.dkrPool.Purge(s.initResource))
}

func (s *IntegrationTestSuite) configureDockerResources(chainIDOne, chainIDTwo string) {
	var err error
	s.dkrPool, err = dockertest.NewPool("")
	s.Require().NoError(err)

	s.dkrNet, err = s.dkrPool.CreateNetwork(fmt.Sprintf("%s-%s-testnet", chainIDOne, chainIDTwo))
	s.Require().NoError(err)

	s.valResources = make(map[string][]*dockertest.Resource)
}

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}

func (s *IntegrationTestSuite) upgrade() {
	// submit, deposit, and vote for upgrade proposal
	currentHeightA := s.chainHeight(s.valResources[s.chains[0].ChainMeta.Id][0].Container.ID)
	currentHeightB := s.chainHeight(s.valResources[s.chains[1].ChainMeta.Id][0].Container.ID)
	propHeightA = currentHeightA + 40 + len(s.chains[0].Validators)
	propHeightB = currentHeightB + 40 + len(s.chains[1].Validators)

	s.submitProposal(s.chains[0], propHeightA)
	s.submitProposal(s.chains[1], propHeightB)
	s.depositProposal(s.chains[0])
	s.depositProposal(s.chains[1])
	var wg sync.WaitGroup
	wg.Add(2)
	// use goroutines to vote on both chains concurrently (no possibility of sequence mismatch)
	go s.voteProposal(s.chains[0], &wg)
	go s.voteProposal(s.chains[1], &wg)
	wg.Wait()
	// wait till all chains halt at upgrade height
	for _, c := range s.chains {
		if c.ChainMeta.Id == chain.ChainAID {
			propHeight = propHeightA
		} else {
			propHeight = propHeightB
		}
		for i := range c.Validators {
			// use counter to ensure no new blocks are being created
			counter := 0
			s.T().Logf("waiting to reach upgrade height on %s validator container: %s", s.valResources[c.ChainMeta.Id][i].Container.Name[1:], s.valResources[c.ChainMeta.Id][i].Container.ID)
			s.Require().Eventually(
				func() bool {
					currentHeight := s.chainHeight(s.valResources[c.ChainMeta.Id][i].Container.ID)
					if currentHeight != propHeight {
						s.T().Logf("current block height on %s is %v, waiting for block %v container: %s", s.valResources[c.ChainMeta.Id][i].Container.Name[1:], currentHeight, propHeight, s.valResources[c.ChainMeta.Id][i].Container.ID)
					}
					if currentHeight > propHeight {
						panic("chain did not halt at upgrade height")
					}
					if currentHeight == propHeight {
						counter++
					}
					return counter == 3
				},
				5*time.Minute,
				time.Second,
			)
			s.T().Logf("reached upgrade height on %s container: %s", s.valResources[c.ChainMeta.Id][i].Container.Name[1:], s.valResources[c.ChainMeta.Id][i].Container.ID)
		}
	}

	// remove all containers so we can upgrade them to the new version
	for _, chain := range s.chains {
		for i := range chain.Validators {
			var opts docker.RemoveContainerOptions
			opts.ID = s.valResources[chain.ChainMeta.Id][i].Container.ID
			opts.Force = true
			s.dkrPool.Client.RemoveContainer(opts)
			s.T().Logf("removed container: %s", s.valResources[chain.ChainMeta.Id][i].Container.Name[1:])
		}
	}
	// upgrade to locally compiled binary concurently
	wg.Add(2)
	go s.upgradeContainers(s.chains[0], &wg)
	go s.upgradeContainers(s.chains[1], &wg)
	wg.Wait()
}

func (s *IntegrationTestSuite) upgradeContainers(c *chain.Chain, wg *sync.WaitGroup) {
	defer wg.Done()
	// upgrade containers to the locally compiled daemon
	s.T().Logf("starting upgrade for chain-id: %s...", c.ChainMeta.Id)
	pwd, err := os.Getwd()
	s.Require().NoError(err)
	for i, val := range c.Validators {
		runOpts := &dockertest.RunOptions{
			Name:       val.Name,
			Repository: "osmosis",
			Tag:        "debug",
			NetworkID:  s.dkrNet.Network.ID,
			User:       "root:root",
			Mounts: []string{
				fmt.Sprintf("%s/:/osmosis/.osmosisd", val.ConfigDir),
				fmt.Sprintf("%s/scripts:/osmosis", pwd),
			},
		}
		resource, err := s.dkrPool.RunWithOptions(runOpts, noRestart)
		s.Require().NoError(err)

		s.valResources[c.ChainMeta.Id][i] = resource
		s.T().Logf("started %s validator container: %s", resource.Container.Name[1:], resource.Container.ID)
	}

	// check that we are creating blocks again
	for i := range c.Validators {
		if c.ChainMeta.Id == chain.ChainAID {
			propHeight = propHeightA
		} else {
			propHeight = propHeightB
		}
		s.Require().Eventually(
			func() bool {
				currentHeight := s.chainHeight(s.valResources[c.ChainMeta.Id][i].Container.ID)
				if currentHeight <= propHeight {
					s.T().Logf("current block height on %s is %v, creating a few blocks container: %s", s.valResources[c.ChainMeta.Id][i].Container.Name[1:], currentHeight, s.valResources[c.ChainMeta.Id][i].Container.ID)
				}
				return currentHeight > propHeight+2
			},
			5*time.Minute,
			time.Second,
		)
		s.T().Logf("upgrade successful on %s validator container: %s", s.valResources[c.ChainMeta.Id][i].Container.Name[1:], s.valResources[c.ChainMeta.Id][i].Container.ID)
	}

}
