package usecase

import "github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"

type RouteImpl = routeImpl

func (r Router) FindRoutes(tokenInDenom, tokenOutDenom string, currentRoute domain.Route, poolsUsed []bool, previousTokenOutDenoms []string) ([]domain.Route, error) {
	return r.findRoutes(tokenInDenom, tokenOutDenom, currentRoute, poolsUsed, previousTokenOutDenoms)
}
