package configurer

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/config"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/containers"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/initialization"
)

type UpgradeSettings struct {
	IsEnabled  bool
	Version    string
	ForkHeight int64 // non-zero height implies that this is a fork upgrade.
}

type UpgradeConfigurer struct {
	baseConfigurer
	upgradeVersion string
	forkHeight     int64 // forkHeight > 0 implies that this is a fork upgrade. Otherwise, proposal upgrade.
}

var _ Configurer = (*UpgradeConfigurer)(nil)

func NewUpgradeConfigurer(t *testing.T, chainConfigs []*chain.Config, setupTests setupFn, containerManager *containers.Manager, upgradeVersion string, forkHeight int64) Configurer {
	return &UpgradeConfigurer{
		baseConfigurer: baseConfigurer{
			chainConfigs:     chainConfigs,
			containerManager: containerManager,
			setupTests:       setupTests,
			t:                t,
		},
		forkHeight:     forkHeight,
		upgradeVersion: upgradeVersion,
	}
}

func (uc *UpgradeConfigurer) ConfigureChains() error {
	for _, chainConfig := range uc.chainConfigs {
		if err := uc.ConfigureChain(chainConfig); err != nil {
			return err
		}
	}
	return nil
}

func (uc *UpgradeConfigurer) ConfigureChain(chainConfig *chain.Config) error {
	uc.t.Logf("starting upgrade e2e infrastructure for chain-id: %s", chainConfig.Id)
	tmpDir, err := os.MkdirTemp("", "osmosis-e2e-testnet-")
	if err != nil {
		return err
	}

	numVal := float32(len(chainConfig.ValidatorInitConfigs))
	chainConfig.VotingPeriod = config.PropDepositBlocks + numVal*config.PropVoteBlocks + config.PropBufferBlocks

	validatorConfigBytes, err := json.Marshal(chainConfig.ValidatorInitConfigs)
	if err != nil {
		return err
	}

	forkHeight := uc.forkHeight
	if forkHeight > 0 {
		forkHeight = forkHeight - config.ForkHeightPreUpgradeOffset
	}

	chainInitResource, err := uc.containerManager.RunChainInitResource(chainConfig.Id, int(chainConfig.VotingPeriod), validatorConfigBytes, tmpDir, int(forkHeight))

	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%v/%v-encode", tmpDir, chainConfig.Id)
	uc.t.Logf("serialized init file for chain-id %v: %v", chainConfig.Id, fileName)

	// loop through the reading and unmarshaling of the init file a total of maxRetries or until error is nil
	// without this, test attempts to unmarshal file before docker container is finished writing
	var initializedChain initialization.Chain
	for i := 0; i < config.MaxRetries; i++ {
		initializedChainBytes, _ := os.ReadFile(fileName)
		err = json.Unmarshal(initializedChainBytes, &initializedChain)
		if err == nil {
			break
		}

		if i == config.MaxRetries-1 {
			if err != nil {
				return err
			}
		}

		if i > 0 {
			time.Sleep(1 * time.Second)
		}
	}
	if err := uc.containerManager.PurgeResource(chainInitResource); err != nil {
		return err
	}
	uc.initializeChainConfigFromInitChain(&initializedChain, chainConfig)
	return nil
}

func (uc *UpgradeConfigurer) CreatePreUpgradeState() {
	chainA := uc.chainConfigs[0]
	chainB := uc.chainConfigs[1]

	chainA.SendIBC(chainB, chainB.NodeConfigs[0].PublicAddress, initialization.OsmoToken)
	chainB.SendIBC(chainA, chainA.NodeConfigs[0].PublicAddress, initialization.OsmoToken)
	chainA.SendIBC(chainB, chainB.NodeConfigs[0].PublicAddress, initialization.StakeToken)
	chainB.SendIBC(chainA, chainA.NodeConfigs[0].PublicAddress, initialization.StakeToken)

	chainA.CreatePool("pool1A.json", initialization.ValidatorWalletName)
	chainB.CreatePool("pool1B.json", initialization.ValidatorWalletName)
}

func (uc *UpgradeConfigurer) RunSetup() error {
	return uc.setupTests(uc)
}

func (uc *UpgradeConfigurer) RunUpgrade() error {
	if uc.forkHeight > 0 {
		return uc.runForkUpgrade()
	}
	return uc.runProposalUpgrade()
}

