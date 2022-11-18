package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Creates a new SearcherRoutes object
func NewSearcherRoutes(routes []*Route, tokenA, tokenB string) SearcherRoutes {
	// sort tokenA and tokenB
	if tokenA > tokenB {
		tokenA, tokenB = tokenB, tokenA
	}

	return SearcherRoutes{
		Routes: routes,
		TokenA: tokenA,
		TokenB: tokenB,
	}
}

func (sr *SearcherRoutes) Validate() error {
	// Validate that the token pair is valid
	if sr.TokenA == "" || sr.TokenB == "" {
		return sdkerrors.Wrap(ErrInvalidTokenName, "token name cannot be empty")
	}

	// There must be routes within the SearcherRoutes
	if sr.Routes == nil || len(sr.Routes) == 0 {
		return sdkerrors.Wrap(ErrInvalidRoute, "no routes were entered")
	}

	// Iterate through all of the possible routes for this pool
	for _, route := range sr.Routes {
		if len(route.Pools) != 3 {
			return sdkerrors.Wrapf(ErrInvalidRoute, "route %s has %d pools, but should have 3", route, len(route.Pools))
		}

		uniquePools := make(map[uint64]bool)
		for _, pool := range route.Pools {
			uniquePools[pool] = true
		}

		// There must be at least three pools hops for it to be a valid route
		if len(uniquePools) != 3 {
			return sdkerrors.Wrapf(ErrInvalidRoute, "the length of the entered cyclic arbitrage route must hit at least three pools: entered number of pools %d", len(uniquePools))
		}
	}

	return nil
}
