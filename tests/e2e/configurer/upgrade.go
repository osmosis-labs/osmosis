package configurer

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/configurer/chain"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/configurer/config"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/containers"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/initialization"
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
	t.Helper()
	return &UpgradeConfigurer{
		baseConfigurer: baseConfigurer{
			chainConfigs:     chainConfigs,
			containerManager: containerManager,
			setupTests:       setupTests,
			syncUntilHeight:  forkHeight + defaultSyncUntilHeight,
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
	tmpDir, err := os.MkdirTemp("", "symphony-e2e-testnet-")
	if err != nil {
		return err
	}

	validatorConfigBytes, err := json.Marshal(chainConfig.ValidatorInitConfigs)
	if err != nil {
		return err
	}

	forkHeight := uc.forkHeight
	if forkHeight > 0 {
		forkHeight = forkHeight - config.ForkHeightPreUpgradeOffset
	}

	chainInitResource, err := uc.containerManager.RunChainInitResource(chainConfig.Id, int(chainConfig.VotingPeriod), int(chainConfig.ExpeditedVotingPeriod), validatorConfigBytes, tmpDir, int(forkHeight))
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

func (uc *UpgradeConfigurer) CreatePreUpgradeState() error {
	// Create a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	chainA := uc.chainConfigs[0]
	chainANode, err := chainA.GetDefaultNode()
	if err != nil {
		return err
	}
	chainB := uc.chainConfigs[1]
	chainBNode, err := chainB.GetDefaultNode()
	if err != nil {
		return err
	}

	wg.Add(2)

	go func() {
		defer wg.Done()
		chainA.SendIBC(chainB, chainBNode.PublicAddress, initialization.MelodyToken)
		chainA.SendIBC(chainB, chainBNode.PublicAddress, initialization.StakeToken)
	}()

	go func() {
		defer wg.Done()
		chainB.SendIBC(chainA, chainANode.PublicAddress, initialization.MelodyToken)
		chainB.SendIBC(chainA, chainANode.PublicAddress, initialization.StakeToken)
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	var (
		poolShareDenom             = make([]string, 2)
		preUpgradePoolId           = make([]uint64, 2)
		preUpgradeStableSwapPoolId = make([]uint64, 2)
	)

	// Increment the WaitGroup counter for each goroutine
	wg.Add(4)

	// Chain A

	go func() {
		defer wg.Done()
		preUpgradePoolId[0] = chainANode.CreateBalancerPool("pool1A.json", initialization.ValidatorWalletName)
		poolShareDenom[0] = fmt.Sprintf("gamm/pool/%d", preUpgradePoolId[0])
		chainANode.EnableSuperfluidAsset(chainA, poolShareDenom[0], true)
	}()

	go func() {
		defer wg.Done()
		preUpgradeStableSwapPoolId[0] = chainANode.CreateStableswapPool("stablePool.json", initialization.ValidatorWalletName)
	}()

	// Chain B

	go func() {
		defer wg.Done()
		preUpgradePoolId[1] = chainBNode.CreateBalancerPool("pool1B.json", initialization.ValidatorWalletName)
		poolShareDenom[1] = fmt.Sprintf("gamm/pool/%d", preUpgradePoolId[1])
		chainBNode.EnableSuperfluidAsset(chainB, poolShareDenom[1], true)
	}()

	go func() {
		defer wg.Done()
		preUpgradeStableSwapPoolId[1] = chainBNode.CreateStableswapPool("stablePool.json", initialization.ValidatorWalletName)
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	config.PreUpgradePoolId = preUpgradePoolId
	config.PreUpgradeStableSwapPoolId = preUpgradeStableSwapPoolId

	var (
		lockupWallet           = make([]string, 2)
		lockupWalletSuperfluid = make([]string, 2)
		stableswapWallet       = make([]string, 2)
	)

	wg.Add(6)

	amountToFund := "10000000000000000000"
	// Chain A
	go func() {
		defer wg.Done()
		// Setup wallets and send tokens to wallets (only chainA)
		lockupWallet[0] = chainANode.CreateWalletAndFund(config.LockupWallet[0], []string{
			amountToFund + poolShareDenom[0],
		}, chainA)
	}()

	go func() {
		defer wg.Done()
		lockupWalletSuperfluid[0] = chainANode.CreateWalletAndFund(config.LockupWalletSuperfluid[0], []string{
			amountToFund + poolShareDenom[0],
		}, chainA)
	}()

	go func() {
		defer wg.Done()
		stableswapWallet[0] = chainANode.CreateWalletAndFund(config.StableswapWallet[0], []string{
			"100000stake",
		}, chainA)
	}()

	// Chain B
	go func() {
		defer wg.Done()
		// Setup wallets and send tokens to wallets (only chainA)
		lockupWallet[1] = chainBNode.CreateWalletAndFund(config.LockupWallet[1], []string{
			amountToFund + poolShareDenom[1],
		}, chainB)
	}()

	go func() {
		defer wg.Done()
		lockupWalletSuperfluid[1] = chainBNode.CreateWalletAndFund(config.LockupWalletSuperfluid[1], []string{
			amountToFund + poolShareDenom[1],
		}, chainB)
	}()

	go func() {
		defer wg.Done()
		stableswapWallet[1] = chainBNode.CreateWalletAndFund(config.StableswapWallet[1], []string{
			"100000stake",
		}, chainB)
	}()

	wg.Wait()

	config.LockupWallet = lockupWallet
	config.LockupWalletSuperfluid = lockupWalletSuperfluid
	config.StableswapWallet = stableswapWallet

	wg.Add(6)

	var errCh = make(chan error, 2)

	// Chain A

	go func() {
		defer wg.Done()
		// test swap exact amount in for stable swap pool
		chainANode.SwapExactAmountIn("2000stake", "1", fmt.Sprintf("%d", config.PreUpgradeStableSwapPoolId[0]), "note", config.StableswapWallet[0])
	}()

	go func() {
		defer wg.Done()
		uc.t.Logf("Uploading rate limiting contract to chainA")
		_, err := chainANode.SetupRateLimiting("", chainANode.QueryGovModuleAccount(), chainA, true)
		errCh <- err
	}()

	go func() {
		defer wg.Done()
		uc.t.Logf("Lock and add to existing lock for both regular and superfluid lockups on chainA")
		chainANode.LockAndAddToExistingLock(chainA, osmomath.NewInt(1000000000000000000), poolShareDenom[0], config.LockupWallet[0], config.LockupWalletSuperfluid[0])
	}()

	// Chain B

	go func() {
		defer wg.Done()
		// test swap exact amount in for stable swap pool
		chainBNode.SwapExactAmountIn("2000stake", "1", fmt.Sprintf("%d", config.PreUpgradeStableSwapPoolId[1]), "note", config.StableswapWallet[1])
	}()

	go func() {
		defer wg.Done()
		uc.t.Logf("Uploading rate limiting contract to chainB")
		_, err := chainBNode.SetupRateLimiting("", chainBNode.QueryGovModuleAccount(), chainB, true)
		errCh <- err
	}()

	go func() {
		defer wg.Done()
		uc.t.Logf("Lock and add to existing lock for both regular and superfluid lockups on chainB")
		chainBNode.LockAndAddToExistingLock(chainB, osmomath.NewInt(1000000000000000000), poolShareDenom[1], config.LockupWallet[1], config.LockupWalletSuperfluid[1])
	}()

	wg.Wait()

	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (uc *UpgradeConfigurer) RunSetup() error {
	return uc.setupTests(uc)
}

func (uc *UpgradeConfigurer) RunUpgrade() error {
	var err error
	if uc.forkHeight > 0 {
		uc.runForkUpgrade()
	} else {
		err = uc.runProposalUpgrade()
	}
	if err != nil {
		return err
	}

	// Check if the nodes are running
	for chainIndex, chainConfig := range uc.chainConfigs {
		chain := uc.baseConfigurer.GetChainConfig(chainIndex)
		for validatorIdx := range chainConfig.NodeConfigs {
			node := chain.NodeConfigs[validatorIdx]
			// Check node status
			_, err = node.Status()
			if err != nil {
				uc.t.Errorf("node is not running after upgrade, chain-id %s, node %s", chainConfig.Id, node.Name)
				return err
			}
			uc.t.Logf("node %s upgraded successfully, address %s", node.Name, node.PublicAddress)
		}
	}
	return nil
}

func (uc *UpgradeConfigurer) runProposalUpgrade() error {
	// submit, deposit, and vote for upgrade proposal
	// prop height = current height + voting period + time it takes to submit proposal + small buffer
	for _, chainConfig := range uc.chainConfigs {
		node, err := chainConfig.GetDefaultNode()
		if err != nil {
			return err
		}
		currentHeight, err := node.QueryCurrentHeight()
		if err != nil {
			return err
		}
		chainConfig.UpgradePropHeight = currentHeight + int64(chainConfig.VotingPeriod) + int64(config.PropSubmitBlocks) + int64(config.PropBufferBlocks)
		propNumber := node.SubmitUpgradeProposal(uc.upgradeVersion, chainConfig.UpgradePropHeight, sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(config.InitialMinDeposit)), true)

		node.DepositProposal(propNumber, false)

		chain.AllValsVoteOnProposal(chainConfig, propNumber)
	}

	// wait till all chains halt at upgrade height
	for _, chainConfig := range uc.chainConfigs {
		uc.t.Logf("waiting to reach upgrade height on chain %s", chainConfig.Id)
		chainConfig.WaitUntilHeight(chainConfig.UpgradePropHeight)
		uc.t.Logf("upgrade height reached on chain %s", chainConfig.Id)
	}

	// remove all containers so we can upgrade them to the new version
	for _, chainConfig := range uc.chainConfigs {
		for _, validatorConfig := range chainConfig.NodeConfigs {
			err := uc.containerManager.RemoveNodeResource(validatorConfig.Name)
			if err != nil {
				return err
			}
		}
	}

	// remove all containers so we can upgrade them to the new version
	for _, chainConfig := range uc.chainConfigs {
		if err := uc.upgradeContainers(chainConfig, chainConfig.UpgradePropHeight); err != nil {
			return err
		}
	}
	return nil
}

func (uc *UpgradeConfigurer) runForkUpgrade() {
	for _, chainConfig := range uc.chainConfigs {
		uc.t.Logf("waiting to reach fork height on chain %s", chainConfig.Id)
		chainConfig.WaitUntilHeight(uc.forkHeight)
		uc.t.Logf("fork height reached on chain %s", chainConfig.Id)
	}
}

func (uc *UpgradeConfigurer) upgradeContainers(chainConfig *chain.Config, propHeight int64) error {
	// upgrade containers to the locally compiled daemon
	uc.t.Logf("starting upgrade for chain-id: %s...", chainConfig.Id)
	uc.containerManager.SymphonyRepository = containers.CurrentBranchSymphonyRepository
	uc.containerManager.SymphonyTag = containers.CurrentBranchSymphonyTag

	for _, node := range chainConfig.NodeConfigs {
		if err := node.Run(); err != nil {
			return err
		}
	}

	uc.t.Logf("waiting to upgrade containers on chain %s", chainConfig.Id)
	chainConfig.WaitUntilHeight(propHeight)
	uc.t.Logf("upgrade successful on chain %s", chainConfig.Id)
	return nil
}
