package e2e

import (
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	chainA := s.networks[0].GetChain()
	chainB := s.networks[1].GetChain()

	// compare coins of reciever pre and post IBC send
	// diff should only be the amount sent
	s.sendIBC(chainA, chainB, chainB.Validators[0].PublicAddress, chain.OsmoToken)
}

func (s *IntegrationTestSuite) TestStateSync() {
	_, err := s.networks[0].RunValidator(3, true)
	s.Require().NoError(err)

	currentChainHeight, err := s.networks[0].GetCurrentHeightFromValidator(0)
	s.Require().NoError(err)

	err = s.networks[0].WaitUntilHeight(3, currentChainHeight)
	s.Require().NoError(err)
}
