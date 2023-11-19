package usecase

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

type (
	RouterUseCaseImpl = routerUseCaseImpl

	QuoteImpl = quoteImpl
)

const (
	OsmoPrecisionMultiplier = osmoPrecisionMultiplier
	NoTotalValueLockedError = noTotalValueLockedError
)

func (r Router) FindRoutes(tokenInDenom, tokenOutDenom string, currentRoute domain.Route, poolsUsed []bool, previousTokenOutDenoms []string) ([]domain.Route, error) {
	return r.findRoutes(tokenInDenom, tokenOutDenom, currentRoute, poolsUsed, previousTokenOutDenoms)
}

func (r Router) GetBestSplitRoutesQuote(routes []domain.Route, tokenIn sdk.Coin) (quote domain.Quote, err error) {
	return r.estimateBestSplitRouteQuote(routes, tokenIn)
}

func (r *Router) ValidateAndFilterRoutes(routes []domain.Route, tokenInDenom string) ([]domain.Route, error) {
	return r.validateAndFilterRoutes(routes, tokenInDenom)
}

func (r *routerUseCaseImpl) InitializeRouter(ctx context.Context) (*Router, error) {
	return r.initializeRouter(ctx)
}

func (r *routerUseCaseImpl) HandleRoutes(ctx context.Context, router *Router, tokenInDenom, tokenOutDenom string) ([]domain.Route, error) {
	return r.handleRoutes(ctx, router, tokenInDenom, tokenOutDenom)
}

func (r *Router) GetOptimalQuote(tokenIn sdk.Coin, tokenOutDenom string, routes []domain.Route) (domain.Quote, error) {
	return r.getOptimalQuote(tokenIn, tokenOutDenom, routes)
}
