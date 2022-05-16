package e2e

import (
	"github.com/osmosis-labs/osmosis/v8/tests/e2e/chain"
)

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	s.Run("send_uosmo_to_chainB", func() {
		// compare coins of reciever pre and post IBC send
		// diff should only be the amount sent
		s.sendIBC(s.chains[0], s.chains[1], s.chains[1].Validators[0].PublicAddress, chain.OsmoToken)
	})
}
