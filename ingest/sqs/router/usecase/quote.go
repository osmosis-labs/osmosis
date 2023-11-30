package usecase

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
)

type quoteImpl struct {
	AmountIn     sdk.Coin            "json:\"amount_in\""
	AmountOut    osmomath.Int        "json:\"amount_out\""
	Route        []domain.SplitRoute "json:\"route\""
	EffectiveFee osmomath.Dec        "json:\"effective_fee\""
}

// PrepareResult implements domain.Quote.
// PrepareResult mutates the quote to prepare
// it with the data formatted for output to the client.
// Specifically:
// It strips away unnecessary fields from each pool in the route.
// Computes an effective spread factor from all routes.
//
// Returns the updated route and the effective spread factor.
func (q *quoteImpl) PrepareResult() ([]domain.SplitRoute, osmomath.Dec) {
	totalAmountIn := q.AmountIn.Amount.ToLegacyDec()
	totalFeeAcrossRoutes := osmomath.ZeroDec()

	for i, route := range q.Route {
		routeTotalFee := osmomath.ZeroDec()
		routeAmountInFraction := route.GetAmountIn().ToLegacyDec().Quo(totalAmountIn)

		// Calculate the spread factor across pools in the route
		for _, pool := range route.GetPools() {
			poolSpreadFactor := pool.GetSpreadFactor()
			poolTakerFee := pool.GetTakerFee()

			totalPoolFee := poolSpreadFactor.Add(poolTakerFee)

			routeTotalFee.AddMut(
				//  (1 - routeSpreadFactor) * poolSpreadFactor
				osmomath.OneDec().SubMut(routeTotalFee).MulTruncateMut(totalPoolFee),
			)
		}

		// Update the spread factor pro-rated by the amount in
		totalFeeAcrossRoutes.AddMut(routeTotalFee.MulMut(routeAmountInFraction))

		q.Route[i].PrepareResultPools()
	}

	q.EffectiveFee = totalFeeAcrossRoutes

	return q.Route, q.EffectiveFee
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

// GetEffectiveSpreadFactor implements Quote.
func (q *quoteImpl) GetEffectiveSpreadFactor() osmomath.Dec {
	return q.EffectiveFee
}

// String implements domain.Quote.
func (q *quoteImpl) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Quote: %s in for %s out \n", q.AmountIn, q.AmountOut))

	for _, route := range q.Route {
		builder.WriteString(route.String())
	}

	return builder.String()
}

var _ domain.Quote = &quoteImpl{}
