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
		// support routes of varying length (with the exception of length 1)
		if len(route.Trades) <= 1 {
			return fmt.Errorf("there must be at least two trades per route")
		}

		// Out and in denoms must match
		if route.Trades[0].TokenIn != route.Trades[len(route.Trades)-1].TokenOut {
			return fmt.Errorf("the first and last trades must have matching denoms")
		}

		foundPair := false
		foundPlaceholder := false
		for _, trade := range route.Trades {
			if tp.TokenOut == trade.TokenIn && tp.TokenIn == trade.TokenOut {
				foundPair = true

				if trade.Pool == 0 {
					foundPlaceholder = true
				}
				break
			}
		}

		if !foundPair {
			return fmt.Errorf("the token pair that is going to be arbitraged must appear in each route")
		}

		if !foundPlaceholder {
			return fmt.Errorf("there must be a placeholder pool id of 0 for the token pair that we are arbitraging")
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
