package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	dockerconfig "github.com/osmosis-labs/osmosis/v7/tests/e2e/docker"
	net "github.com/osmosis-labs/osmosis/v7/tests/e2e/network"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
)

type status struct {
	LatestHeight string `json:"latest_block_height"`
}

type syncInfo struct {
	SyncInfo status `json:"SyncInfo"`
}

const (
	// osmosis version being upgraded to (folder must exist here https://github.com/osmosis-labs/osmosis/tree/main/app/upgrades)
	upgradeVersion = "v9"

	// max retries for json unmarshalling
	maxRetries = 60
)

var (
	// whatever number of validator configs get posted here are how many validators that will spawn on chain A and B respectively
	validatorConfigsChainA = []*chain.ValidatorConfig{
		{
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   20,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   30,
			SnapshotKeepRecent: 1,
		},
		{
			Pruning:            "custom",
			PruningKeepRecent:  "10000",
			PruningInterval:    "13",
			SnapshotInterval:   15,
			SnapshotKeepRecent: 3,
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

type IntegrationTestSuite struct {
	suite.Suite

	tmpDirs  []string
	networks []*net.Network

	workingDirectory string

	dockerImages    *dockerconfig.ImageConfig
	dockerResources *dockerconfig.Resources
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")

	s.networks = make([]*net.Network, 0, 2)

	// The e2e test flow is as follows:
	//
	// 1. Configure two chains - chan A and chain B.
	//   * For each chain, set up two validators
	//   * Initialize configs and genesis for all validators.
	// 2. Start both networks.
	// 3. Run IBC relayer betweeen the two chains.
	// 4. Execute various e2e tests, including IBC.
	var (
		skipUpgrade bool
		err         error
	)

	s.workingDirectory, err = os.Getwd()
	s.Require().NoError(err)

	if str := os.Getenv("OSMOSIS_E2E_SKIP_UPGRADE"); len(str) > 0 {
		skipUpgrade, err = strconv.ParseBool(str)
		s.Require().NoError(err)
	}

	s.dockerImages = dockerconfig.NewImageConfig(!skipUpgrade)

	s.dockerResources, err = dockerconfig.NewResources()
	s.Require().NoError(err)

	s.configureChain(chain.ChainAID, validatorConfigsChainA)
	s.configureChain(chain.ChainBID, validatorConfigsChainB)

	for _, network := range s.networks {
		networkResources, err := network.RunValidators()
		s.Require().NoError(err)
		s.dockerResources.Validators[network.GetChain().ChainMeta.Id] = networkResources
	}

	// Run a relayer between every possible pair of chains.
	for i := 0; i < len(s.networks); i++ {
		for j := i + 1; j < len(s.networks); j++ {
			s.runIBCRelayer(s.networks[i].GetChain(), s.networks[j].GetChain())
		}
	}

	if !skipUpgrade {
		s.createPreUpgradeState()
		s.upgrade()
		s.runPostUpgradeTests()
	}

	// Stop a validator container so that we can restart it later
	// for testing state sync.
	if err := s.networks[0].RemoveValidatorContainer(3); err != nil {
		s.Require().NoError(err)
	}

	maxSnapshotInterval := uint64(0)

	for _, valConfig := range validatorConfigsChainA {
		if valConfig.SnapshotInterval > maxSnapshotInterval {
			maxSnapshotInterval = valConfig.SnapshotInterval
		}
	}

	// ensure we cover enough heights for a few snapshots to be taken

	doneCondition := func(syncInfo coretypes.SyncInfo) bool {
		return syncInfo.LatestBlockHeight > int64(maxSnapshotInterval)*2
	}

	err = s.networks[0].WaitUntil(0, doneCondition)
	s.Require().NoError(err)

	currentHeight, err := s.networks[0].GetCurrentHeightFromValidator(0)
	s.Require().NoError(err)

	// Ensure that state sync trust height is slightly lower than the latest
	// snapshot of every node
	stateSyncTrustHeight := int64(currentHeight - int64(float32(maxSnapshotInterval)*1.5))
	stateSyncTrustHash, err := s.networks[0].GetHashFromBlock(stateSyncTrustHeight)
	s.Require().NoError(err)

	//blockId := coretypes.ResultBlock
	err = configureNodeForStateSync(s.networks[0].GetChain().Validators[3].ConfigDir, stateSyncTrustHeight, stateSyncTrustHash)
	s.Require().NoError(err)
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

	err := s.dockerResources.Purge()
	s.Require().NoError(err)

	for _, network := range s.networks {
		os.RemoveAll(network.GetChain().ChainMeta.DataDir)
	}

	for _, td := range s.tmpDirs {
		os.RemoveAll(td)
	}
}

func (s *IntegrationTestSuite) runIBCRelayer(chainA *chain.Chain, chainB *chain.Chain) {
	s.T().Log("starting Hermes relayer container...")

	tmpDir, err := ioutil.TempDir("", "osmosis-e2e-testnet-hermes-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	osmoAVal := chainA.Validators[0]
	osmoBVal := chainB.Validators[0]
	hermesCfgPath := path.Join(tmpDir, "hermes")

	s.Require().NoError(os.MkdirAll(hermesCfgPath, 0o755))
	_, err = util.CopyFile(
		filepath.Join("./scripts/", "hermes_bootstrap.sh"),
		filepath.Join(hermesCfgPath, "hermes_bootstrap.sh"),
	)
	s.Require().NoError(err)

	s.dockerResources.Hermes, err = s.dockerResources.Pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s-%s-relayer", chainA.ChainMeta.Id, chainB.ChainMeta.Id),
			Repository: s.dockerImages.RelayerRepository,
			Tag:        s.dockerImages.RelayerTag,
			NetworkID:  s.dockerResources.Network.Network.ID,
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
				fmt.Sprintf("OSMO_A_E2E_VAL_HOST=%s", s.dockerResources.Validators[chainA.ChainMeta.Id][0].Container.Name[1:]),
				fmt.Sprintf("OSMO_B_E2E_VAL_HOST=%s", s.dockerResources.Validators[chainB.ChainMeta.Id][0].Container.Name[1:]),
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

	endpoint := fmt.Sprintf("http://%s/state", s.dockerResources.Hermes.GetHostPort("3031/tcp"))
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

	s.T().Logf("started Hermes relayer container: %s", s.dockerResources.Hermes.Container.ID)

	// XXX: Give time to both networks to start, otherwise we might see gRPC
	// transport errors.
	time.Sleep(10 * time.Second)

	// create the client, connection and channel between the two Osmosis chains
	s.connectIBCChains(chainA, chainB)
}

func (s *IntegrationTestSuite) configureChain(chainId string, validatorConfigs []*chain.ValidatorConfig) {
	s.T().Logf("starting e2e infrastructure for chain-id: %s", chainId)
	tmpDir, err := ioutil.TempDir("", "osmosis-e2e-testnet-")

	s.T().Logf("temp directory for chain-id %v: %v", chainId, tmpDir)
	s.Require().NoError(err)

	validatorConfigBytes, err := json.Marshal(validatorConfigs)
	s.Require().NoError(err)

	newNetwork := net.New(s.T(), len(s.networks), len(validatorConfigs), s.dockerResources, s.dockerImages, s.workingDirectory)

	votingPeriodDuration := time.Duration(int(newNetwork.GetVotingPeriod()) * 1000000000)

	initResource, err := s.dockerResources.Pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s", chainId),
			Repository: s.dockerImages.InitRepository,
			Tag:        s.dockerImages.InitTag,
			NetworkID:  s.dockerResources.Network.Network.ID,
			Cmd: []string{
				fmt.Sprintf("--data-dir=%s", tmpDir),
				fmt.Sprintf("--chain-id=%s", chainId),
				fmt.Sprintf("--config=%s", validatorConfigBytes),
				fmt.Sprintf("--voting-period=%v", votingPeriodDuration),
			},
			User: "root:root",
			Mounts: []string{
				fmt.Sprintf("%s:%s", tmpDir, tmpDir),
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	fileName := fmt.Sprintf("%v/%v-encode", tmpDir, chainId)
	s.T().Logf("serialized init file for chain-id %v: %v", chainId, fileName)

	// loop through the reading and unmarshaling of the init file a total of maxRetries or until error is nil
	// without this, test attempts to unmarshal file before docker container is finished writing
	for i := 0; i < maxRetries; i++ {
		encJson, _ := os.ReadFile(fileName)
		err = json.Unmarshal(encJson, newNetwork.GetChain())
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
	s.Require().NoError(s.dockerResources.Pool.Purge(initResource))

	s.networks = append(s.networks, newNetwork)
}

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}

func (s *IntegrationTestSuite) upgrade() {
	// submit, deposit, and vote for upgrade proposal
	// prop height = current height + voting period + time it takes to submit proposal + small buffer
	for _, network := range s.networks {
		currentHeight, err := network.GetCurrentHeightFromValidator(0)
		s.Require().NoError(err)

		network.CalclulateAndSetProposalHeight(currentHeight)
		proposalHeight := network.GetProposalHeight()

		curChain := network.GetChain()
		s.submitProposal(curChain, proposalHeight)
		s.depositProposal(curChain)
		s.voteProposal(network)
	}

	// wait till all chains halt at upgrade height
	for _, network := range s.networks {
		curChain := network.GetChain()

		for i := range curChain.Validators {

			// use counter to ensure no new blocks are being created
			counter := 0
			s.T().Logf("waiting to reach upgrade height on %s validator container: %s", s.dockerResources.Validators[curChain.ChainMeta.Id][i].Container.Name[1:], s.dockerResources.Validators[curChain.ChainMeta.Id][i].Container.ID)
			s.Require().Eventually(
				func() bool {
					currentHeight, err := network.GetCurrentHeightFromValidator(i)
					s.Require().NoError(err)
					propHeight := network.GetProposalHeight()

					if currentHeight != propHeight {
						s.T().Logf("current block height on %s is %v, waiting for block %v container: %s", s.dockerResources.Validators[curChain.ChainMeta.Id][i].Container.Name[1:], currentHeight, propHeight, s.dockerResources.Validators[curChain.ChainMeta.Id][i].Container.ID)
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
			s.T().Logf("reached upgrade height on %s container: %s", s.dockerResources.Validators[curChain.ChainMeta.Id][i].Container.Name[1:], s.dockerResources.Validators[curChain.ChainMeta.Id][i].Container.ID)
		}
	}

	// remove all containers so we can upgrade them to the new version
	for _, network := range s.networks {
		curChain := network.GetChain()
		for i := range curChain.Validators {
			if err := network.RemoveValidatorContainer(i); err != nil {
				s.Require().NoError(err)
			}
		}
	}

	for _, network := range s.networks {
		s.upgradeContainers(network)
	}
}

func (s *IntegrationTestSuite) upgradeContainers(network *net.Network) {
	// upgrade containers to the locally compiled daemon
	chain := network.GetChain()
	s.T().Logf("starting upgrade for chain-id: %s...", chain.ChainMeta.Id)
	pwd, err := os.Getwd()
	s.Require().NoError(err)

	for _, val := range chain.Validators {
		runOpts := &dockertest.RunOptions{
			Name:       val.Name,
			Repository: dockerconfig.LocalOsmoRepository,
			Tag:        dockerconfig.LocalOsmoTag,
			NetworkID:  s.dockerResources.Network.Network.ID,
			User:       "root:root",
			Mounts: []string{
				fmt.Sprintf("%s/:/osmosis/.osmosisd", val.ConfigDir),
				fmt.Sprintf("%s/scripts:/osmosis", pwd),
			},
		}
		resource, err := s.dockerResources.Pool.RunWithOptions(runOpts, noRestart)
		s.Require().NoError(err)

		s.dockerResources.Validators[chain.ChainMeta.Id][val.Index] = resource
		s.T().Logf("started %s validator container: %s", resource.Container.Name[1:], resource.Container.ID)
	}

	propHeight := network.GetProposalHeight()
	// check that we are creating blocks again
	for i := range chain.Validators {
		s.Require().Eventually(
			func() bool {
				currentHeight, err := network.GetCurrentHeightFromValidator(i)
				s.Require().NoError(err)
				if currentHeight <= propHeight {
					s.T().Logf("current block height on %s is %v, waiting to create blocks container: %s", s.dockerResources.Validators[chain.ChainMeta.Id][i].Container.Name[1:], currentHeight, s.dockerResources.Validators[chain.ChainMeta.Id][i].Container.ID)
				}
				return currentHeight > propHeight
			},
			5*time.Minute,
			time.Second,
		)
		s.T().Logf("upgrade successful on %s validator container: %s", s.dockerResources.Validators[chain.ChainMeta.Id][i].Container.Name[1:], s.dockerResources.Validators[chain.ChainMeta.Id][i].Container.ID)
	}
}

func (s *IntegrationTestSuite) createPreUpgradeState() {
	chainA := s.networks[0].GetChain()
	chainB := s.networks[1].GetChain()

	s.sendIBC(chainA, chainB, chainB.Validators[0].PublicAddress, chain.OsmoToken)
	s.sendIBC(chainB, chainA, chainA.Validators[0].PublicAddress, chain.OsmoToken)
	s.sendIBC(chainA, chainB, chainB.Validators[0].PublicAddress, chain.StakeToken)
	s.sendIBC(chainB, chainA, chainA.Validators[0].PublicAddress, chain.StakeToken)
	s.createPool(chainA, "pool1A.json")
	s.createPool(chainB, "pool1B.json")
}

func (s *IntegrationTestSuite) runPostUpgradeTests() {
	chainA := s.networks[0].GetChain()
	chainB := s.networks[1].GetChain()

	s.sendIBC(chainA, chainB, chainB.Validators[0].PublicAddress, chain.OsmoToken)
	s.sendIBC(chainB, chainA, chainA.Validators[0].PublicAddress, chain.OsmoToken)
	s.sendIBC(chainA, chainB, chainB.Validators[0].PublicAddress, chain.StakeToken)
	s.sendIBC(chainB, chainA, chainA.Validators[0].PublicAddress, chain.StakeToken)
	s.createPool(chainA, "pool2A.json")
	s.createPool(chainB, "pool2B.json")
}
