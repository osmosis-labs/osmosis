//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	appparams "github.com/osmosis-labs/osmosis/v11/app/params"
	"github.com/osmosis-labs/osmosis/v11/tests/e2e/configurer/config"
	"github.com/osmosis-labs/osmosis/v11/tests/e2e/initialization"
)

const lockupWallet = "lockup-wallet"
const lockupWalletSuperfluid = "lockup-wallet-superfluid"

// Test01IBCTokenTransfer tests that IBC token transfers work as expected.
// This test must precede Test02CreatePoolPostUpgrade. That's why it is prefixed with "01"
// to ensure correct ordering. See Test02CreatePoolPostUpgrade for more details.
func (s *IntegrationTestSuite) Test01IBCTokenTransfer() {
	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	}
	chainA := s.configurer.GetChainConfig(0)
	chainB := s.configurer.GetChainConfig(1)
	chainA.SendIBC(chainB, chainB.NodeConfigs[0].PublicAddress, initialization.OsmoToken)
	chainB.SendIBC(chainA, chainA.NodeConfigs[0].PublicAddress, initialization.OsmoToken)
	chainA.SendIBC(chainB, chainB.NodeConfigs[0].PublicAddress, initialization.StakeToken)
	chainB.SendIBC(chainA, chainA.NodeConfigs[0].PublicAddress, initialization.StakeToken)
}

// Test02CreatePoolPostUpgrade tests that a pool can be created.
// It attempts to create a pool with both native and IBC denoms.
// As a result, it must run after Test01IBCTokenTransfer.
// This is the reason for prefixing the name with 02 to ensure
// correct order.
func (s *IntegrationTestSuite) Test02CreatePool() {
	chainA := s.configurer.GetChainConfig(0)
	chainB := s.configurer.GetChainConfig(1)
	chainANode, err := chainA.GetDefaultNode()
	s.NoError(err)
	chainBNode, err := chainB.GetDefaultNode()
	s.NoError(err)

	chainANode.CreatePool("nativeDenomPool.json", initialization.ValidatorWalletName)
	chainBNode.CreatePool("nativeDenomPool.json", initialization.ValidatorWalletName)

	if s.skipIBC {
		s.T().Log("skipping creating pool with IBC denoms because IBC tests are disabled")
		return
	}

	chainANode.CreatePool("ibcDenomPool.json", initialization.ValidatorWalletName)
	chainBNode.CreatePool("ibcDenomPool.json", initialization.ValidatorWalletName)
}

// Test03AddToExistingLockPostUpgrade tests lockups to both regular and superfluid locks.
// Specifically, we ensure addToExistingLock works for both preupgrade and postupgrade locks.
// This must be run after Test02CreatePool, otherwise, in the event we skip upgrade testing
// we will not have any gamm pool assets to lock up.
// This is the reason for prefixing the name with 03 to ensure
// correct order.
func (s *IntegrationTestSuite) Test03AddToExistingLockPostUpgrade() {
	chainA := s.configurer.GetChainConfig(0)
	chainB := s.configurer.GetChainConfig(1)
	chainANode, err := chainA.GetDefaultNode()
	s.NoError(err)
	chainBNode, err := chainB.GetDefaultNode()
	s.NoError(err)
	if s.skipUpgrade {
		// enable superfluid assets
		chainA.EnableSuperfluidAsset("gamm/pool/1")
		chainB.EnableSuperfluidAsset("gamm/pool/1")

		// setup wallets and send gamm tokens to these wallets (both chains)
		lockupWalletAddrA, lockupWalletSuperfluidAddrA := chainANode.CreateWallet(lockupWallet), chainANode.CreateWallet(lockupWalletSuperfluid)
		chainANode.BankSend("10000000000000000000gamm/pool/1", chainA.NodeConfigs[0].PublicAddress, lockupWalletAddrA)
		chainANode.BankSend("10000000000000000000gamm/pool/1", chainA.NodeConfigs[0].PublicAddress, lockupWalletSuperfluidAddrA)
		lockupWalletAddrB, lockupWalletSuperfluidAddrB := chainBNode.CreateWallet(lockupWallet), chainBNode.CreateWallet(lockupWalletSuperfluid)
		chainBNode.BankSend("10000000000000000000gamm/pool/1", chainB.NodeConfigs[0].PublicAddress, lockupWalletAddrB)
		chainBNode.BankSend("10000000000000000000gamm/pool/1", chainB.NodeConfigs[0].PublicAddress, lockupWalletSuperfluidAddrB)

		// test lock and add to existing lock for both regular and superfluid lockups (both chains)
		chainA.LockAndAddToExistingLock(sdk.NewInt(1000000000000000000), "gamm/pool/1", lockupWalletAddrA, lockupWalletSuperfluidAddrA)
		chainB.LockAndAddToExistingLock(sdk.NewInt(1000000000000000000), "gamm/pool/1", lockupWalletAddrB, lockupWalletSuperfluidAddrB)
		return
	}
	// ensure we can add to existing locks and superfluid locks that existed pre upgrade on chainA
	lockupWalletAddrA, lockupWalletSuperfluidAddrA := chainANode.GetWallet("lockup-wallet"), chainANode.GetWallet("lockup-wallet-superfluid")
	chainANode.AddToExistingLock(sdk.NewInt(1000000000000000000), "gamm/pool/1", "240s", lockupWalletAddrA)
	chainANode.AddToExistingLock(sdk.NewInt(1000000000000000000), "gamm/pool/1", "240s", lockupWalletSuperfluidAddrA)

	// setup wallets and send gamm tokens to these wallets on chainB
	lockupWalletAddrB, lockupWalletSuperfluidAddrB := chainBNode.CreateWallet(lockupWallet), chainBNode.CreateWallet(lockupWalletSuperfluid)
	chainBNode.BankSend("10000000000000000000gamm/pool/1", chainB.NodeConfigs[0].PublicAddress, lockupWalletAddrB)
	chainBNode.BankSend("10000000000000000000gamm/pool/1", chainB.NodeConfigs[0].PublicAddress, lockupWalletSuperfluidAddrB)

	// ensure we can add to new locks and superfluid locks on chainB
	chainB.LockAndAddToExistingLock(sdk.NewInt(1000000000000000000), "gamm/pool/1", lockupWalletAddrB, lockupWalletSuperfluidAddrB)
}

