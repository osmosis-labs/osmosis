package usecase

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

type quoteImpl struct {
	AmountIn              sdk.Coin            "json:\"amount_in\""
	AmountOut             osmomath.Int        "json:\"amount_out\""
	Route                 []domain.SplitRoute "json:\"route\""
	EffectiveSpreadFactor osmomath.Dec        "json:\"effective_spread_factor\""
}

// PrepareResult implements domain.Quote.
// PrepareResult mutates the quote to prepare
// it with the data formatted for output to the client.
// Specifically:
// It strips away unnecessary fields from each pool in the route.
// Computes an effective spread factor from all routes.
func (q *quoteImpl) PrepareResult() {

	totalAmountIn := q.AmountIn.Amount.ToLegacyDec()
	totalSpreadFactorAcrossRoutes := osmomath.ZeroDec()

	for _, route := range q.Route {

		routeSpreadFactor := osmomath.ZeroDec()
		routeAmountInFraction := route.GetAmountIn().ToLegacyDec().Quo(totalAmountIn)

		// Calculate the spread factor across pools in the route
		for _, pool := range route.GetPools() {
			spreadFactor := pool.GetSQSPoolModel().SpreadFactor

			routeSpreadFactor = routeSpreadFactor.AddMut(
				osmomath.OneDec().SubMut(routeSpreadFactor).MulTruncateMut(spreadFactor),
			)
		}

		totalSpreadFactorAcrossRoutes = totalSpreadFactorAcrossRoutes.AddMut(routeSpreadFactor.MulMut(routeAmountInFraction))

		route.PrepareResultPools()
	}

	q.EffectiveSpreadFactor = totalSpreadFactorAcrossRoutes
}

// GetAmountIn implements Quote.
func (q *quoteImpl) GetAmountIn() sdk.Coin {
	return q.AmountIn
}

// GetAmountOut implements Quote.
func (q *quoteImpl) GetAmountOut() osmomath.Int {
	return q.AmountOut
}

// GetRoute implements Quote.
func (q *quoteImpl) GetRoute() []domain.SplitRoute {
	return q.Route
}

var _ domain.Quote = &quoteImpl{}
