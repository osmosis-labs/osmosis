package e2e

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	s.Run("send_uosmo_to_chainB", func() {
		// compare coins of receiver pre and post IBC send
		// diff should only be the amount sent
		s.sendIBC(s.chains[0], s.chains[1], s.chains[1].Validators[0].PublicAddress, chain.OsmoToken)
	})
}

func (s *IntegrationTestSuite) TestSuperfluidVoting() {
	s.Run("superfluid_vote_chainA", func() {
		s.submitSuperfluidProposal(s.chains[0], "gamm/pool/1")
		s.depositProposal(s.chains[0])
		s.voteProposal(s.chains[0])
		//s.lockTokens(s.chains[0], "100000000000000000000gamm/pool/2", "1814400s")
		fmt.Printf("DELADDR %s", s.chains[0].Validators[1].OperAddress)
		s.superfluidDelegate(s.chains[0], "100000000000000000000gamm/pool/1", s.chains[0].Validators[1].OperAddress)
	})
}
