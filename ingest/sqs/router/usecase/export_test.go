package usecase

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/route"
)

type (
	RouterUseCaseImpl = routerUseCaseImpl

	QuoteImpl = quoteImpl
)

const (
	OsmoPrecisionMultiplier = osmoPrecisionMultiplier
	NoTotalValueLockedError = noTotalValueLockedError
)

func (r Router) GetBestSplitRoutesQuote(routes []route.RouteImpl, tickModelMap map[uint64]domain.TickModel, tokenIn sdk.Coin) (quote domain.Quote, err error) {
	return r.estimateBestSplitRouteQuote(routes, tickModelMap, tokenIn)
}

func (r *Router) ValidateAndFilterRoutes(routes []route.RouteImpl, tokenInDenom string) ([]route.RouteImpl, error) {
	return r.validateAndFilterRoutes(routes, tokenInDenom)
}

func (r *routerUseCaseImpl) InitializeRouter(ctx context.Context) (*Router, error) {
	return r.initializeRouter(ctx)
}

func (r *routerUseCaseImpl) HandleRoutes(ctx context.Context, router *Router, tokenInDenom, tokenOutDenom string) ([]route.RouteImpl, error) {
	return r.handleRoutes(ctx, router, tokenInDenom, tokenOutDenom)
}

func (r *Router) GetOptimalQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string, routes []route.RouteImpl, tickModelMap map[uint64]domain.TickModel) (domain.Quote, error) {
	return r.getOptimalQuote(ctx, tokenIn, tokenOutDenom, routes, tickModelMap)
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
