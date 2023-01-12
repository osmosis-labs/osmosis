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
		if len(route.Trades) < 3 {
			return fmt.Errorf("there must be at least 3 trades in a route")
		}

		// Out and in denoms must match
		if route.Trades[0].TokenIn != route.Trades[len(route.Trades)-1].TokenOut {
			return fmt.Errorf("the first and last trades must have matching denoms")
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
