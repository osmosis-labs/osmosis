package usecase

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

type (
	RouteImpl        = routeImpl
	RoutablePoolImpl = routablePoolImpl
)

func (r Router) FindRoutes(tokenInDenom, tokenOutDenom string, currentRoute domain.Route, poolsUsed []bool, previousTokenOutDenoms []string) ([]domain.Route, error) {
	return r.findRoutes(tokenInDenom, tokenOutDenom, currentRoute, poolsUsed, previousTokenOutDenoms)
}

func (r Router) GetBestSplitRoutesQuote(routes []domain.Route, tokenIn sdk.Coin) (quote domain.Quote, err error) {
	return r.getBestSplitRoutesQuote(routes, tokenIn)
}

func ValidateRoutes(routes []domain.Route, tokenInDenom string) error {
	return validateRoutes(routes, tokenInDenom)
}
