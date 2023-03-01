package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Creates a new TokenPairArbRoutes object
func NewTokenPairArbRoutes(routes []Route, tokenA, tokenB string) TokenPairArbRoutes {
	return TokenPairArbRoutes{
		ArbRoutes: routes,
		TokenIn:   tokenA,
		TokenOut:  tokenB,
	}
}

func ValidateTokenPairArbRoutes(tokenPairArbRoutes []TokenPairArbRoutes) error {
	if tokenPairArbRoutes == nil {
		return fmt.Errorf("token pair arb routes cannot be nil")
	}

	seenPairs := make(map[string]bool)
	for _, tokenPairArbRoute := range tokenPairArbRoutes {
		if err := tokenPairArbRoute.Validate(); err != nil {
			return err
		}

		// Ensure that the token pair is unique
		pair := tokenPairArbRoute.TokenIn + "/" + tokenPairArbRoute.TokenOut
		if seenPairs[pair] {
			return fmt.Errorf("duplicate token pair %s", pair)
		}
		seenPairs[pair] = true
	}

	return nil
}

func (tp *TokenPairArbRoutes) Validate() error {
	if tp == nil {
		return fmt.Errorf("token pair cannot be nil")
	}

	// Validate that the token pair is valid
	if tp.TokenIn == "" || tp.TokenOut == "" {
		return fmt.Errorf("token names cannot be empty")
	}

	if tp.ArbRoutes == nil || len(tp.ArbRoutes) == 0 {
		return fmt.Errorf("there must be at least one route")
	}

	// Iterate through all of the possible routes for this pool
	for _, route := range tp.ArbRoutes {
		if route.StepSize.IsNil() || route.StepSize.LT(sdk.OneInt()) {
			return fmt.Errorf("step size must be greater than 0")
		}

		// Validate that the route is valid
		if err := isValidRoute(route); err != nil {
			return err
		}

		// Validate that the route has a placeholder pool id for the token pair that we are arbitraging
		if err := hasPlaceholderPool(tp.TokenIn, tp.TokenOut, route.Trades); err != nil {
			return err
		}
	}

	return nil
}

// isValidRoute checks that the route has more than 1 trade, that the first and last trades have matching denoms,
// and that the denoms match across hops
func isValidRoute(route Route) error {
	// support routes of varying length (with the exception of length 1)
	if len(route.Trades) <= 1 {
		return fmt.Errorf("there must be at least two trades (hops) per route")
	}

	// Out and in denoms must match
	if route.Trades[0].TokenIn != route.Trades[len(route.Trades)-1].TokenOut {
		return fmt.Errorf("the first and last trades must have matching denoms")
	}

	// Iterate through all of the trades in the route
	prevDenom := route.Trades[0].TokenOut
	for _, trade := range route.Trades[1:] {
		// Validate that the denoms match
		if prevDenom != trade.TokenIn {
			return fmt.Errorf("the denoms must match across hops")
		}

		// Update the previous denom
		prevDenom = trade.TokenOut
	}

	return nil
}

// hasPlaceholderPool checks that the route has a placeholder pool id (id of 0) for the token pair that we are arbitraging
func hasPlaceholderPool(swapInDenom, swapOutDenom string, trades []Trade) error {
	foundPair := false
	foundPlaceholder := false
	for _, trade := range trades {
		if swapOutDenom == trade.TokenIn && swapInDenom == trade.TokenOut {
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

	return nil
}

func NewRoutes(trades []Trade) Route {
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
