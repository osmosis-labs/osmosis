package pools_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/pools"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

// Tests no slippage quotes and validation edge cases aroun transmuter pools.
func (s *RoutablePoolTestSuite) TestCalculateTokenOutByTokenIn_Transmuter() {
	defaultAmount := DefaultAmt0
	defaultBalances := sdk.NewCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(ETH, defaultAmount))

	tests := map[string]struct {
		tokenIn           sdk.Coin
		tokenOutDenom     string
		balances          sdk.Coins
		isInvalidPoolType bool
		expectError       error
	}{
		"valid transmuter quote": {
			tokenIn:       sdk.NewCoin(USDC, defaultAmount),
			tokenOutDenom: ETH,
			balances:      defaultBalances,
		},
		"error: token in is larger than balance of token in": {
			tokenIn:       sdk.NewCoin(USDC, defaultAmount.Add(osmomath.OneInt())),
			tokenOutDenom: ETH,
			balances:      defaultBalances,

			expectError: domain.TransmuterInsufficientBalanceError{
				Denom:         USDC,
				BalanceAmount: defaultAmount.String(),
				Amount:        defaultAmount.Add(osmomath.OneInt()).String(),
			},
		},
		"error: token in is larger than balance of token out": {
			tokenIn:       sdk.NewCoin(USDC, defaultAmount),
			tokenOutDenom: ETH,

			// Make token out amount 1 smaller than the default amount
			balances: sdk.NewCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(ETH, defaultAmount.Sub(osmomath.OneInt()))),

			expectError: domain.TransmuterInsufficientBalanceError{
				Denom:         ETH,
				BalanceAmount: defaultAmount.Sub(osmomath.OneInt()).String(),
				Amount:        defaultAmount.String(),
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.Setup()

			cosmwasmPool := s.PrepareCustomTransmuterPool(s.TestAccs[0], []string{tc.tokenIn.Denom, tc.tokenOutDenom})

			poolType := cosmwasmPool.GetType()

			mock := &mocks.MockRoutablePool{ChainPoolModel: cosmwasmPool.AsSerializablePool(), Balances: tc.balances, PoolType: poolType}
			routablePool := pools.NewRoutablePool(mock, tc.tokenOutDenom, noTakerFee)

			// Overwrite pool type for edge case testing
			if tc.isInvalidPoolType {
				mock.PoolType = poolmanagertypes.Concentrated
			}

			tokenOut, err := routablePool.CalculateTokenOutByTokenIn(tc.tokenIn)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)

			// No slippage swaps on success
			s.Require().Equal(tc.tokenIn.Amount, tokenOut.Amount)
		})
	}
}
