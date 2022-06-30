package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/containers"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/initialization"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
)

type status struct {
	LatestHeight string `json:"latest_block_height"`
}

type syncInfo struct {
	SyncInfo status `json:"SyncInfo"`
}

type validatorConfig struct {
	validator       initialization.Validator
	operatorAddress string
}

type chainConfig struct {
	// voting period is number of blocks it takes to deposit, 1.2 seconds per validator to vote on the prop, and a buffer.
	votingPeriod float32
	// upgrade proposal height for chain.
	propHeight int
	forkHeight int
	// Indexes of the validators to skip from running during initialization.
	// This is needed for testing functionality like state-sync where we would
	// like to start a node during tests post-initialization.
	skipRunValidatorIndexes map[int]struct{}
	latestProposalNumber    int
	latestLockNumber        int
	meta                    initialization.ChainMeta
	validators              []*validatorConfig
}

const (
	// Environment variable name to skip the upgrade tests
	skipUpgradeEnv = "OSMOSIS_E2E_SKIP_UPGRADE"
	// Environment variable name to skip the IBC tests
	skipIBCEnv = "OSMOSIS_E2E_SKIP_IBC"
	// Environment variable name to determine if this upgrade is a fork
	forkHeightEnv = "OSMOSIS_E2E_FORK_HEIGHT"
	// Environment variable name to skip cleaning up Docker resources in teardown
	skipCleanupEnv = "OSMOSIS_E2E_SKIP_CLEANUP"
	// Environment variable name to determine what version we are upgrading to
	upgradeVersionEnv = "OSMOSIS_E2E_UPGRADE_VERSION"
	// if not skipping upgrade, how many blocks we allow for fork to run pre upgrade state creation
	forkHeightPreUpgradeOffset int = 60
	// estimated number of blocks it takes to submit for a proposal
	propSubmitBlocks float32 = 10
	// estimated number of blocks it takes to deposit for a proposal
	propDepositBlocks float32 = 10
	// number of blocks it takes to vote for a single validator to vote for a proposal
	propVoteBlocks float32 = 1.2
	// number of blocks used as a calculation buffer
	propBufferBlocks float32 = 5
	// max retries for json unmarshalling
	maxRetries = 60
)

