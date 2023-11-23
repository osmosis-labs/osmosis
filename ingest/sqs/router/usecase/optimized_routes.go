package usecase

import (
	"errors"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/route"
)

// getOptimalQuote returns the optimal quote by estimating the optimal route(s) through pools
// Considers all routes and splits.
func (r *Router) getOptimalQuote(tokenIn sdk.Coin, tokenOutDenom string, routes []route.RouteImpl) (domain.Quote, error) {
	for _, route := range routes {
		r.logger.Info("route", zap.Any("route", route))
	}

	bestSingleRouteQuote, err := r.estimateBestSingleRouteQuote(routes, tokenIn)
	if err != nil {
		return nil, err
	}

	r.logger.Info("bestSingleRouteQuote ", zap.Stringer("quote", bestSingleRouteQuote))

	bestSplitRouteQuote, err := r.estimateBestSplitRouteQuote(routes, tokenIn)
	if err != nil {
		return nil, err
	}

	r.logger.Info("bestSplitRouteQuote ", zap.Any("out", bestSingleRouteQuote.GetAmountOut()))

	// If the split route quote is better than the single route quote, return the split route quote
	if bestSplitRouteQuote.GetAmountOut().GT(bestSingleRouteQuote.GetAmountOut()) {
		routes := bestSplitRouteQuote.GetRoute()

		r.logger.Debug("split route is selected", zap.Int("route_count", len(routes)))
		for _, route := range routes {
			r.logger.Debug("route", zap.Stringer("route", route))
		}

		return bestSplitRouteQuote, nil
	}

	r.logger.Debug("single route is selected")
	r.logger.Debug("route", zap.Stringer("route", bestSingleRouteQuote.GetRoute()[0]))

	// Otherwise return the single route quote
	return bestSingleRouteQuote, nil
}

// getSingleRouteQuote returns the best single route quote for the given tokenIn and tokenOutDenom.
func (r *Router) getBestSingleRouteQuote(tokenIn sdk.Coin, tokenOutDenom string, routes []route.RouteImpl) (quote domain.Quote, err error) {
	bestSingleRouteQuote, err := r.estimateBestSingleRouteQuote(routes, tokenIn)
	if err != nil {
		return nil, err
	}

	r.logger.Info("bestSingleRouteQuote ", zap.Any("out", bestSingleRouteQuote.GetAmountOut()))

	return bestSingleRouteQuote, nil
}

func (r *Router) estimateBestSingleRouteQuote(routes []route.RouteImpl, tokenIn sdk.Coin) (quote domain.Quote, err error) {
	if len(routes) == 0 {
		return nil, errors.New("no routes were provided")
	}

	var (
		bestRoute RouteWithOutAmount
	)
	for _, route := range routes {
		directRouteTokenOut, err := route.CalculateTokenOutByTokenIn(tokenIn)
		if err != nil {
			r.logger.Debug("skipping single route due to error in estimate", zap.Error(err))
			continue
		}

		if !directRouteTokenOut.IsNil() && (bestRoute.OutAmount.IsNil() || directRouteTokenOut.Amount.LT(bestRoute.OutAmount)) {
			bestRoute = RouteWithOutAmount{
				RouteImpl: route,
				InAmount:  tokenIn.Amount,
				OutAmount: directRouteTokenOut.Amount,
			}
		}
	}

	if bestRoute.OutAmount.IsNil() {
		return nil, errors.New("did not find a working direct route")
	}

	finalQuote := &quoteImpl{
		AmountIn:  tokenIn,
		AmountOut: bestRoute.OutAmount,
		Route:     []domain.SplitRoute{&bestRoute},
	}

	return finalQuote, nil
}

