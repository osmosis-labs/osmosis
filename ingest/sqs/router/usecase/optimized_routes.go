package usecase

import (
	"errors"
	"sort"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"
)

// getSingleRouteQuote returns the best single route quote for the given tokenIn and tokenOutDenom.
// Returns error if router repository is not set on the router.
func (r *Router) getBestSingleRouteQuote(tokenIn sdk.Coin, routes []route.RouteImpl) (quote domain.Quote, err error) {
	if r.routerRepository == nil {
		return nil, ErrNilRouterRepository
	}
	if r.poolsUsecase == nil {
		return nil, ErrNilPoolsRepository
	}

	bestSingleRouteQuote, _, err := r.estimateBestSingleRouteQuote(routes, tokenIn)
	if err != nil {
		return nil, err
	}

	return bestSingleRouteQuote, nil
}

// Returns best quote as well as all routes sorted by amount out and error if any.
// CONTRACT: router repository must be set on the router.
// CONTRACT: pools reporitory must be set on the router
func (r *Router) estimateBestSingleRouteQuote(routes []route.RouteImpl, tokenIn sdk.Coin) (quote domain.Quote, sortedRoutesByAmtOut []RouteWithOutAmount, err error) {
	if len(routes) == 0 {
		return nil, nil, errors.New("no routes were provided")
	}

	routesWithAmountOut := make([]RouteWithOutAmount, 0, len(routes))

	for _, route := range routes {
		directRouteTokenOut, err := route.CalculateTokenOutByTokenIn(tokenIn)
		if err != nil {
			r.logger.Debug("skipping single route due to error in estimate", zap.Error(err))
			continue
		}

		if directRouteTokenOut.Amount.IsNil() {
			directRouteTokenOut.Amount = osmomath.ZeroInt()
		}

		routesWithAmountOut = append(routesWithAmountOut, RouteWithOutAmount{
			RouteImpl: route,
			InAmount:  tokenIn.Amount,
			OutAmount: directRouteTokenOut.Amount,
		})
	}

	// Sort by amount out in descending order
	sort.Slice(routesWithAmountOut, func(i, j int) bool {
		return routesWithAmountOut[i].OutAmount.GT(routesWithAmountOut[j].OutAmount)
	})

	bestRoute := routesWithAmountOut[0]

	finalQuote := &quoteImpl{
		AmountIn:  tokenIn,
		AmountOut: bestRoute.OutAmount,
		Route:     []domain.SplitRoute{&bestRoute},
	}

	return finalQuote, routesWithAmountOut, nil
}

// validateAndFilterRoutes validates all routes. Specifically:
// - all routes have at least one pool.
// - all routes have the same final token out denom.
// - the final token out denom is not the same as the token in denom.
// - intermediary pools in the route do not contain the token in denom or token out denom.
// - the previous pool token out denom is in the current pool.
// - the current pool token out denom is in the current pool.
// Returns error if not. Nil otherwise.
func (r *Router) validateAndFilterRoutes(candidateRoutes [][]candidatePoolWrapper, tokenInDenom string) (route.CandidateRoutes, error) {
	var (
		tokenOutDenom  string
		filteredRoutes []route.CandidateRoute
	)

	uniquePoolIDs := make(map[uint64]struct{})

ROUTE_LOOP:
	for i, candidateRoute := range candidateRoutes {
		if len(candidateRoute) == 0 {
			return route.CandidateRoutes{}, NoPoolsInRouteError{RouteIndex: i}
		}

		lastPool := candidateRoute[len(candidateRoute)-1]
		currentRouteTokenOutDenom := lastPool.TokenOutDenom

		// Validate that route pools do not have the token in denom or token out denom
		previousTokenOut := tokenInDenom

		uniquePoolIDsIntraRoute := make(map[uint64]struct{}, len(candidateRoute))

		for j, currentPool := range candidateRoute {
			if _, ok := uniquePoolIDs[currentPool.ID]; !ok {
				uniquePoolIDs[currentPool.ID] = struct{}{}
			}

			// Skip routes for which we have already seen the pool ID within that route.
			if _, ok := uniquePoolIDsIntraRoute[currentPool.ID]; ok {
				continue ROUTE_LOOP
			} else {
				uniquePoolIDsIntraRoute[currentPool.ID] = struct{}{}
			}

			currentPoolDenoms := candidateRoute[j].PoolDenoms
			currentPoolTokenOutDenom := currentPool.TokenOutDenom

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
				if j > 0 && j < len(candidateRoute)-1 {
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
				return route.CandidateRoutes{}, PreviousTokenOutDenomNotInPoolError{RouteIndex: i, PoolId: currentPool.ID, PreviousTokenOutDenom: previousTokenOut}
			}

			// Ensure that the current pool token out denom is in the current pool.
			if !foundCurrentTokenOut {
				return route.CandidateRoutes{}, CurrentTokenOutDenomNotInPoolError{RouteIndex: i, PoolId: currentPool.ID, CurrentTokenOutDenom: currentPoolTokenOutDenom}
			}

			// Update previous token out denom
			previousTokenOut = currentPoolTokenOutDenom
		}

		if i > 0 {
			// Ensure that all routes have the same final token out denom
			if currentRouteTokenOutDenom != tokenOutDenom {
				return route.CandidateRoutes{}, TokenOutMismatchBetweenRoutesError{TokenOutDenomRouteA: tokenOutDenom, TokenOutDenomRouteB: currentRouteTokenOutDenom}
			}
		}

		tokenOutDenom = currentRouteTokenOutDenom

		// Update filtered routes if this route passed all checks
		filteredRoute := route.CandidateRoute{
			Pools: make([]route.CandidatePool, 0, len(candidateRoute)),
		}

		// Convert route to the final output format
		for _, pool := range candidateRoute {
			filteredRoute.Pools = append(filteredRoute.Pools, route.CandidatePool{
				ID:            pool.ID,
				TokenOutDenom: pool.TokenOutDenom,
			})
		}

		filteredRoutes = append(filteredRoutes, filteredRoute)
	}

	if tokenOutDenom == tokenInDenom {
		return route.CandidateRoutes{}, TokenOutDenomMatchesTokenInDenomError{Denom: tokenOutDenom}
	}

	return route.CandidateRoutes{
		Routes:        filteredRoutes,
		UniquePoolIDs: uniquePoolIDs,
	}, nil
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