func (uc *UpgradeConfigurer) runProposalUpgrade() error {
	// submit, deposit, and vote for upgrade proposal
	// prop height = current height + voting period + time it takes to submit proposal + small buffer
	for _, chainConfig := range uc.chainConfigs {
		currentHeight := chainConfig.QueryCurrentChainHeightFromValidator(0)
		chainConfig.PropHeight = currentHeight + int(chainConfig.VotingPeriod) + int(config.PropSubmitBlocks) + int(config.PropBufferBlocks)

		chainConfig.SubmitUpgradeProposal(uc.upgradeVersion)
		chainConfig.DepositProposal()
		chainConfig.VoteYesProposal()
	}

	// wait till all chains halt at upgrade height
	for _, chainConfig := range uc.chainConfigs {
		for validatorIndex := range chainConfig.NodeConfigs {
			validatorResource, exists := uc.containerManager.GetValidatorResource(chainConfig.Id, validatorIndex)
			require.True(uc.t, exists, "validator container not found: chain id %s, valIdx %d ", chainConfig.Id, validatorIndex)
			containerId := validatorResource.Container.ID
			containerName := validatorResource.Container.Name[1:]

			// use counter to ensure no new blocks are being created
			counter := 0
			uc.t.Logf("waiting to reach upgrade height on %s validator container: %s", containerName, containerId)
			require.Eventually(
				uc.t,
				func() bool {
					currentHeight := chainConfig.QueryCurrentChainHeightFromValidator(validatorIndex)
					if currentHeight != chainConfig.PropHeight {
						uc.t.Logf("current block height on %s is %v, waiting for block %v container: %s", containerName, currentHeight, chainConfig.PropHeight, containerId)
					}
					if currentHeight > chainConfig.PropHeight {
						panic("chain did not halt at upgrade height")
					}
					if currentHeight == chainConfig.PropHeight {
						counter++
					}
					return counter == 3
				},
				5*time.Minute,
				time.Second,
			)
			uc.t.Logf("reached upgrade height on %s container: %s", containerName, containerId)
		}
	}

	// remove all containers so we can upgrade them to the new version
	for _, chainConfig := range uc.chainConfigs {
		for _, validatorConfig := range chainConfig.NodeConfigs {
			err := uc.containerManager.RemoveValidatorResource(chainConfig.Id, validatorConfig.Name)
			if err != nil {
				return err
			}
		}
	}

	// remove all containers so we can upgrade them to the new version
	for _, chainConfig := range uc.chainConfigs {
		uc.upgradeContainers(chainConfig, chainConfig.PropHeight)
	}
	return nil
}

func (uc *UpgradeConfigurer) runForkUpgrade() error {
	for _, chainConfig := range uc.chainConfigs {
		for i := range chainConfig.NodeConfigs {
			validatorResource, exists := uc.containerManager.GetValidatorResource(chainConfig.Id, i)
			if !exists {
				return fmt.Errorf("validator container not found: chain id %s, valIdx %d ", chainConfig.Id, i)
			}
			containerId := validatorResource.Container.ID
			containerName := validatorResource.Container.Name[1:]

			uc.t.Logf("waiting to reach fork height on %s validator container: %s", containerName, containerId)
			require.Eventually(
				uc.t,
				func() bool {
					currentHeight := chainConfig.QueryCurrentChainHeightFromValidator(i)
					if int64(currentHeight) < uc.forkHeight {
						uc.t.Logf("current block height on %s is %v, waiting for block %v container: %s", containerName, currentHeight, uc.forkHeight, containerId)
						return false
					}
					return true
				},
				5*time.Minute,
				time.Second,
			)
			uc.t.Logf("successfully got past fork height on %s container: %s", containerName, containerId)
		}
	}
	return nil
}

func (uc *UpgradeConfigurer) upgradeContainers(chainConfig *chain.Config, propHeight int) {
	// upgrade containers to the locally compiled daemon
	uc.t.Logf("starting upgrade for chain-id: %s...", chainConfig.Id)
	uc.containerManager.OsmosisRepository = containers.CurrentBranchOsmoRepository
	uc.containerManager.OsmosisTag = containers.CurrentBranchOsmoTag
	for _, val := range chainConfig.NodeConfigs {
		validatorResource, err := uc.containerManager.RunValidatorResource(chainConfig.Id, val.Name, val.ConfigDir)
		require.NoError(uc.t, err)
		uc.t.Logf("started %s validator container: %s", validatorResource.Container.Name[1:], validatorResource.Container.ID)
	}

	// check that we are creating blocks again
	for validatorIndex := range chainConfig.NodeConfigs {
		validatorResource, exists := uc.containerManager.GetValidatorResource(chainConfig.Id, validatorIndex)
		require.True(uc.t, exists, "validator container not found: chain id %s, valIdx %d ", chainConfig.Id, validatorIndex)
		containerId := validatorResource.Container.ID
		containerName := validatorResource.Container.Name[1:]

		require.Eventually(
			uc.t,
			func() bool {
				currentHeight := chainConfig.QueryCurrentChainHeightFromValidator(validatorIndex)
				if currentHeight <= propHeight {
					uc.t.Logf("current block height on %s is %v, waiting to create blocks container: %s", containerName, currentHeight, containerId)
				}
				return currentHeight > propHeight
			},
			5*time.Minute,
			time.Second,
		)
		uc.t.Logf("upgrade successful on %s validator container: %s", containerName, containerId)
	}
}
