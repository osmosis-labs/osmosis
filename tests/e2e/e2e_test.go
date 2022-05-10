package e2e

import (
	"time"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	s.Run("send_uosmo_to_chainB", func() {
		// compare coins of reciever pre and post IBC send
		// diff should only be the amount sent
		recipient := s.chains[1].Validators[0].PublicAddress
		balancesBPre, err2 := s.queryBalances(s.valResources[s.chains[1].ChainMeta.Id][0].Container.ID, s.chains[1].Validators[0].PublicAddress)
		s.Require().NoError(err2)
		s.sendIBC(s.chains[0], s.chains[1], recipient, chain.OsmoToken)

		s.Require().Eventually(
			func() bool {
				balancesBPost, err2 := s.queryBalances(s.valResources[s.chains[1].ChainMeta.Id][0].Container.ID, s.chains[1].Validators[0].PublicAddress)
				s.Require().NoError(err2)
				ibcCoin := balancesBPost.Sub(balancesBPre)
				s.Require().True(ibcCoin.Len() == 1)
				tokenPre := balancesBPre.AmountOfNoDenomValidation(ibcCoin[0].Denom)
				tokenPost := balancesBPost.AmountOfNoDenomValidation(ibcCoin[0].Denom)
				resPre := chain.OsmoToken.Amount
				resPost := tokenPost.Sub(tokenPre)
				return resPost.Uint64() == resPre.Uint64()
			},
			time.Minute,
			5*time.Second,
		)
	})
}
