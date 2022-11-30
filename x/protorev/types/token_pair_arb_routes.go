package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Creates a new TokenPairArbRoutes object
func NewTokenPairArbRoutes(routes []*Route, tokenA, tokenB string) TokenPairArbRoutes {
	return TokenPairArbRoutes{
		ArbRoutes: routes,
		TokenA:    tokenA,
		TokenB:    tokenB,
	}
}

func (tp *TokenPairArbRoutes) Validate() error {
	// Validate that the token pair is valid
	if tp.TokenA == "" || tp.TokenB == "" {
		return sdkerrors.Wrap(ErrInvalidTokenName, "token name cannot be empty")
	}

	// The list cannot be nil
	if tp.ArbRoutes == nil {
		return sdkerrors.Wrap(ErrInvalidRoute, "no routes were entered")
	}

	// Iterate through all of the possible routes for this pool
	for _, route := range tp.ArbRoutes {
		if len(route.Trades) != 3 {
			return sdkerrors.Wrapf(ErrInvalidRoute, "route %s has %d pools, but should have 3", route, len(route.Trades))
		}

		// In denoms must match either osmo or atom
		if route.Trades[0].DenomA != AtomDenomination && route.Trades[0].DenomA != OsmosisDenomination {
			return sdkerrors.Wrapf(ErrInvalidArbDenom, "route has invalid first pool denom: %s", route.Trades[0].DenomA)
		}

		// Out and in denoms must match
		if route.Trades[0].DenomA != route.Trades[2].DenomB {
			return sdkerrors.Wrapf(ErrInvalidRoute, "route has invalid first and last pool denoms: %s -> %s", route.Trades[0].DenomA, route.Trades[2].DenomB)
		}

		uniquePools := make(map[uint64]bool)
		for _, trade := range route.Trades {
			uniquePools[trade.Pool] = true
		}

		// There must be at least three pools hops for it to be a valid route
		if len(uniquePools) != 3 {
			return sdkerrors.Wrapf(ErrInvalidRoute, "the length of the entered cyclic arbitrage route must be exactly three pools: entered number of pools %d", len(uniquePools))
		}
	}

	return nil
}

func NewRoutes(trades []*Trade) Route {
	return Route{
		Trades: trades,
	}
}

func NewTrade(pool uint64, tokenA, tokenB string) Trade {
	return Trade{
		Pool:   pool,
		DenomA: tokenA,
		DenomB: tokenB,
	}
}
