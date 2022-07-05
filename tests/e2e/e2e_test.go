package e2e

import (
	"fmt"
	"strconv"
	"time"

	"github.com/osmosis-labs/osmosis/v10/tests/e2e/initialization"
)

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	chainA := s.chainConfigs[0]
	chainB := s.chainConfigs[1]
	// compare coins of receiver pre and post IBC send
	// diff should only be the amount sent
	s.sendIBC(chainA, chainB, chainB.validators[0].validator.PublicAddress, initialization.OsmoToken)
}

func (s *IntegrationTestSuite) TestSuperfluidVoting() {
<<<<<<< HEAD
	const walletName = "wallet"
=======
	if s.skipUpgrade {
		// TODO: https://github.com/osmosis-labs/osmosis/issues/1843
		s.T().Skip("Superfluid tests are broken when upgrade is skipped. To be fixed in #1843")
	}
	const walletName = "superfluid-wallet"

>>>>>>> 1da14b8b (e2e: refactor initialization with single node logic (#1963))
	chainA := s.chainConfigs[0]
	s.submitSuperfluidProposal(chainA, "gamm/pool/1")
	s.depositProposal(chainA)
	s.voteProposal(chainA)
	walletAddr := s.createWallet(chainA, 0, walletName)
	// send gamm tokens to validator's other wallet (non self-delegation wallet)
	s.sendTx(chainA, 0, "100000000000000000000gamm/pool/1", chainA.validators[0].validator.PublicAddress, walletAddr)
	// lock tokens from validator 0 on chain A
	s.lockTokens(chainA, 0, "100000000000000000000gamm/pool/1", "240s", walletName)
	// superfluid delegate from validator 0 non self-delegation wallet to validator 1 on chain A
	s.superfluidDelegate(chainA, chainA.validators[1].operatorAddress, walletName)
	// create a text prop, deposit and vote yes
	s.submitTextProposal(chainA, "superfluid vote overwrite test")
	s.depositProposal(chainA)
	s.voteProposal(chainA)
	// set delegator vote to no
	s.voteNoProposal(chainA, 0, walletName)

	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[chainA.meta.Id][0].GetHostPort("1317/tcp"))
	sfProposalNumber := strconv.Itoa(chainA.latestProposalNumber)
	s.Require().Eventually(
		func() bool {
			noTotal, yesTotal, noWithVetoTotal, abstainTotal, err := s.queryPropTally(chainAAPIEndpoint, sfProposalNumber)
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
	noTotal, _, _, _, _ := s.queryPropTally(chainAAPIEndpoint, sfProposalNumber)
	noTotalFinal, err := strconv.Atoi(noTotal.String())
	s.Require().NoError(err)

	s.Require().Eventually(
		func() bool {
			intAccountBalance, err := s.queryIntermediaryAccount(chainA, chainAAPIEndpoint, "gamm/pool/1", chainA.validators[1].operatorAddress)
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
