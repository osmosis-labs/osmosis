package e2e

import (
	"fmt"
	"strconv"
	"time"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/initialization"
)

func (s *IntegrationTestSuite) TestCreatePoolPostUpgrade() {
	if s.skipUpgrade {
		s.T().Skip("pool creation tests are broken when upgrade is skipped. To be fixed in #1843")
	}

	chainA := s.configurer.GetChainConfig(0)
	chainA.CreatePool("pool2A.json", initialization.ValidatorWalletName)
	chainA.CreatePool("pool2B.json", initialization.ValidatorWalletName)
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

	chainA := s.configurer.GetChainConfig(0)

	chainA.SubmitSuperfluidProposal("gamm/pool/1")

	chainA.DepositProposal()
	chainA.VoteYesProposal()
	walletAddr := chainA.CreateWallet(0, walletName)
	// send gamm tokens to validator's other wallet (non self-delegation wallet)
	chainA.BankSend(0, "100000000000000000000gamm/pool/1", chainA.NodeConfigs[0].PublicAddress, walletAddr)
	// lock tokens from validator 0 on chain A
	chainA.LockTokens(0, "100000000000000000000gamm/pool/1", "240s", walletName)
	// superfluid delegate from validator 0 non self-delegation wallet to validator 1 on chain A
	chainA.SuperfluidDelegate(chainA.NodeConfigs[1].OperatorAddress, walletName)
	// create a text prop, deposit and vote yes
	chainA.SubmitTextProposal("superfluid vote overwrite test")
	chainA.DepositProposal()
	chainA.VoteYesProposal()
	// set delegator vote to no
	chainA.VoteNoProposal(0, walletName)

	sfProposalNumber := strconv.Itoa(chainA.LatestProposalNumber)
	s.Eventually(
		func() bool {
			noTotal, yesTotal, noWithVetoTotal, abstainTotal, err := chainA.QueryPropTally(0, sfProposalNumber)
			if err != nil {
				return false
			}
			if abstainTotal.Int64()+noTotal.Int64()+noWithVetoTotal.Int64()+yesTotal.Int64() <= 0 {
				return false
			}
			return true
		},
		1*time.Minute,
		time.Second,
		"Osmosis node failed to retrieve prop tally",
	)
	noTotal, _, _, _, _ := chainA.QueryPropTally(0, sfProposalNumber)
	noTotalFinal, err := strconv.Atoi(noTotal.String())
	s.NoError(err)

	s.Eventually(
		func() bool {
			intAccountBalance, err := chainA.QueryIntermediaryAccount(0, "gamm/pool/1", chainA.NodeConfigs[1].OperatorAddress)
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
		time.Second,
		"superfluid delegation vote overwrite not working as expected",
	)
}