var (
	// whatever number of validator configs get posted here are how many validators that will spawn on chain A and B respectively
	validatorConfigsChainA = []*initialization.ValidatorConfig{
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
	validatorConfigsChainB = []*initialization.ValidatorConfig{
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

	tmpDirs          []string
	chainConfigs     []*chainConfig
	containerManager *containers.Manager
	skipUpgrade      bool
	skipIBC          bool
	isFork           bool
	forkHeight       int
	upgradeVersion   string
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")
	var forkHeight64 int64
	var err error

	s.chainConfigs = make([]*chainConfig, 0, 2)

	// The e2e test flow is as follows:
	//
	// 1. Configure two chains - chan A and chain B.
	//   * For each chain, set up two validators
	//   * Initialize configs and genesis for all validators.
	// 2. Start both networks.
	// 3. Run IBC relayer betweeen the two chains.
	// 4. Execute various e2e tests, including IBC.

	if str := os.Getenv(skipUpgradeEnv); len(str) > 0 {
		s.skipUpgrade, err = strconv.ParseBool(str)
		s.Require().NoError(err)

		if s.skipUpgrade {
			s.T().Log(fmt.Sprintf("%s was true, skipping upgrade tests", skipIBCEnv))
		}
	}

	if str := os.Getenv(forkHeightEnv); len(str) > 0 {
		forkHeight64, err = strconv.ParseInt(str, 0, 64)
		s.Require().NoError(err)
		s.forkHeight = int(forkHeight64)
		s.isFork = true

		s.T().Log(fmt.Sprintf("fork upgrade is enabled, %s was set to height %v", forkHeightEnv, s.forkHeight))
	}

	if str := os.Getenv(skipIBCEnv); len(str) > 0 {
		s.skipIBC, err = strconv.ParseBool(str)
		s.Require().NoError(err)

		if s.skipIBC {
			s.T().Log(fmt.Sprintf("%s was true, skipping IBC tests", skipIBCEnv))

			if !s.skipUpgrade {
				s.T().Fatal("If upgrade is enabled, IBC must be enabled as well.")
			}
		}
	}

	if str := os.Getenv(upgradeVersionEnv); len(str) > 0 {
		s.upgradeVersion = str

		s.T().Log(fmt.Sprintf("upgrade version set to %s", s.upgradeVersion))
	}

	s.containerManager, err = containers.NewManager(!s.skipUpgrade, s.isFork)
	require.NoError(s.T(), err)

	s.configureChain(initialization.ChainAID, validatorConfigsChainA, map[int]struct{}{
		3: {}, // skip validator at index 3
	})

	// We don't need a second chain if IBC is disabled
	if !s.skipIBC {
		s.configureChain(initialization.ChainBID, validatorConfigsChainB, map[int]struct{}{})
	}

	for i, chainConfig := range s.chainConfigs {
		s.runValidators(chainConfig, i*10)
		s.extractValidatorOperatorAddresses(chainConfig)
	}

	if !s.skipIBC {
		// Run a relayer between every possible pair of chains.
		for i := 0; i < len(s.chainConfigs); i++ {
			for j := i + 1; j < len(s.chainConfigs); j++ {
				s.runIBCRelayer(s.chainConfigs[i], s.chainConfigs[j])
			}
		}
	}

	if !s.skipUpgrade {
		s.createPreUpgradeState()

		if s.isFork {
			s.upgradeFork()
		} else {
			s.upgrade()
		}
	}

	s.runPostUpgradeTests()
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if str := os.Getenv(skipCleanupEnv); len(str) > 0 {
		skipCleanup, err := strconv.ParseBool(str)
		s.Require().NoError(err)

		if skipCleanup {
			s.T().Log("skipping e2e resources clean up...")
			return
		}
	}

	s.T().Log("tearing down e2e integration test suite...")

	err := s.containerManager.ClearResources()
	s.Require().NoError(err)

	for _, chainConfig := range s.chainConfigs {
		os.RemoveAll(chainConfig.meta.DataDir)
	}

	for _, td := range s.tmpDirs {
		os.RemoveAll(td)
	}
}

func (s *IntegrationTestSuite) runValidators(chainConfig *chainConfig, portOffset int) {
	s.T().Logf("starting %s validator containers...", chainConfig.meta.Id)
	for i, val := range chainConfig.validators {
		// Skip some validators from running during set up.
		// This is needed for testing functionality like
		// state-sync where we might want to start some validators during tests.
		if _, ok := chainConfig.skipRunValidatorIndexes[i]; ok {
			s.T().Logf("skipping %s validator with index %d from running...", val.validator.Name, i)
			continue
		}

		validatorResource, err := s.containerManager.RunValidatorResource(chainConfig.meta.Id, val.validator.Name, val.validator.ConfigDir)
		require.NoError(s.T(), err)
		s.T().Logf("started %s validator container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	}

	validatorHostPort, err := s.containerManager.GetValidatorHostPort(chainConfig.meta.Id, 0, "26657/tcp")
	require.NoError(s.T(), err)

	rpcClient, err := rpchttp.New(fmt.Sprintf("tcp://%s", validatorHostPort), "/websocket")
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

func (s *IntegrationTestSuite) runIBCRelayer(chainA *chainConfig, chainB *chainConfig) {
	s.T().Log("starting Hermes relayer container...")

	tmpDir, err := os.MkdirTemp("", "osmosis-e2e-testnet-hermes-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	osmoAVal := chainA.validators[0].validator
	osmoBVal := chainB.validators[0].validator
	hermesCfgPath := path.Join(tmpDir, "hermes")

	s.Require().NoError(os.MkdirAll(hermesCfgPath, 0o755))
	_, err = util.CopyFile(
		filepath.Join("./scripts/", "hermes_bootstrap.sh"),
		filepath.Join(hermesCfgPath, "hermes_bootstrap.sh"),
	)
	s.Require().NoError(err)

	hermesResource, err := s.containerManager.RunHermesResource(chainA.meta.Id, osmoAVal.Mnemonic, chainB.meta.Id, osmoBVal.Mnemonic, hermesCfgPath)
	require.NoError(s.T(), err)

	endpoint := fmt.Sprintf("http://%s/state", hermesResource.GetHostPort("3031/tcp"))
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

	s.T().Logf("started Hermes relayer container: %s", hermesResource.Container.ID)

	// XXX: Give time to both networks to start, otherwise we might see gRPC
	// transport errors.
	time.Sleep(10 * time.Second)

	// create the client, connection and channel between the two Osmosis chains
	s.connectIBCChains(chainA, chainB)
}

func (s *IntegrationTestSuite) configureChain(chainId string, validatorConfigs []*initialization.ValidatorConfig, skipValidatorIndexes map[int]struct{}) {
	s.T().Logf("starting e2e infrastructure for chain-id: %s", chainId)
	tmpDir, err := os.MkdirTemp("", "osmosis-e2e-testnet-")

	s.T().Logf("temp directory for chain-id %v: %v", chainId, tmpDir)
	s.Require().NoError(err)

	validatorConfigBytes, err := json.Marshal(validatorConfigs)
	s.Require().NoError(err)

	numVal := float32(len(validatorConfigs))

	newChainConfig := chainConfig{
		votingPeriod:            propDepositBlocks + numVal*propVoteBlocks + propBufferBlocks,
		skipRunValidatorIndexes: skipValidatorIndexes,
	}

	// If upgrade is skipped, we can use the chain initialization logic from
	// current branch directly. As a result, there is no need to run this
	// via Docker.

	if s.skipUpgrade {
		initializedChain, err := initialization.Init(chainId, tmpDir, validatorConfigs, time.Duration(newChainConfig.votingPeriod), s.forkHeight)
		s.Require().NoError(err)
		s.initializeChainConfig(&newChainConfig, initializedChain)
		return
	}

	if s.isFork {
		s.forkHeight = s.forkHeight - forkHeightPreUpgradeOffset
	}

	initResource, err := s.containerManager.RunChainInitResource(chainId, int(newChainConfig.votingPeriod), validatorConfigBytes, tmpDir, s.forkHeight)
	s.Require().NoError(err)

	fileName := fmt.Sprintf("%v/%v-encode", tmpDir, chainId)
	s.T().Logf("serialized init file for chain-id %v: %v", chainId, fileName)
	var initializedChain initialization.Chain
	// loop through the reading and unmarshaling of the init file a total of maxRetries or until error is nil
	// without this, test attempts to unmarshal file before docker container is finished writing
	for i := 0; i < maxRetries; i++ {
		encJson, _ := os.ReadFile(fileName)
		// err = json.Unmarshal(encJson, &newChainConfig.validators)
		err = json.Unmarshal(encJson, &initializedChain)
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

	s.Require().NoError(s.containerManager.PurgeResource(initResource))

	s.initializeChainConfig(&newChainConfig, &initializedChain)
}

func (s *IntegrationTestSuite) initializeChainConfig(chainConfig *chainConfig, initializedChain *initialization.Chain) {
	chainConfig.meta.DataDir = initializedChain.ChainMeta.DataDir
	chainConfig.meta.Id = initializedChain.ChainMeta.Id

	chainConfig.validators = make([]*validatorConfig, 0, len(initializedChain.Validators))
	for _, val := range initializedChain.Validators {
		chainConfig.validators = append(chainConfig.validators, &validatorConfig{validator: *val})
	}

	s.chainConfigs = append(s.chainConfigs, chainConfig)
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
	for _, chainConfig := range s.chainConfigs {
		currentHeight := s.getCurrentChainHeight(chainConfig, 0)
		chainConfig.propHeight = currentHeight + int(chainConfig.votingPeriod) + int(propSubmitBlocks) + int(propBufferBlocks)
		s.submitUpgradeProposal(chainConfig)
		s.depositProposal(chainConfig)
		s.voteProposal(chainConfig)
	}

	// wait till all chains halt at upgrade height
	for _, chainConfig := range s.chainConfigs {
		for i := range chainConfig.validators {
			if _, ok := chainConfig.skipRunValidatorIndexes[i]; ok {
				continue
			}

			validatorResource, exists := s.containerManager.GetValidatorResource(chainConfig.meta.Id, i)
			require.True(s.T(), exists)
			containerId := validatorResource.Container.ID
			containerName := validatorResource.Container.Name[1:]

			// use counter to ensure no new blocks are being created
			counter := 0
			s.T().Logf("waiting to reach upgrade height on %s validator container: %s", containerName, containerId)
			s.Require().Eventually(
				func() bool {
					currentHeight := s.getCurrentChainHeight(chainConfig, i)
					if currentHeight != chainConfig.propHeight {
						s.T().Logf("current block height on %s is %v, waiting for block %v container: %s", containerName, currentHeight, chainConfig.propHeight, containerId)
					}
					if currentHeight > chainConfig.propHeight {
						panic("chain did not halt at upgrade height")
					}
					if currentHeight == chainConfig.propHeight {
						counter++
					}
					return counter == 3
				},
				5*time.Minute,
				time.Second,
			)
			s.T().Logf("reached upgrade height on %s container: %s", containerName, containerId)
		}
	}

	// remove all containers so we can upgrade them to the new version
	for _, chainConfig := range s.chainConfigs {
		for valIdx, val := range chainConfig.validators {
			if _, ok := chainConfig.skipRunValidatorIndexes[valIdx]; ok {
				continue
			}
			containerName := val.validator.Name
			err := s.containerManager.RemoveValidatorResource(chainConfig.meta.Id, containerName)
			s.Require().NoError(err)
			s.T().Logf("removed container: %s", containerName)
		}
	}

	for _, chainConfig := range s.chainConfigs {
		s.upgradeContainers(chainConfig, chainConfig.propHeight)
	}
}

func (s *IntegrationTestSuite) upgradeFork() {

	for _, chainConfig := range s.chainConfigs {

		for i := range chainConfig.validators {
			if _, ok := chainConfig.skipRunValidatorIndexes[i]; ok {
				continue
			}

			validatorResource, exists := s.containerManager.GetValidatorResource(chainConfig.meta.Id, i)
			require.True(s.T(), exists)
			containerId := validatorResource.Container.ID
			containerName := validatorResource.Container.Name[1:]

			s.T().Logf("waiting to reach fork height on %s validator container: %s", containerName, containerId)
			s.Require().Eventually(
				func() bool {
					currentHeight := s.getCurrentChainHeight(chainConfig, i)
					if currentHeight < s.forkHeight {
						s.T().Logf("current block height on %s is %v, waiting for block %v container: %s", containerName, currentHeight, s.forkHeight, containerId)
						return false
					}
					return true
				},
				5*time.Minute,
				time.Second,
			)
			s.T().Logf("successfully got past fork height on %s container: %s", containerName, containerId)
		}
	}
}

func (s *IntegrationTestSuite) upgradeContainers(chainConfig *chainConfig, propHeight int) {
	// upgrade containers to the locally compiled daemon
	chain := chainConfig
	s.T().Logf("starting upgrade for chain-id: %s...", chain.meta.Id)

	s.containerManager.OsmosisRepository = containers.CurrentBranchOsmoRepository
	s.containerManager.OsmosisTag = containers.CurrentBranchOsmoTag

	for i, val := range chain.validators {
		if _, ok := chainConfig.skipRunValidatorIndexes[i]; ok {
			continue
		}
		validatorResource, err := s.containerManager.RunValidatorResource(chainConfig.meta.Id, val.validator.Name, val.validator.ConfigDir)
		require.NoError(s.T(), err)
		s.T().Logf("started %s validator container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	}

	// check that we are creating blocks again
	for i := range chain.validators {
		if _, ok := chainConfig.skipRunValidatorIndexes[i]; ok {
			continue
		}

		validatorResource, exists := s.containerManager.GetValidatorResource(chainConfig.meta.Id, i)
		require.True(s.T(), exists)

		s.Require().Eventually(
			func() bool {
				currentHeight := s.getCurrentChainHeight(chainConfig, i)
				if currentHeight <= propHeight {
					s.T().Logf("current block height on %s is %v, waiting to create blocks container: %s", validatorResource.Container.Name[1:], currentHeight, validatorResource.Container.ID)
				}
				return currentHeight > propHeight
			},
			5*time.Minute,
			time.Second,
		)
		s.T().Logf("upgrade successful on %s validator container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	}
}

func (s *IntegrationTestSuite) createPreUpgradeState() {
	chainA := s.chainConfigs[0]
	chainB := s.chainConfigs[1]

	s.sendIBC(chainA, chainB, chainB.validators[0].validator.PublicAddress, initialization.OsmoToken)
	s.sendIBC(chainB, chainA, chainA.validators[0].validator.PublicAddress, initialization.OsmoToken)
	s.sendIBC(chainA, chainB, chainB.validators[0].validator.PublicAddress, initialization.StakeToken)
	s.sendIBC(chainB, chainA, chainA.validators[0].validator.PublicAddress, initialization.StakeToken)
	s.createPool(chainA, "pool1A.json")
	s.createPool(chainB, "pool1B.json")
}

func (s *IntegrationTestSuite) runPostUpgradeTests() {
	if s.skipIBC {
		return
	}

	chainA := s.chainConfigs[0]
	chainB := s.chainConfigs[1]

	s.sendIBC(chainA, chainB, chainB.validators[0].validator.PublicAddress, initialization.OsmoToken)
	s.sendIBC(chainB, chainA, chainA.validators[0].validator.PublicAddress, initialization.OsmoToken)
	s.sendIBC(chainA, chainB, chainB.validators[0].validator.PublicAddress, initialization.StakeToken)
	s.sendIBC(chainB, chainA, chainA.validators[0].validator.PublicAddress, initialization.StakeToken)
	s.createPool(chainA, "pool2A.json")
	s.createPool(chainB, "pool2B.json")
}