// CONTRACT: all routes are valid. Must be validated by the caller with validateRoutes method.
func (r *Router) estimateBestSplitRouteQuote(routes []route.RouteImpl, tokenIn sdk.Coin) (quote domain.Quote, err error) {
	if len(routes) == 1 {
		return r.estimateBestSingleRouteQuote(routes, tokenIn)
	}

	r.logger.Debug("estimateBestSplitRoutesQuote", zap.Int("routes_count", len(routes)), zap.Stringer("token_in", tokenIn))
	bestSplit, err := r.splitRecursive(tokenIn, routes, Split{
		Routes:          []domain.SplitRoute{},
		CurrentTotalOut: osmomath.ZeroInt(),
	})
	if err != nil {
		return nil, err
	}

	r.logger.Debug("bestSplit", zap.Any("value", bestSplit))

	finalQuote := &quoteImpl{
		AmountIn:  tokenIn,
		AmountOut: bestSplit.CurrentTotalOut,
		Route:     bestSplit.Routes,
	}

	return finalQuote, nil
}

// validateAndFilterRoutes validates all routes. Specifically:
// - all routes have at least one pool.
// - all routes have the same final token out denom.
// - the final token out denom is not the same as the token in denom.
// - intermediary pools in the route do not contain the token in denom or token out denom.
// - the previous pool token out denom is in the current pool.
// - the current pool token out denom is in the current pool.
// Returns error if not. Nil otherwise.
func (r *Router) validateAndFilterRoutes(routes []route.RouteImpl, tokenInDenom string) ([]route.RouteImpl, error) {
	var (
		tokenOutDenom  string
		filteredRoutes []route.RouteImpl
	)

	uniquePoolIDs := make(map[uint64]struct{})

ROUTE_LOOP:
	for i, route := range routes {
		currentRoutePools := route.GetPools()
		if len(currentRoutePools) == 0 {
			return nil, NoPoolsInRouteError{RouteIndex: i}
		}

		lastPool := route.GetPools()[len(route.GetPools())-1]
		currentRouteTokenOutDenom := lastPool.GetTokenOutDenom()

		// Validate that route pools do not have the token in denom or token out denom
		previousTokenOut := tokenInDenom
		for j, currentPool := range currentRoutePools {
			// Skip routes for which we have already seen the pool ID
			if _, ok := uniquePoolIDs[currentPool.GetId()]; ok {
				continue ROUTE_LOOP
			} else {
				uniquePoolIDs[currentPool.GetId()] = struct{}{}
			}

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
						r.logger.Warn("route skipped - found token in intermediary pool", zap.Error(RoutePoolWithTokenInDenomError{RouteIndex: i, TokenInDenom: tokenInDenom}))
						continue ROUTE_LOOP
					}

					if denom == currentRouteTokenOutDenom {
						r.logger.Warn("route skipped- found token out in intermediary pool", zap.Error(RoutePoolWithTokenOutDenomError{RouteIndex: i, TokenOutDenom: currentPoolTokenOutDenom}))
						continue ROUTE_LOOP
					}
				}
			}

			// Ensure that the previous pool token out denom is in the current pool.
			if !foundPreviousTokenOut {
				return nil, PreviousTokenOutDenomNotInPoolError{RouteIndex: i, PoolId: currentPool.GetId(), PreviousTokenOutDenom: previousTokenOut}
			}

			// Ensure that the current pool token out denom is in the current pool.
			if !foundCurrentTokenOut {
				return nil, CurrentTokenOutDenomNotInPoolError{RouteIndex: i, PoolId: currentPool.GetId(), CurrentTokenOutDenom: currentPoolTokenOutDenom}
			}

			// Update previous token out denom
			previousTokenOut = currentPoolTokenOutDenom
		}

		if i > 0 {
			// Ensure that all routes have the same final token out denom
			if currentRouteTokenOutDenom != tokenOutDenom {
				return nil, TokenOutMismatchBetweenRoutesError{TokenOutDenomRouteA: tokenOutDenom, TokenOutDenomRouteB: currentRouteTokenOutDenom}
			}
		}

		tokenOutDenom = currentRouteTokenOutDenom

		// Update filtered routes if this route passed all checks
		filteredRoutes = append(filteredRoutes, route)
	}

	if tokenOutDenom == tokenInDenom {
		return nil, TokenOutDenomMatchesTokenInDenomError{Denom: tokenOutDenom}
	}

	return filteredRoutes, nil
}

type RouteWithOutAmount struct {
	route.RouteImpl
	OutAmount osmomath.Int "json:\"out_amount\""
	InAmount  osmomath.Int "json:\"in_amount\""
}

