package e2e

import (
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	chainA := s.chainConfigs[0].chain
	chainB := s.chainConfigs[1].chain

	// compare coins of reciever pre and post IBC send
	// diff should only be the amount sent
	s.sendIBC(chainA, chainB, chainB.Nodes[0].PublicAddress, chain.OsmoToken)
}
