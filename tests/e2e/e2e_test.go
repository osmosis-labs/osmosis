package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	appparams "github.com/osmosis-labs/osmosis/v10/app/params"
	"github.com/osmosis-labs/osmosis/v10/tests/e2e/configurer/config"
	"github.com/osmosis-labs/osmosis/v10/tests/e2e/initialization"
)

func (s *IntegrationTestSuite) TestCreatePoolPostUpgrade() {
	if s.skipUpgrade {
		s.T().Skip("pool creation tests are broken when upgrade is skipped. To be fixed in #1843")
	}
	chain := s.configurer.GetChainConfig(0)
	node, err := chain.GetDefaultNode()
	s.NoError(err)

	node.CreatePool("pool2A.json", initialization.ValidatorWalletName)
	node.CreatePool("pool2B.json", initialization.ValidatorWalletName)
}

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
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

func (s *IntegrationTestSuite) TestSuperfluidVoting() {
	if s.skipUpgrade {
		// TODO: https://github.com/osmosis-labs/osmosis/issues/1843
		s.T().Skip("Superfluid tests are broken when upgrade is skipped. To be fixed in #1843")
	}
	const walletName = "superfluid-wallet"

	chain := s.configurer.GetChainConfig(0)
	node, err := chain.GetDefaultNode()
	s.NoError(err)

	// enable superfluid via proposal.
	node.SubmitSuperfluidProposal("gamm/pool/1", sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.InitialMinDeposit)))
	chain.LatestProposalNumber += 1
	node.DepositProposal(chain.LatestProposalNumber)
	for _, node := range chain.NodeConfigs {
		node.VoteYesProposal(initialization.ValidatorWalletName, chain.LatestProposalNumber)
	}

	walletAddr := node.CreateWallet(walletName)
	// send gamm tokens to node's other wallet (non self-delegation wallet)
	node.BankSend("100000000000000000000gamm/pool/1", chain.NodeConfigs[0].PublicAddress, walletAddr)
	// lock tokens from node 0 on chain A
	node.LockTokens("100000000000000000000gamm/pool/1", "240s", walletName)
	chain.LatestLockNumber += 1
	// superfluid delegate from non self-delegation wallet to validator 1 on chain.
	node.SuperfluidDelegate(chain.LatestLockNumber, chain.NodeConfigs[1].OperatorAddress, walletName)

	// create a text prop, deposit and vote yes
	node.SubmitTextProposal("superfluid vote overwrite test", sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.InitialMinDeposit)))
	chain.LatestProposalNumber += 1
	node.DepositProposal(chain.LatestProposalNumber)
	for _, node := range chain.NodeConfigs {
		node.VoteYesProposal(initialization.ValidatorWalletName, chain.LatestProposalNumber)
	}

	// set delegator vote to no
	node.VoteNoProposal(walletName, chain.LatestProposalNumber)

	s.Eventually(
		func() bool {
			noTotal, yesTotal, noWithVetoTotal, abstainTotal, err := node.QueryPropTally(chain.LatestProposalNumber)
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
	noTotal, _, _, _, _ := node.QueryPropTally(chain.LatestProposalNumber)
	noTotalFinal, err := strconv.Atoi(noTotal.String())
	s.NoError(err)

	s.Eventually(
		func() bool {
			intAccountBalance, err := node.QueryIntermediaryAccount("gamm/pool/1", chain.NodeConfigs[1].OperatorAddress)
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

	chain := s.configurer.GetChainConfig(0)
	runningNode, err := chain.GetDefaultNode()
	s.Require().NoError(err)

	persistenrPeers := chain.GetPersistentPeers()

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
		chain.Id,
		tempDir,
		filepath.Join(runningNode.ConfigDir, "config", "genesis.json"),
		stateSynchingNodeConfig,
		time.Duration(chain.VotingPeriod),
		trustHeight,
		trustHash,
		stateSyncRPCServers,
		persistenrPeers,
	)
	s.Require().NoError(err)

	stateSynchingNode := chain.CreateNode(nodeInit)

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
	err = chain.RemoveNode(stateSynchingNode.Name)
	s.NoError(err)
}
