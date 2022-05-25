package e2e

import (
	"fmt"
	"strconv"
	"time"
)

func (s *IntegrationTestSuite) TestSuperfluidVoting() {
	chainA := s.chainConfigs[0].chain
	s.Run("superfluid_vote_chainA", func() {
		s.submitSuperfluidProposal(chainA, "gamm/pool/1")
		s.depositProposal(chainA)
		s.voteProposal(s.chainConfigs[0])
		// send gamm tokens to validator's other wallet (non self-delegation wallet)
		s.sendTx(chainA, 0, "100000000000000000000gamm/pool/1", chainA.Validators[0].PublicAddress, chainA.Validators[0].PublicAddress2)
		// lock tokens from validator 0 on chain A
		s.lockTokens(chainA, 0, "100000000000000000000gamm/pool/1", "240s", "val2")
		// superfluid delegate from validator 0 non self-delegation wallet to validator 1 on chain A
		s.superfluidDelegate(chainA, s.chainConfigs[0].chain.Validators[1].OperAddress, "val2")
		// create a text prop, deposit and vote yes
		s.submitTextProposal(chainA, "superfluid vote overwrite test")
		s.depositProposal(chainA)
		s.voteProposal(s.chainConfigs[0])
		// set delegator vote to no
		s.voteNoProposal(chainA, 0, "val2")

		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[chainA.ChainMeta.Id][0].GetHostPort("1317/tcp"))
		sfProposalNumber := strconv.Itoa(chainA.PropNumber)
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
				intAccountBalance, err := s.queryIntermediaryAccount(chainA, chainAAPIEndpoint, "gamm/pool/1", chainA.Validators[1].OperAddress)
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
