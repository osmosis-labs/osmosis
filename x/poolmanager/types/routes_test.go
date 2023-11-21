package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
)

const (
	foo   = "foo"
	bar   = "bar"
	baz   = "baz"
	uosmo = "uosmo"
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

func TestAddEdge(t *testing.T) {
	tests := []struct {
		name   string
		start  string
		end    string
		token  string
		poolID uint64
	}{
		{
			name:   "Test 1",
			start:  "start1",
			end:    "end1",
			token:  "token1",
			poolID: uint64(1),
		},
		{
			name:   "Test 2",
			start:  "start2",
			end:    "end2",
			token:  "token2",
			poolID: uint64(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &RoutingGraph{}
			g.AddEdge(tt.start, tt.end, tt.token, tt.poolID)

			if len(g.Entries) != 1 {
				t.Errorf("Expected 1 entry, got %d", len(g.Entries))
			}

			startEntry := g.Entries[0]
			if startEntry.Key != tt.start {
				t.Errorf("Expected key to be %s, got %s", tt.start, startEntry.Key)
			}

			if len(startEntry.Value.Entries) != 1 {
				t.Errorf("Expected 1 inner entry, got %d", len(startEntry.Value.Entries))
			}

			endEntry := startEntry.Value.Entries[0]
			if endEntry.Key != tt.end {
				t.Errorf("Expected key to be %s, got %s", tt.end, endEntry.Key)
			}

			if len(endEntry.Value.Routes) != 1 {
				t.Errorf("Expected 1 route, got %d", len(endEntry.Value.Routes))
			}

			route := endEntry.Value.Routes[0]
			if route.PoolId != tt.poolID {
				t.Errorf("Expected poolID to be %d, got %d", tt.poolID, route.PoolId)
			}

			if route.Token != tt.token {
				t.Errorf("Expected token to be %s, got %s", tt.token, route.Token)
			}
		})
	}
}
