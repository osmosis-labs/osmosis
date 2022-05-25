package e2e

import (
	"fmt"
	"strconv"
	"time"
)

func (s *IntegrationTestSuite) TestSuperfluidVoting() {
	s.Run("superfluid_vote_chainA", func() {
		s.submitSuperfluidProposal(s.chains[0], "gamm/pool/1")
		s.depositProposal(s.chains[0])
		s.voteProposal(s.chains[0])
		// send gamm tokens to validator's other wallet (non self-delegation wallet)
		s.sendTx(s.chains[0], 0, "100000000000000000000gamm/pool/1", s.chains[0].Validators[0].PublicAddress, s.chains[0].Validators[0].PublicAddress2)
		// lock tokens from validator 0 on chain A
		s.lockTokens(s.chains[0], 0, "100000000000000000000gamm/pool/1", "240s", "val2")
		// superfluid delegate from validator 0 non self-delegation wallet to validator 1 on chain A
		s.superfluidDelegate(s.chains[0], "100000000000000000000gamm/pool/1", s.chains[0].Validators[1].OperAddress, "val2")
		// create a text prop, deposit and vote yes
		s.submitTextProposal(s.chains[0], "superfluid vote overwrite test")
		s.depositProposal(s.chains[0])
		s.voteProposal(s.chains[0])
		// set delegator vote to no
		s.voteNoProposal(s.chains[0], 0, "val2")

		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chains[0].ChainMeta.Id][0].GetHostPort("1317/tcp"))
		sfProposalNumber := strconv.Itoa(s.chains[0].PropNumber)
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
				intAccountBalance, err := s.queryIntermediaryAccount(s.chains[0], chainAAPIEndpoint, "gamm/pool/1", s.chains[0].Validators[1].OperAddress)
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
	})
}
