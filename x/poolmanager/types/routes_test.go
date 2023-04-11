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

	defaultSingleRouteOneHop = []SwapAmountInSplitRoute{
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

	defaultTwoHopRoutes = []SwapAmountInRoute{
		{
			PoolId:        fooBarPoolId,
			TokenOutDenom: bar,
		},
		{
			PoolId:        barBazPoolId,
			TokenOutDenom: baz,
		},
	}

	defaultSingleRouteTwoHops = SwapAmountInSplitRoute{
		Pools:         defaultTwoHopRoutes,
		TokenInAmount: twentyFiveBaseUnitsAmount,
	}

	defaultSingleRouteThreeHops = SwapAmountInSplitRoute{
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
)

func TestValidateSplitRoutes(t *testing.T) {
	tests := []struct {
		name      string
		routes    []SwapAmountInSplitRoute
		expectErr error
	}{
		{
			name:   "single route one hop",
			routes: defaultSingleRouteOneHop,
		},
		{
			name:   "single route two hops",
			routes: []SwapAmountInSplitRoute{defaultSingleRouteTwoHops},
		},
		{
			name:   "multi route two and three hops",
			routes: []SwapAmountInSplitRoute{defaultSingleRouteTwoHops, defaultSingleRouteThreeHops},
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
				defaultSingleRouteTwoHops,
				defaultSingleRouteTwoHops,
			},
			expectErr: ErrDuplicateRoutesNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSplitRoutes(tt.routes)

			if tt.expectErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
