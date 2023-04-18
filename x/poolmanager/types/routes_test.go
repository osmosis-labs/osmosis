package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

const (
	foo   = "foo"
	bar   = "bar"
	baz   = "baz"
	uosmo = "uosmo"
)

var (
	defaultPoolInitAmount     = sdk.NewInt(10_000_000_000)
	twentyFiveBaseUnitsAmount = sdk.NewInt(25_000_000)

	fooCoin   = sdk.NewCoin(foo, defaultPoolInitAmount)
	barCoin   = sdk.NewCoin(bar, defaultPoolInitAmount)
	bazCoin   = sdk.NewCoin(baz, defaultPoolInitAmount)
	uosmoCoin = sdk.NewCoin(uosmo, defaultPoolInitAmount)

	// Note: These are iniialized in such a way as it makes
	// it easier to reason about the test cases.
	fooBarCoins    = sdk.NewCoins(fooCoin, barCoin)
	fooBarPoolId   = uint64(1)
	fooBazCoins    = sdk.NewCoins(fooCoin, bazCoin)
	fooBazPoolId   = fooBarPoolId + 1
	fooUosmoCoins  = sdk.NewCoins(fooCoin, uosmoCoin)
	fooUosmoPoolId = fooBazPoolId + 1
	barBazCoins    = sdk.NewCoins(barCoin, bazCoin)
	barBazPoolId   = fooUosmoPoolId + 1
	barUosmoCoins  = sdk.NewCoins(barCoin, uosmoCoin)
	barUosmoPoolId = barBazPoolId + 1
	bazUosmoCoins  = sdk.NewCoins(bazCoin, uosmoCoin)
	bazUosmoPoolId = barUosmoPoolId + 1

	// Amount in default routes

	defaultSingleRouteOneHopAmountIn = []SwapAmountInSplitRoute{
		{
			Pools: []SwapAmountInRoute{
				{
					PoolId:        fooBarPoolId,
					TokenOutDenom: bar,
				},
			},
			TokenInAmount: twentyFiveBaseUnitsAmount,
		},
	}

	defaultTwoHopRoutesAmountIn = []SwapAmountInRoute{
		{
			PoolId:        fooBarPoolId,
			TokenOutDenom: bar,
		},
		{
			PoolId:        barBazPoolId,
			TokenOutDenom: baz,
		},
	}

	defaultSingleRouteTwoHopsAmountIn = SwapAmountInSplitRoute{
		Pools:         defaultTwoHopRoutesAmountIn,
		TokenInAmount: twentyFiveBaseUnitsAmount,
	}

	defaultSingleRouteThreeHopsAmountIn = SwapAmountInSplitRoute{
		Pools: []SwapAmountInRoute{
			{
				PoolId:        fooBarPoolId,
				TokenOutDenom: bar,
			},
			{
				PoolId:        barUosmoPoolId,
				TokenOutDenom: uosmo,
			},
			{
				PoolId:        bazUosmoPoolId,
				TokenOutDenom: baz,
			},
		},
		TokenInAmount: sdk.NewInt(twentyFiveBaseUnitsAmount.Int64() * 3),
	}

	// Amount out default routes

	defaultSingleRouteOneHopAmounOut = []SwapAmountOutSplitRoute{
		{
			Pools: []SwapAmountOutRoute{
				{
					PoolId:       fooBarPoolId,
					TokenInDenom: foo,
				},
			},
			TokenOutAmount: twentyFiveBaseUnitsAmount,
		},
	}

	defaultTwoHopRoutesAmountOut = []SwapAmountOutRoute{
		{
			PoolId:       fooBarPoolId,
			TokenInDenom: foo,
		},
		{
			PoolId:       barBazPoolId,
			TokenInDenom: bar,
		},
	}

	defaultSingleRouteTwoHopsAmountOut = SwapAmountOutSplitRoute{
		Pools:          defaultTwoHopRoutesAmountOut,
		TokenOutAmount: twentyFiveBaseUnitsAmount,
	}

	defaultSingleRouteThreeHopsAmountOut = SwapAmountOutSplitRoute{
		Pools: []SwapAmountOutRoute{
			{
				PoolId:       fooBarPoolId,
				TokenInDenom: foo,
			},
			{
				PoolId:       barUosmoPoolId,
				TokenInDenom: bar,
			},
			{
				PoolId:       bazUosmoPoolId,
				TokenInDenom: uosmo,
			},
		},
		TokenOutAmount: sdk.NewInt(twentyFiveBaseUnitsAmount.Int64() * 3),
	}
)

