package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v32/app/params"
)

const (
	foo   = "foo"
	bar   = "bar"
	baz   = "baz"
	uosmo = appparams.BaseCoinUnit
)

var (
	twentyFiveBaseUnitsAmount = osmomath.NewInt(25_000_000)

	// Note: These are iniialized in such a way as it makes
	// it easier to reason about the test cases.
	fooBarPoolId   = uint64(1)
	fooBazPoolId   = fooBarPoolId + 1
	fooUosmoPoolId = fooBazPoolId + 1
	barBazPoolId   = fooUosmoPoolId + 1
	barUosmoPoolId = barBazPoolId + 1
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
		TokenInAmount: osmomath.NewInt(twentyFiveBaseUnitsAmount.Int64() * 3),
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
		TokenOutAmount: osmomath.NewInt(twentyFiveBaseUnitsAmount.Int64() * 3),
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
					TokenInAmount: osmomath.OneInt(),
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
					TokenInAmount: osmomath.OneInt(),
				},
				{
					Pools: []SwapAmountInRoute{
						{
							PoolId:        2,
							TokenOutDenom: baz,
						},
					},
					TokenInAmount: osmomath.OneInt(),
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
		{
			name: "non-adjacent duplicate routes",
			routes: []SwapAmountInSplitRoute{
				defaultSingleRouteTwoHopsAmountIn,
				{
					Pools: []SwapAmountInRoute{{
						PoolId:        fooBazPoolId,
						TokenOutDenom: baz,
					}},
					TokenInAmount: twentyFiveBaseUnitsAmount,
				},
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
					TokenOutAmount: osmomath.OneInt(),
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
					TokenOutAmount: osmomath.OneInt(),
				},
				{
					Pools: []SwapAmountOutRoute{
						{
							PoolId:       2,
							TokenInDenom: baz,
						},
					},
					TokenOutAmount: osmomath.OneInt(),
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
		{
			name: "non-adjacent duplicate routes",
			routes: []SwapAmountOutSplitRoute{
				defaultSingleRouteTwoHopsAmountOut,
				{
					Pools: []SwapAmountOutRoute{{
						PoolId:       fooBazPoolId,
						TokenInDenom: foo,
					}},
					TokenOutAmount: twentyFiveBaseUnitsAmount,
				},
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

func TestIntermediateDenoms(t *testing.T) {

	tests := map[string]struct {
		route          SwapAmountInRoutes
		expectedDenoms []string
	}{
		"happy path: one intermediate denom": {
			route: SwapAmountInRoutes([]SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
			}),

			expectedDenoms: []string{bar},
		},
		"multiple intermediate denoms": {
			route: SwapAmountInRoutes([]SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
				{
					PoolId:        5,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        3,
					TokenOutDenom: foo,
				},
			}),

			expectedDenoms: []string{bar, baz, uosmo},
		},
		"no intermediate denoms (single pool)": {
			route: SwapAmountInRoutes([]SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
			}),

			// Note that we expect the function to fail quietly
			expectedDenoms: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualIntermediateDenoms := tc.route.IntermediateDenoms()
			require.Equal(t, tc.expectedDenoms, actualIntermediateDenoms)
		})
	}
}

func TestRoutesKey(t *testing.T) {
	inA := SwapAmountInRoutes{{PoolId: fooBarPoolId, TokenOutDenom: bar}}
	inAB := SwapAmountInRoutes{{PoolId: fooBarPoolId, TokenOutDenom: bar}, {PoolId: barBazPoolId, TokenOutDenom: baz}}
	inBA := SwapAmountInRoutes{{PoolId: barBazPoolId, TokenOutDenom: baz}, {PoolId: fooBarPoolId, TokenOutDenom: bar}}

	require.Equal(t, inAB.Key(), SwapAmountInRoutes{{PoolId: fooBarPoolId, TokenOutDenom: bar}, {PoolId: barBazPoolId, TokenOutDenom: baz}}.Key())
	require.NotEqual(t, inA.Key(), inAB.Key(), "prefix route must have a distinct key")
	require.NotEqual(t, inAB.Key(), inBA.Key(), "hop order must be part of the key")

	// Length prefixing keeps keys distinct when denom bytes could be read as part of an adjacent field.
	split1 := SwapAmountInRoutes{{PoolId: 1, TokenOutDenom: "abc/def"}}
	split2 := SwapAmountInRoutes{{PoolId: 1, TokenOutDenom: "abc"}, {PoolId: 1, TokenOutDenom: "def"}}
	require.NotEqual(t, split1.Key(), split2.Key())

	outAB := SwapAmountOutRoutes{{PoolId: fooBarPoolId, TokenInDenom: foo}, {PoolId: barBazPoolId, TokenInDenom: bar}}
	outBA := SwapAmountOutRoutes{{PoolId: barBazPoolId, TokenInDenom: bar}, {PoolId: fooBarPoolId, TokenInDenom: foo}}
	require.NotEqual(t, outAB.Key(), outBA.Key())
}

func TestContainsDuplicateRoute(t *testing.T) {
	a := SwapAmountInRoutes{{PoolId: fooBarPoolId, TokenOutDenom: bar}}
	b := SwapAmountInRoutes{{PoolId: fooBazPoolId, TokenOutDenom: baz}}
	c := SwapAmountInRoutes{{PoolId: fooUosmoPoolId, TokenOutDenom: uosmo}}

	tests := []struct {
		name   string
		routes []SwapAmountInRoutes
		want   bool
	}{
		{name: "empty", routes: nil, want: false},
		{name: "single", routes: []SwapAmountInRoutes{a}, want: false},
		{name: "unique", routes: []SwapAmountInRoutes{a, b, c}, want: false},
		{name: "adjacent duplicate", routes: []SwapAmountInRoutes{a, a, b}, want: true},
		{name: "non-adjacent duplicate", routes: []SwapAmountInRoutes{a, b, a}, want: true},
		{name: "duplicate last pair", routes: []SwapAmountInRoutes{a, b, b}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, containsDuplicateRoute(tt.routes))
		})
	}
}

// TestValidateSplitRouteLargeInput guards against reintroducing quadratic duplicate
// detection, which runs in ValidateBasic before any gas is charged.
func TestValidateSplitRouteLargeInput(t *testing.T) {
	const numRoutes = 50_000
	routes := make([]SwapAmountInSplitRoute, numRoutes)
	for i := range routes {
		routes[i] = SwapAmountInSplitRoute{
			Pools:         []SwapAmountInRoute{{PoolId: uint64(i + 1), TokenOutDenom: uosmo}},
			TokenInAmount: osmomath.OneInt(),
		}
	}

	start := time.Now()
	require.NoError(t, ValidateSwapAmountInSplitRoute(routes))
	require.Less(t, time.Since(start), 2*time.Second)

	routes[numRoutes-1] = routes[0]
	require.ErrorIs(t, ValidateSwapAmountInSplitRoute(routes), ErrDuplicateRoutesNotAllowed)
}
