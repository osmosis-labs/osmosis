package usecase

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

type (
	RouteImpl                    = routeImpl
	RoutableCFMMPoolImpl         = routableCFMMPoolImpl
	RoutableConcentratedPoolImpl = routableConcentratedPoolImpl
	RoutableTransmuterPoolImpl   = routableTransmuterPoolImpl
)

const OsmoPrecisionMultiplier = osmoPrecisionMultiplier

func (r Router) FindRoutes(tokenInDenom, tokenOutDenom string, currentRoute domain.Route, poolsUsed []bool, previousTokenOutDenoms []string) ([]domain.Route, error) {
	return r.findRoutes(tokenInDenom, tokenOutDenom, currentRoute, poolsUsed, previousTokenOutDenoms)
}

func (r Router) GetBestSplitRoutesQuote(routes []domain.Route, tokenIn sdk.Coin) (quote domain.Quote, err error) {
	return r.estimateBestSplitRouteQuote(routes, tokenIn)
}

func (r *Router) ValidateAndFilterRoutes(routes []domain.Route, tokenInDenom string) ([]domain.Route, error) {
	return r.validateAndFilterRoutes(routes, tokenInDenom)
}

func (r *Router) GetCandidateRoutes(tokenInDenom, tokenOutDenom string) ([]domain.Route, error) {
	return r.getCandidateRoutes(tokenInDenom, tokenOutDenom)
}
