package usecase

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"
)

type (
	RouterUseCaseImpl = routerUseCaseImpl

	QuoteImpl = quoteImpl

	CandidatePoolWrapper = candidatePoolWrapper
)

const (
	OsmoPrecisionMultiplier = osmoPrecisionMultiplier
	NoTotalValueLockedError = noTotalValueLockedError
)

func (r *Router) ValidateAndFilterRoutes(candidateRoutes [][]candidatePoolWrapper, tokenInDenom string) (route.CandidateRoutes, error) {
	return r.validateAndFilterRoutes(candidateRoutes, tokenInDenom)
}

func (r *routerUseCaseImpl) InitializeRouter() *Router {
	return r.initializeRouter()
}

func (r *routerUseCaseImpl) HandleRoutes(ctx context.Context, router *Router, tokenInDenom, tokenOutDenom string) (candidateRoutes route.CandidateRoutes, err error) {
	return r.handleCandidateRoutes(ctx, router, tokenInDenom, tokenOutDenom)
}

func (r *Router) EstimateBestSingleRouteQuote(routes []route.RouteImpl, tokenIn sdk.Coin) (domain.Quote, []RouteWithOutAmount, error) {
	return r.estimateBestSingleRouteQuote(routes, tokenIn)
}

// GetSortedPoolIDs returns the sorted pool IDs.
// The sorting is initialized in NewRouter() by preferredPoolIDs and TVL.
// Only used for tests.
func (r Router) GetSortedPoolIDs() []uint64 {
	sortedPoolIDs := make([]uint64, len(r.sortedPools))
	for i, pool := range r.sortedPools {
		sortedPoolIDs[i] = pool.GetId()
	}
	return sortedPoolIDs
}