func TestValidateSwapAmountInSplitRoute(t *testing.T) {
	tests := []struct {
		name      string
		routes    []SwapAmountInSplitRoute
		expectErr error
	}{
		{
			name:   "single route one hop",
			routes: defaultSingleRouteOneHopAmountIn,
		},
		{
			name:   "single route two hops",
			routes: []SwapAmountInSplitRoute{defaultSingleRouteTwoHopsAmountIn},
		},
		{
			name:   "multi route two and three hops",
			routes: []SwapAmountInSplitRoute{defaultSingleRouteTwoHopsAmountIn, defaultSingleRouteThreeHopsAmountIn},
		},
		{
			name:      "empty split routes",
			routes:    []SwapAmountInSplitRoute{},
			expectErr: ErrEmptyRoutes,
		},
		{
			name: "empty multihop route",
			routes: []SwapAmountInSplitRoute{
				{
					Pools:         []SwapAmountInRoute{},
					TokenInAmount: sdk.OneInt(),
				},
			},
			expectErr: ErrEmptyRoutes,
		},
		{
			name: "invalid final token out",
			routes: []SwapAmountInSplitRoute{
				{
					Pools: []SwapAmountInRoute{
						{
							PoolId: 1,

							TokenOutDenom: bar,
						},
					},
					TokenInAmount: sdk.OneInt(),
				},
				{
					Pools: []SwapAmountInRoute{
						{
							PoolId:        2,
							TokenOutDenom: baz,
						},
					},
					TokenInAmount: sdk.OneInt(),
				},
			},
			expectErr: InvalidFinalTokenOutError{TokenOutGivenA: bar, TokenOutGivenB: baz},
		},
		{
			name: "duplicate routes",
			routes: []SwapAmountInSplitRoute{
				defaultSingleRouteTwoHopsAmountIn,
				defaultSingleRouteTwoHopsAmountIn,
			},
			expectErr: ErrDuplicateRoutesNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSwapAmountInSplitRoute(tt.routes)

			if tt.expectErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateSwapAmountOutSplitRoute(t *testing.T) {
	tests := []struct {
		name      string
		routes    []SwapAmountOutSplitRoute
		expectErr error
	}{
		{
			name:   "single route one hop",
			routes: defaultSingleRouteOneHopAmounOut,
		},
		{
			name:   "single route two hops",
			routes: []SwapAmountOutSplitRoute{defaultSingleRouteTwoHopsAmountOut},
		},
		{
			name:   "multi route two and three hops",
			routes: []SwapAmountOutSplitRoute{defaultSingleRouteTwoHopsAmountOut, defaultSingleRouteThreeHopsAmountOut},
		},
		{
			name:      "empty split routes",
			routes:    []SwapAmountOutSplitRoute{},
			expectErr: ErrEmptyRoutes,
		},
		{
			name: "empty multihop route",
			routes: []SwapAmountOutSplitRoute{
				{
					Pools:          []SwapAmountOutRoute{},
					TokenOutAmount: sdk.OneInt(),
				},
			},
			expectErr: ErrEmptyRoutes,
		},
		{
			name: "invalid first token in",
			routes: []SwapAmountOutSplitRoute{
				{
					Pools: []SwapAmountOutRoute{
						{
							PoolId: 1,

							TokenInDenom: bar,
						},
					},
					TokenOutAmount: sdk.OneInt(),
				},
				{
					Pools: []SwapAmountOutRoute{
						{
							PoolId:       2,
							TokenInDenom: baz,
						},
					},
					TokenOutAmount: sdk.OneInt(),
				},
			},
			expectErr: InvalidFinalTokenOutError{TokenOutGivenA: bar, TokenOutGivenB: baz},
		},
		{
			name: "duplicate routes",
			routes: []SwapAmountOutSplitRoute{
				defaultSingleRouteTwoHopsAmountOut,
				defaultSingleRouteTwoHopsAmountOut,
			},
			expectErr: ErrDuplicateRoutesNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSwapAmountOutSplitRoute(tt.routes)

			if tt.expectErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
