package types

import (
	"fmt"
)

// Creates a new TokenPairArbRoutes object
func NewTokenPairArbRoutes(routes []*Route, tokenA, tokenB string) TokenPairArbRoutes {
	return TokenPairArbRoutes{
		ArbRoutes: routes,
		TokenIn:   tokenA,
		TokenOut:  tokenB,
	}
}

func (tp *TokenPairArbRoutes) Validate() error {
	// Validate that the token pair is valid
	if tp.TokenIn == "" || tp.TokenOut == "" {
		return fmt.Errorf("token names cannot be empty")
	}

	// The list cannot be nil
	if tp.ArbRoutes == nil {
		return fmt.Errorf("the list of routes cannot be nil")
	}

	// Iterate through all of the possible routes for this pool
	for _, route := range tp.ArbRoutes {
		if len(route.Trades) != 3 {
			return fmt.Errorf("there must be exactly 3 trades in a route")
		}

		// In denoms must match either osmo or atom
		if route.Trades[0].TokenIn != AtomDenomination && route.Trades[0].TokenIn != OsmosisDenomination {
			return fmt.Errorf("the first trade must have either osmo or atom as the in denom")
		}

		// Out and in denoms must match
		if route.Trades[0].TokenIn != route.Trades[2].TokenOut {
			return fmt.Errorf("the first and last trades must have matching denoms")
		}

		uniquePools := make(map[uint64]bool)
		for _, trade := range route.Trades {
			uniquePools[trade.Pool] = true
		}

		// There must be at least three pools hops for it to be a valid route
		if len(uniquePools) != 3 {
			return fmt.Errorf("There must be exactly 3 unique pools in a route")
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
		Pool:     pool,
		TokenIn:  tokenA,
		TokenOut: tokenB,
	}
}
