package e2e

import (
	"fmt"
	"strconv"
	"time"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

func (s *IntegrationTestSuite) TestCreatePoolPostUpgrade() {
	if s.skipUpgrade {
		s.T().Skip()
	}

	chainA := s.configurer.GetChainConfig(0)
	s.configurer.CreatePool(chainA, "pool2A.json")
	s.configurer.CreatePool(chainA, "pool2B.json")
}

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	if s.skipIBC {
		s.T().Skip()
	}

	chainA := s.configurer.GetChainConfig(0)
	chainB := s.configurer.GetChainConfig(1)

	s.configurer.SendIBC(chainA, chainB, chainB.ValidatorConfigs[0].PublicAddress, chain.OsmoToken)
	s.configurer.SendIBC(chainB, chainA, chainA.ValidatorConfigs[0].PublicAddress, chain.OsmoToken)
	s.configurer.SendIBC(chainA, chainB, chainB.ValidatorConfigs[0].PublicAddress, chain.StakeToken)
	s.configurer.SendIBC(chainB, chainA, chainA.ValidatorConfigs[0].PublicAddress, chain.StakeToken)
}

func (s *IntegrationTestSuite) TestSuperfluidVoting() {
	chainA := s.configurer.GetChainConfig(0)

	s.configurer.SubmitSuperfluidProposal(chainA, "gamm/pool/1")
	s.configurer.DepositProposal(chainA)
	s.configurer.VoteYesProposal(chainA)
	walletAddr := s.configurer.CreateWallet(chainA, 0, "wallet")
	// send gamm tokens to validator's other wallet (non self-delegation wallet)
	s.configurer.BankSend(chainA, 0, "100000000000000000000gamm/pool/1", chainA.ValidatorConfigs[0].PublicAddress, walletAddr)
	// lock tokens from validator 0 on chain A
	s.configurer.LockTokens(chainA, 0, "100000000000000000000gamm/pool/1", "240s", "wallet")
	// superfluid delegate from validator 0 non self-delegation wallet to validator 1 on chain A
	s.configurer.SuperfluidDelegate(chainA, chainA.ValidatorConfigs[1].OperatorAddress, "wallet")
	// create a text prop, deposit and vote yes
	s.configurer.SubmitTextProposal(chainA, "superfluid vote overwrite test")
	s.configurer.DepositProposal(chainA)
	s.configurer.VoteYesProposal(chainA)
	// set delegator vote to no
	s.configurer.VoteNoProposal(chainA, 0, "wallet")

	sfProposalNumber := strconv.Itoa(chainA.LatestProposalNumber)
	s.Require().Eventually(
		func() bool {
			noTotal, yesTotal, noWithVetoTotal, abstainTotal, err := s.configurer.QueryPropTally(chainA.Id, 0, sfProposalNumber)
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
	noTotal, _, _, _, _ := s.configurer.QueryPropTally(chainA.Id, 0, sfProposalNumber)
	noTotalFinal, err := strconv.Atoi(noTotal.String())
	s.Require().NoError(err)

	s.Require().Eventually(
		func() bool {
			intAccountBalance, err := s.configurer.QueryIntermediaryAccount(chainA.Id, 0, "gamm/pool/1", chainA.ValidatorConfigs[1].OperatorAddress)
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
