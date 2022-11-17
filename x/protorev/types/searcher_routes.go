package types

import (
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Creates a new SearcherRoutes object
func NewSearcherRoutes(arbDenom string, routes []*Route) SearcherRoutes {
	return SearcherRoutes{
		ArbDenom: strings.ToUpper(arbDenom),
		Routes:   routes,
	}
}

func (sr *SearcherRoutes) Validate() error {
	// The arb denomination must be tradable
	if sr.ArbDenom != AtomDenomination && sr.ArbDenom != OsmosisDenomination {
		return sdkerrors.Wrapf(ErrInvalidArbDenom, "entered denomination was %s but only %s and %s are allowed", sr.ArbDenom, AtomDenomination, OsmosisDenomination)
	}

	if sr.Routes != nil && len(sr.Routes) == 0 {
		return sdkerrors.Wrap(ErrInvalidRoute, "no routes were entered")
	}

	// Iterate through all of the possible routes for this pool
	for _, route := range sr.Routes {
		uniquePools := make(map[uint64]bool)
		for _, pool := range route.Pools {
			uniquePools[pool] = true
		}

		// There must be at least three pools hit for it to be a valid route
		if len(uniquePools) < 3 {
			return sdkerrors.Wrapf(ErrInvalidRoute, "the length of the entered cyclic arbitrage route must hit at least three pools: entered number of pools %d", len(uniquePools))
		}
	}

	return nil
}
