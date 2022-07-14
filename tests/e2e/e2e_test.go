package e2e

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

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
	node.SubmitSuperfluidProposal("gamm/pool/1")
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
	node.SubmitTextProposal("superfluid vote overwrite test")
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
	s.NoError(err)

	stateSyncHostPort := fmt.Sprintf("%s:26657", runningNode.Name)

	trustHeight, err := runningNode.QueryCurrentHeight()
	s.NoError(err)

	trustHash, err := runningNode.QueryHashFromBlock(trustHeight)
	s.NoError(err)

	stateSyncRPCServers := []string{stateSyncHostPort, stateSyncHostPort}

	stateSynchingNodeConfig := &initialization.NodeConfig{
		Name:               "state-sync",
		Pruning:            "default",
		PruningKeepRecent:  "0",
		PruningInterval:    "0",
		SnapshotInterval:   1500,
		SnapshotKeepRecent: 2,
	}

	nodeInit, err := initialization.InitSingleNode(
		chain.Id,
		chain.DataDir,
		filepath.Join(runningNode.ConfigDir, "config", "genesis.json"),
		stateSynchingNodeConfig,
		time.Duration(chain.VotingPeriod),
		trustHeight,
		trustHash,
		stateSyncRPCServers,
	)
	s.NoError(err)

	stateSynchingNode := chain.CreateNodeConfig(nodeInit)

	// State sync starts here
	err = stateSynchingNode.Run()
	s.NoError(err)

	stateSynchingNode.QueryCurrentHeight()

	s.Require().Eventually(func() bool {

		syncHeight, err := stateSynchingNode.QueryCurrentHeight()
		s.NoError(err)

		runningHeight, err := runningNode.QueryCurrentHeight()
		s.NoError(err)

		return syncHeight == runningHeight
	},
		3*time.Minute,
		2*time.Second,
	)
}
