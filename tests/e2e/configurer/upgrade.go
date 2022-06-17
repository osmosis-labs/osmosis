package configurer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/containers"
)

type UpgradeConfigurer struct {
	baseConfigurer
}

var _ Configurer = (*UpgradeConfigurer)(nil)

func NewUpgradeConfigurer(t *testing.T, chainConfigs []*ChainConfig, setupTests setupFn, containerManager *containers.Manager) Configurer {
	return &UpgradeConfigurer{
		baseConfigurer: baseConfigurer{
			chainConfigs:     chainConfigs,
			containerManager: containerManager,
			setupTests:       setupTests,
			t:                t,
		},
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

func (uc *UpgradeConfigurer) ConfigureChain(chainConfig *ChainConfig) error {
	uc.t.Logf("starting upgrade e2e infrastructure for chain-id: %s", chainConfig.chainId)
	tmpDir, err := ioutil.TempDir("", "osmosis-e2e-testnet-")
	if err != nil {
		return err
	}
	uc.t.Logf("temp directory for chain-id %v: %v", chainConfig.chainId, tmpDir)

	validatorConfigBytes, err := json.Marshal(chainConfig.validatorConfig)
	if err != nil {
		return err
	}

	numVal := float32(len(chainConfig.validatorConfig))

	chainConfig.votingPeriod = PropDepositBlocks + numVal*PropVoteBlocks + PropBufferBlocks

	votingPeriodDuration := time.Duration(int(chainConfig.votingPeriod) * 1000000000)

	initResource, err := uc.containerManager.Pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s", chainConfig.chainId),
			Repository: uc.containerManager.ImageConfig.InitRepository,
			Tag:        uc.containerManager.ImageConfig.InitTag,
			NetworkID:  uc.containerManager.Network.Network.ID,
			Cmd: []string{
				fmt.Sprintf("--data-dir=%s", tmpDir),
				fmt.Sprintf("--chain-id=%s", chainConfig.chainId),
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
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%v/%v-encode", tmpDir, chainConfig.chainId)
	uc.t.Logf("serialized init file for chain-id %v: %v", chainConfig.chainId, fileName)

	// loop through the reading and unmarshaling of the init file a total of maxRetries or until error is nil
	// without this, test attempts to unmarshal file before docker container is finished writing
	for i := 0; i < MaxRetries; i++ {
		initializedChainBytes, _ := os.ReadFile(fileName)
		err = json.Unmarshal(initializedChainBytes, &chainConfig.chain)
		if err == nil {
			break
		}

		if i == MaxRetries-1 {
			if err != nil {
				return err
			}
		}

		if i > 0 {
			time.Sleep(1 * time.Second)
		}
	}
	if err := uc.containerManager.Pool.Purge(initResource); err != nil {
		return err
	}
	return nil
}

func (uc *UpgradeConfigurer) RunSetup() error {
	return uc.setupTests(uc)
}

func (uc *UpgradeConfigurer) RunUpgrade() error {
	// submit, deposit, and vote for upgrade proposal
	// prop height = current height + voting period + time it takes to submit proposal + small buffer
	for _, chainConfig := range uc.chainConfigs {
		currentHeight := uc.getCurrentChainHeight(uc.containerManager.ValResources[chainConfig.chain.ChainMeta.Id][0].Container.ID)
		chainConfig.propHeight = currentHeight + int(chainConfig.votingPeriod) + int(PropSubmitBlocks) + int(PropBufferBlocks)
		uc.submitProposal(chainConfig.chain, chainConfig.propHeight)
		uc.depositProposal(chainConfig.chain)
		uc.voteProposal(chainConfig)
	}

	// wait till all chains halt at upgrade height
	for _, chainConfig := range uc.chainConfigs {
		curChain := chainConfig.chain

		for i := range chainConfig.chain.Validators {
			if _, ok := chainConfig.skipRunValidatorIndexes[i]; ok {
				continue
			}

			// use counter to ensure no new blocks are being created
			counter := 0
			uc.t.Logf("waiting to reach upgrade height on %s validator container: %s", uc.containerManager.ValResources[curChain.ChainMeta.Id][i].Container.Name[1:], uc.containerManager.ValResources[curChain.ChainMeta.Id][i].Container.ID)
			require.Eventually(
				uc.t,
				func() bool {
					currentHeight := uc.getCurrentChainHeight(uc.containerManager.ValResources[curChain.ChainMeta.Id][i].Container.ID)
					if currentHeight != chainConfig.propHeight {
						uc.t.Logf("current block height on %s is %v, waiting for block %v container: %s", uc.containerManager.ValResources[curChain.ChainMeta.Id][i].Container.Name[1:], currentHeight, chainConfig.propHeight, uc.containerManager.ValResources[curChain.ChainMeta.Id][i].Container.ID)
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
			uc.t.Logf("reached upgrade height on %s container: %s", uc.containerManager.ValResources[curChain.ChainMeta.Id][i].Container.Name[1:], uc.containerManager.ValResources[curChain.ChainMeta.Id][i].Container.ID)
		}
	}

	// remove all containers so we can upgrade them to the new version
	for _, chainConfig := range uc.chainConfigs {
		curChain := chainConfig.chain
		for valIdx := range curChain.Validators {
			if _, ok := chainConfig.skipRunValidatorIndexes[valIdx]; ok {
				continue
			}

			var opts docker.RemoveContainerOptions
			opts.ID = uc.containerManager.ValResources[curChain.ChainMeta.Id][valIdx].Container.ID
			opts.Force = true
			if err := uc.containerManager.Pool.Client.RemoveContainer(opts); err != nil {
				return err
			}
			uc.t.Logf("removed container: %s", uc.containerManager.ValResources[curChain.ChainMeta.Id][valIdx].Container.Name[1:])
		}
	}

	// remove all containers so we can upgrade them to the new version
	for _, chainConfig := range uc.chainConfigs {
		uc.upgradeContainers(chainConfig, chainConfig.propHeight)
	}
	return nil
}

func (uc *UpgradeConfigurer) upgradeContainers(chainConfig *ChainConfig, propHeight int) {
	// upgrade containers to the locally compiled daemon
	chain := chainConfig.chain
	uc.t.Logf("starting upgrade for chain-id: %s...", chain.ChainMeta.Id)
	pwd, err := os.Getwd()
	require.NoError(uc.t, err)
	for i, val := range chain.Validators {
		if _, ok := chainConfig.skipRunValidatorIndexes[i]; ok {
			continue
		}

		runOpts := &dockertest.RunOptions{
			Name:       val.Name,
			Repository: containers.CurrentBranchOsmoRepository,
			Tag:        containers.CurrentBranchOsmoTag,
			NetworkID:  uc.containerManager.Network.Network.ID,
			User:       "root:root",
			Mounts: []string{
				fmt.Sprintf("%s/:/osmosis/.osmosisd", val.ConfigDir),
				fmt.Sprintf("%s/scripts:/osmosis", pwd),
			},
		}
		resource, err := uc.containerManager.Pool.RunWithOptions(runOpts, noRestart)
		require.NoError(uc.t, err)

		uc.containerManager.ValResources[chain.ChainMeta.Id][i] = resource
		uc.t.Logf("started %s validator container: %s", resource.Container.Name[1:], resource.Container.ID)
	}

	// check that we are creating blocks again
	for i := range chain.Validators {
		if _, ok := chainConfig.skipRunValidatorIndexes[i]; ok {
			continue
		}

		require.Eventually(
			uc.t,
			func() bool {
				currentHeight := uc.getCurrentChainHeight(uc.containerManager.ValResources[chain.ChainMeta.Id][i].Container.ID)
				if currentHeight <= propHeight {
					uc.t.Logf("current block height on %s is %v, waiting to create blocks container: %s", uc.containerManager.ValResources[chain.ChainMeta.Id][i].Container.Name[1:], currentHeight, uc.containerManager.ValResources[chain.ChainMeta.Id][i].Container.ID)
				}
				return currentHeight > propHeight
			},
			5*time.Minute,
			time.Second,
		)
		uc.t.Logf("upgrade successful on %s validator container: %s", uc.containerManager.ValResources[chain.ChainMeta.Id][i].Container.Name[1:], uc.containerManager.ValResources[chain.ChainMeta.Id][i].Container.ID)
	}
}

func (uc *UpgradeConfigurer) CreatePreUpgradeState() {
	chainA := uc.chainConfigs[0].chain
	chainB := uc.chainConfigs[1].chain

	uc.SendIBC(chainA, chainB, chainB.Validators[0].PublicAddress, chain.OsmoToken)
	uc.SendIBC(chainB, chainA, chainA.Validators[0].PublicAddress, chain.OsmoToken)
	uc.SendIBC(chainA, chainB, chainB.Validators[0].PublicAddress, chain.StakeToken)
	uc.SendIBC(chainB, chainA, chainA.Validators[0].PublicAddress, chain.StakeToken)
	uc.CreatePool(chainA.ChainMeta.Id, 0, "pool1A.json")
	uc.CreatePool(chainB.ChainMeta.Id, 0, "pool1B.json")
}