// Test04SuperfluidVoting tests that superfluid voting is functioning as expected.
// It does so by doing the following:
//- creating a pool
// - attempting to submit a proposal to enable superfluid voting in that pool
// - voting yes on the proposal from the validator wallet
// - voting no on the proposal from the delegator wallet
// - ensuring that delegator's wallet overwrites the validator's vote
// This test depends on pool creation to function correctly.
// As a result, it is prefixed by 04 to run after Test02CreatePool.
func (s *IntegrationTestSuite) Test04SuperfluidVoting() {
	chainA := s.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	s.NoError(err)

	// create a text prop, deposit and vote yes
	chainANode.SubmitTextProposal("superfluid vote overwrite test", sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.InitialMinDeposit)), false)
	chainA.LatestProposalNumber += 1
	chainANode.DepositProposal(chainA.LatestProposalNumber, false)
	for _, node := range chainA.NodeConfigs {
		node.VoteYesProposal(initialization.ValidatorWalletName, chainA.LatestProposalNumber)
	}

	// set delegator vote to no
	chainANode.VoteNoProposal(lockupWalletSuperfluid, chainA.LatestProposalNumber)

	s.Eventually(
		func() bool {
			noTotal, yesTotal, noWithVetoTotal, abstainTotal, err := chainANode.QueryPropTally(chainA.LatestProposalNumber)
			if err != nil {
				return false
			}
			if abstainTotal.Int64()+noTotal.Int64()+noWithVetoTotal.Int64()+yesTotal.Int64() <= 0 {
				return false
			}
			return true
		},
		1*time.Minute,
		10*time.Millisecond,
		"Osmosis node failed to retrieve prop tally",
	)
	noTotal, _, _, _, _ := chainANode.QueryPropTally(chainA.LatestProposalNumber)
	noTotalFinal, err := strconv.Atoi(noTotal.String())
	s.NoError(err)

	s.Eventually(
		func() bool {
			intAccountBalance, err := chainANode.QueryIntermediaryAccount("gamm/pool/1", chainA.NodeConfigs[1].OperatorAddress)
			s.Require().NoError(err)
			if err != nil {
				return false
			}
			if noTotalFinal != intAccountBalance {
				fmt.Printf("noTotalFinal %v does not match intAccountBalance %v", noTotalFinal, intAccountBalance)
				return false
			}
			return true
		},
		1*time.Minute,
		10*time.Millisecond,
		"superfluid delegation vote overwrite not working as expected",
	)
}

