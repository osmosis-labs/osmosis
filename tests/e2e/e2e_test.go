package e2e

import (
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

func (s *IntegrationTestSuite) TestCreatePoolPostUpgrade() {
	if s.skipUpgrade {
		s.T().Skip()
	}

	chainA := s.configurer.GetChainConfig(0).GetChain()
	s.configurer.CreatePool(chainA.ChainMeta.Id, 0, "pool2A.json")
	s.configurer.CreatePool(chainA.ChainMeta.Id, 0, "pool2B.json")
}

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	if s.skipIBC {
		s.T().Skip()
	}

	chainA := s.configurer.GetChainConfig(0).GetChain()
	chainB := s.configurer.GetChainConfig(1).GetChain()

	s.configurer.SendIBC(chainA, chainB, chainB.Validators[0].PublicAddress, chain.OsmoToken)
	s.configurer.SendIBC(chainB, chainA, chainA.Validators[0].PublicAddress, chain.OsmoToken)
	s.configurer.SendIBC(chainA, chainB, chainB.Validators[0].PublicAddress, chain.StakeToken)
	s.configurer.SendIBC(chainB, chainA, chainA.Validators[0].PublicAddress, chain.StakeToken)
}
