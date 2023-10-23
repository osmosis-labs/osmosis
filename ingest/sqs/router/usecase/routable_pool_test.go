package usecase_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	routerusecase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
)

func (s *RouterTestSuite) TestCalculateTokenOutByTokenIn() {

	tests := map[string]struct {
		tokenIn          sdk.Coin
		tokenOutDenom    string
		expectedTokenOut sdk.Coin
		expectError      error
	}{
		"balancer pool - valid calculation": {
			tokenIn:       sdk.NewCoin("foo", sdk.NewInt(100)),
			tokenOutDenom: "bar",
		},

		// TODO: add tests for other pools once supported.
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.Setup()

			balancerPoolID := s.PrepareBalancerPool()
			balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, balancerPoolID)
			s.Require().NoError(err)

			mock := &mockPool{UnderlyingPool: balancerPool}
			routablePool := routerusecase.NewRoutablePool(mock, tc.tokenOutDenom)

			tokenOut, err := routablePool.CalculateTokenOutByTokenIn(tc.tokenIn)

			if tc.expectError != nil {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// We don't check the exact amount because the correctness of calculations is tested
			// at the pool model layer of abstraction. Here, the goal is to make sure that we get
			// a positive amount when the pool is valid.
			s.Require().True(tokenOut.IsPositive())
		})
	}
}
