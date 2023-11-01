package usecase_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	routerusecase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

// Test quote logic over a specific pool that is of CFMM type.
// CFMM pools are balancert and stableswap.
func (s *RouterTestSuite) TestCalculateTokenOutByTokenIn_CFMM() {
	tests := map[string]struct {
		tokenIn          sdk.Coin
		tokenOutDenom    string
		poolType         poolmanagertypes.PoolType
		expectedTokenOut sdk.Coin
		expectError      error
	}{
		"balancer pool - valid calculation": {
			tokenIn:       sdk.NewCoin("foo", sdk.NewInt(100)),
			tokenOutDenom: "bar",
			poolType:      poolmanagertypes.Balancer,
		},
		"stableswap pool - valid calculation": {
			tokenIn:       sdk.NewCoin("foo", sdk.NewInt(100)),
			tokenOutDenom: "bar",
			poolType:      poolmanagertypes.Stableswap,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.Setup()

			poolID := s.CreatePoolFromType(tc.poolType)
			pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolID)
			s.Require().NoError(err)

			mock := &mockPool{ChainPoolModel: pool}
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