func (s *IntegrationTestSuite) TestStateSync() {
	if s.skipStateSync {
		s.T().Skip()
	}

	chainA := s.configurer.GetChainConfig(0)
	runningNode, err := chainA.GetDefaultNode()
	s.Require().NoError(err)

	persistentPeers := chainA.GetPersistentPeers()

	stateSyncHostPort := fmt.Sprintf("%s:26657", runningNode.Name)
	stateSyncRPCServers := []string{stateSyncHostPort, stateSyncHostPort}

	// get trust height and trust hash.
	trustHeight, err := runningNode.QueryCurrentHeight()
	s.Require().NoError(err)

	trustHash, err := runningNode.QueryHashFromBlock(trustHeight)
	s.Require().NoError(err)

	stateSynchingNodeConfig := &initialization.NodeConfig{
		Name:               "state-sync",
		Pruning:            "default",
		PruningKeepRecent:  "0",
		PruningInterval:    "0",
		SnapshotInterval:   1500,
		SnapshotKeepRecent: 2,
	}

	tempDir, err := os.MkdirTemp("", "osmosis-e2e-statesync-")
	s.Require().NoError(err)

	// configure genesis and config files for the state-synchin node.
	nodeInit, err := initialization.InitSingleNode(
		chainA.Id,
		tempDir,
		filepath.Join(runningNode.ConfigDir, "config", "genesis.json"),
		stateSynchingNodeConfig,
		time.Duration(chainA.VotingPeriod),
		//time.Duration(chainA.ExpeditedVotingPeriod),
		trustHeight,
		trustHash,
		stateSyncRPCServers,
		persistentPeers,
	)
	s.Require().NoError(err)

	stateSynchingNode := chainA.CreateNode(nodeInit)

	// ensure that the running node has snapshots at a height > trustHeight.
	hasSnapshotsAvailable := func(syncInfo coretypes.SyncInfo) bool {
		snapshotHeight := runningNode.SnapshotInterval
		if uint64(syncInfo.LatestBlockHeight) < snapshotHeight {
			s.T().Logf("snapshot height is not reached yet, current (%d), need (%d)", syncInfo.LatestBlockHeight, snapshotHeight)
			return false
		}

		snapshots, err := runningNode.QueryListSnapshots()
		s.Require().NoError(err)

		for _, snapshot := range snapshots {
			if snapshot.Height > uint64(trustHeight) {
				s.T().Log("found state sync snapshot after trust height")
				return true
			}
		}
		s.T().Log("state sync snashot after trust height is not found")
		return false
	}
	runningNode.WaitUntil(hasSnapshotsAvailable)

	// start the state synchin node.
	err = stateSynchingNode.Run()
	s.NoError(err)

	// ensure that the state synching node cathes up to the running node.
	s.Require().Eventually(func() bool {
		stateSyncNodeHeight, err := stateSynchingNode.QueryCurrentHeight()
		s.NoError(err)

		runningNodeHeight, err := runningNode.QueryCurrentHeight()
		s.NoError(err)

		return stateSyncNodeHeight == runningNodeHeight
	},
		3*time.Minute,
		500*time.Millisecond,
	)

	// stop the state synching node.
	err = chainA.RemoveNode(stateSynchingNode.Name)
	s.NoError(err)
}

func (s *IntegrationTestSuite) TestExpeditedProposals() {
	if !s.skipUpgrade {
		s.T().Skip("this can be re-enabled post v12")
	}

	chainA := s.configurer.GetChainConfig(0)
	chainANode, err := chainA.GetDefaultNode()
	s.NoError(err)

	chainANode.SubmitTextProposal("expedited text proposal", sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.InitialMinExpeditedDeposit)), true)
	chainA.LatestProposalNumber += 1
	chainANode.DepositProposal(chainA.LatestProposalNumber, true)
	totalTimeChan := make(chan time.Duration, 1)
	go chainANode.QueryPropStatusTimed(chainA.LatestProposalNumber, "PROPOSAL_STATUS_PASSED", totalTimeChan)
	for _, node := range chainA.NodeConfigs {
		node.VoteYesProposal(initialization.ValidatorWalletName, chainA.LatestProposalNumber)
	}
	// if querying proposal takes longer than timeoutPeriod, stop the goroutine and error
	var elapsed time.Duration
	timeoutPeriod := time.Duration(2 * time.Minute)
	select {
	case elapsed = <-totalTimeChan:
	case <-time.After(timeoutPeriod):
		err := fmt.Errorf("go routine took longer than %s", timeoutPeriod)
		s.Require().NoError(err)
	}

	// compare the time it took to reach pass status to expected expedited voting period
	expeditedVotingPeriodDuration := time.Duration(chainA.ExpeditedVotingPeriod * float32(time.Second))
	timeDelta := elapsed - expeditedVotingPeriodDuration
	// ensure delta is within one second of expected time
	s.Require().Less(timeDelta, 2*time.Second)
	s.T().Logf("expeditedVotingPeriodDuration within one second of expected time: %v", timeDelta)
	close(totalTimeChan)
}