var _ domain.SplitRoute = &RouteWithOutAmount{}

// GetAmountIn implements domain.SplitRoute.
func (r RouteWithOutAmount) GetAmountIn() osmomath.Int {
	return r.InAmount
}

// GetAmountOut implements domain.SplitRoute.
func (r RouteWithOutAmount) GetAmountOut() math.Int {
	return r.OutAmount
}

type Split struct {
	Routes          []domain.SplitRoute
	CurrentTotalOut osmomath.Int
}

// splitRecursive recursively splits the token in amount into the best split from the remaining routes.
// It does not perform single route quote estimate (100% single route split) as we assume that those were already calculated prior to this method.
// Returns the best split and error if any.
// Returs error if the maxSplitIterations is less than 1.
func (r *Router) splitRecursive(remainingTokenIn sdk.Coin, remainingRoutes []route.RouteImpl, currentSplit Split) (bestSplit Split, err error) {
	r.logger.Debug("splitRecursive START", zap.Stringer("remainingTokenIn", remainingTokenIn))

	// Base case, we have no more routes to split and we have a valid split
	if len(remainingRoutes) == 0 {
		return currentSplit, nil
	}

	if r.maxSplitIterations <= 1 {
		return Split{}, fmt.Errorf("maxSplitIterations must be greater than 1, was (%d)", r.maxSplitIterations)
	}

	// TODO: this can be precomputed in constructor
	maxSplitIterationsDec := osmomath.NewDec(int64(r.maxSplitIterations))

	currentRoute := remainingRoutes[0]

	r.logger.Debug("currentRoute ", zap.Any("currentRoute", currentRoute))

	for i := 1; i < r.maxSplitIterations; i++ {
		// TODO: this can be precomputed in constructor
		fraction := osmomath.NewDec(int64(i)).Quo(maxSplitIterationsDec)

		// If only the last route is remaining, consume the full remaining amount in
		if len(remainingRoutes) == 1 {
			fraction = osmomath.OneDec()
		}

		// Since the last remaining route is consumed in full, we only need to run the full estimate once
		if len(remainingRoutes) == 1 && i > 1 {
			break
		}

		currentAmountIn := remainingTokenIn.Amount.ToLegacyDec().Mul(fraction).TruncateInt()

		currentTokenOut, err := currentRoute.CalculateTokenOutByTokenIn(sdk.NewCoin(remainingTokenIn.Denom, currentAmountIn))
		if err != nil {
			r.logger.Debug("skipping split due to error in estimate", zap.Error(err))
			continue
		}

		r.logger.Debug("split", zap.Stringer("remaining_token_in", remainingTokenIn), zap.Stringer("fraction", fraction), zap.Stringer("current_token_in", currentAmountIn), zap.Stringer("current_token_out", currentTokenOut), zap.Any("currentRoute", currentRoute))

		currentSplitCopy := Split{
			Routes:          make([]domain.SplitRoute, len(currentSplit.Routes)),
			CurrentTotalOut: currentSplit.CurrentTotalOut.Add(currentTokenOut.Amount),
		}
		copy(currentSplitCopy.Routes, currentSplit.Routes)

		currentSplitCopy.Routes = append(currentSplitCopy.Routes, &RouteWithOutAmount{
			RouteImpl: currentRoute,
			OutAmount: currentTokenOut.Amount,
			InAmount:  currentAmountIn,
		})

		remainingTokenInCopy := sdk.NewCoin(remainingTokenIn.Denom, remainingTokenIn.Amount.Sub(currentAmountIn))

		currentBestSplit, err := r.splitRecursive(remainingTokenInCopy, remainingRoutes[1:], currentSplitCopy)
		if err != nil {
			return Split{}, err
		}

		if bestSplit.CurrentTotalOut.IsNil() || currentBestSplit.CurrentTotalOut.GT(bestSplit.CurrentTotalOut) {
			bestSplit = currentBestSplit
			r.logger.Debug("selected as best split")
		}
	}

	r.logger.Debug("splitRecursive END", zap.Stringer("remainingTokenIn", remainingTokenIn))
	return bestSplit, nil
}
