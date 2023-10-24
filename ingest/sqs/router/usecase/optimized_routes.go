package usecase

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

type quoteImpl struct {
	AmountIn  sdk.Coin
	AmountOut sdk.Coin
	Route     []domain.Route
}

// GetAmountIn implements Quote.
func (q *quoteImpl) GetAmountIn() sdk.Coin {
	return q.AmountIn
}

// GetAmountOut implements Quote.
func (q *quoteImpl) GetAmountOut() sdk.Coin {
	return q.AmountOut
}

// GetRoute implements Quote.
func (q *quoteImpl) GetRoute() []domain.Route {
	return q.Route
}

var _ domain.Quote = &quoteImpl{}

func (r *Router) getQuote(tokenIn sdk.Coin, tokenOutDenom string) (domain.Quote, error) {
	routes, err := r.getCandidateRoutes(tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return nil, err
	}

	r.logger.Debug("routes ", zap.Int("routes_count", len(routes)))

	// Validate the chosen routes.
	if err := validateRoutes(routes, tokenIn.Denom); err != nil {
		return nil, err
	}

	bestSingleRouteQuote, err := r.getBestSingleRouteQuote(routes, tokenIn)
	if err != nil {
		return nil, err
	}

	r.logger.Info("bestSingleRouteQuote ", zap.Any("value", bestSingleRouteQuote))

	bestSplitRouteQuote, err := r.getBestSplitRoutesQuote(routes, tokenIn)
	if err != nil {
		return nil, err
	}

	r.logger.Info("bestSplitRouteQuote ", zap.Any("value", bestSplitRouteQuote))

	// If the split route quote is better than the single route quote, return the split route quote
	if bestSplitRouteQuote.GetAmountOut().Amount.GT(bestSingleRouteQuote.GetAmountOut().Amount) {
		return bestSplitRouteQuote, nil
	}

	// Otherwise return the single route quote
	return bestSingleRouteQuote, nil
}

func (*Router) getBestSingleRouteQuote(routes []domain.Route, tokenIn sdk.Coin) (quote domain.Quote, err error) {
	if len(routes) == 0 {
		return nil, errors.New("no routes were provided")
	}

	var (
		bestRoute         domain.Route
		bestCoinAmountOut = sdk.Coin{}
	)
	for _, route := range routes {
		directRouteTokenOut, err := route.CalculateTokenOutByTokenIn(tokenIn)
		if err != nil {
			return nil, err
		}

		if !directRouteTokenOut.IsNil() && (bestCoinAmountOut.IsNil() || directRouteTokenOut.Amount.LT(bestCoinAmountOut.Amount)) {
			bestRoute = route
			bestCoinAmountOut = directRouteTokenOut
		}
	}
	return &quoteImpl{
		AmountIn:  tokenIn,
		AmountOut: bestCoinAmountOut,
		Route:     []domain.Route{bestRoute},
	}, nil
}

// CONTRACT: all routes must have the same final token out denom
func (r *Router) getBestSplitRoutesQuote(routes []domain.Route, tokenIn sdk.Coin) (quote domain.Quote, err error) {
	if len(routes) == 0 {
		return nil, errors.New("no routes were provided")
	}

	// Validate that all routes have the same final token out denom
	// Compare the current route final token out denom with the previous route final token out denom
	// starting from the second route
	if err := validateRoutes(routes, tokenIn.Denom); err != nil {
		return nil, err
	}

	if len(routes) == 1 {
		return r.getBestSingleRouteQuote(routes, tokenIn)
	}

	bestSplit, err := r.splitRecursive(tokenIn, routes, Split{
		Routes:          []domain.Route{},
		CurrentTotalOut: osmomath.ZeroInt(),
	})
	if err != nil {
		return nil, err
	}

	return &quoteImpl{
		AmountIn: tokenIn,
		// We are guaranteed that at least one route exists at this point.
		// The contract of the method assumes that all routes have the same final token out denom.
		AmountOut: sdk.NewCoin(routes[0].GetTokenOutDenom(), bestSplit.CurrentTotalOut),
		Route:     bestSplit.Routes,
	}, nil
}

