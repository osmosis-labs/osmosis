package usecase

import (
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

	bestSingleRouteQuote, err := r.getBestSingleRouteQuote(routes, tokenIn, tokenOutDenom)
	if err != nil {
		return nil, err
	}

	r.logger.Debug("bestSingleRouteQuote ", zap.Any("bestSingleRouteQuote", bestSingleRouteQuote))

	return bestSingleRouteQuote, nil
}

func (*Router) getBestSingleRouteQuote(routes []domain.Route, tokenIn sdk.Coin, tokenOutDenom string) (quote domain.Quote, err error) {
	var (
		bestRoute     domain.Route
		bestAmountOut = osmomath.ZeroInt()
	)
	for _, route := range routes {
		directRouteTokenOut, err := route.CalculateTokenOutByTokenIn(tokenIn, tokenOutDenom)
		if err != nil {
			return nil, err
		}

		if !directRouteTokenOut.Amount.IsNil() && (bestAmountOut.IsZero() || directRouteTokenOut.Amount.LT(bestAmountOut)) {
			bestRoute = route
			bestAmountOut = directRouteTokenOut.Amount
		}
	}
	return &quoteImpl{
		AmountIn:  tokenIn,
		AmountOut: sdk.NewCoin(tokenOutDenom, bestAmountOut),
		Route:     []domain.Route{bestRoute},
	}, nil
}