// validateRoutes validates all routes. Specifically:
// - all routes have at least one pool.
// - all routes have the same final token out denom.
// - the final token out denom is not the same as the token in denom.
// - intermediary pools in the route do not contain the token in denom or token out denom.
// - the previous pool token out denom is in the current pool.
// - the current pool token out denom is in the current pool.
// Returns error if not. Nil otherwise.
func validateRoutes(routes []domain.Route, tokenInDenom string) error {
	var tokenOutDenom string
	for i, route := range routes {
		currentRoutePools := route.GetPools()
		if len(currentRoutePools) == 0 {
			return NoPoolsInRoute{RouteIndex: i}
		}

		lastPool := route.GetPools()[len(route.GetPools())-1]
		currentRouteTokenOutDenom := lastPool.GetTokenOutDenom()

		// Validate that route pools do not have the token in denom or token out denom
		previousTokenOut := tokenInDenom
		for j, currentPool := range currentRoutePools {
			currentPoolDenoms := currentRoutePools[j].GetPoolDenoms()
			currentPoolTokenOutDenom := currentPool.GetTokenOutDenom()

			// Check that token in denom and token out denom are in the pool
			// Also check that previous token out is in the pool
			foundPreviousTokenOut := false
			foundCurrentTokenOut := false
			for _, denom := range currentPoolDenoms {
				if denom == previousTokenOut {
					foundPreviousTokenOut = true
				}

				if denom == currentPoolTokenOutDenom {
					foundCurrentTokenOut = true
				}

				// Validate that intermediary pools do not contain the token in denom or token out denom
				if j > 0 && j < len(currentRoutePools)-1 {
					if denom == tokenInDenom {
						return RoutePoolWithTokenInDenomError{RouteIndex: i, TokenInDenom: tokenInDenom}
					}

					if denom == currentRouteTokenOutDenom {
						return RoutePoolWithTokenOutDenomError{RouteIndex: i, TokenOutDenom: currentPoolTokenOutDenom}
					}
				}
			}

			// Ensure that the previous pool token out denom is in the current pool.
			if !foundPreviousTokenOut {
				return PreviousTokenOutDenomNotInPoolError{RouteIndex: i, PoolId: currentPool.GetId(), PreviousTokenOutDenom: previousTokenOut}
			}

			// Ensure that the current pool token out denom is in the current pool.
			if !foundCurrentTokenOut {
				return CurrentTokenOutDenomNotInPoolError{RouteIndex: i, PoolId: currentPool.GetId(), CurrentTokenOutDenom: currentPoolTokenOutDenom}
			}

			// Update previous token out denom
			previousTokenOut = currentPoolTokenOutDenom
		}

		if i > 0 {
			// Ensure that all routes have the same final token out denom
			if currentRouteTokenOutDenom != tokenOutDenom {
				return TokenOutMismatchBetweenRoutesError{TokenOutDenomRouteA: tokenOutDenom, TokenOutDenomRouteB: currentRouteTokenOutDenom}
			}
		}

		tokenOutDenom = currentRouteTokenOutDenom
	}

	if tokenOutDenom == tokenInDenom {
		return TokenOutDenomMatchesTokenInDenomError{Denom: tokenOutDenom}
	}
	return nil
}

type RouteWithOutAmount struct {
	domain.Route
	OutAmount osmomath.Int
}

var _ domain.Route = &RouteWithOutAmount{}

type Split struct {
	Routes          []domain.Route
	CurrentTotalOut osmomath.Int
}

func (r *Router) splitRecursive(remainingTokenIn sdk.Coin, remainingRoutes []domain.Route, currentSplit Split) (bestSplit Split, err error) {
	// Base case, we have no more routes to split and we have a valid split
	if len(remainingRoutes) == 0 {
		return currentSplit, nil
	}

	maxSplitIterationsDec := osmomath.NewDec(int64(r.maxSplitIterations))

	currentRoute := remainingRoutes[0]

	r.logger.Debug("currentRoute ", zap.Stringer("currentRoute", currentRoute))

	for i := 0; i < r.maxSplitIterations; i++ {

		currentAmountIn := remainingTokenIn.Amount.ToLegacyDec().Quo(maxSplitIterationsDec).TruncateInt()

		currentTokenOut, err := currentRoute.CalculateTokenOutByTokenIn(sdk.NewCoin(remainingTokenIn.Denom, currentAmountIn))
		if err != nil {
			return Split{}, err
		}

		currentSplitCopy := Split{
			Routes:          make([]domain.Route, len(currentSplit.Routes)),
			CurrentTotalOut: currentSplit.CurrentTotalOut.Add(currentTokenOut.Amount),
		}
		copy(currentSplitCopy.Routes, currentSplit.Routes)

		currentSplitCopy.Routes = append(currentSplitCopy.Routes, RouteWithOutAmount{
			Route:     currentRoute,
			OutAmount: currentTokenOut.Amount,
		})

		remainingTokenIn = sdk.NewCoin(remainingTokenIn.Denom, remainingTokenIn.Amount.Sub(currentAmountIn))

		r.logger.Debug("split", zap.Stringer("token_in", remainingTokenIn), zap.Stringer("token_out", currentTokenOut))

		currentBestSplit, err := r.splitRecursive(remainingTokenIn, remainingRoutes[1:], currentSplitCopy)
		if err != nil {
			return Split{}, err
		}

		if bestSplit.CurrentTotalOut.IsNil() || currentBestSplit.CurrentTotalOut.GT(bestSplit.CurrentTotalOut) {
			bestSplit = currentBestSplit
			r.logger.Debug("selected as best split")
		}

	}

	return bestSplit, nil
}
