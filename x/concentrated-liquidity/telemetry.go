package concentrated_liquidity

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/hashicorp/go-metrics"

	"github.com/osmosis-labs/osmosis/osmomath"
	types "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

// emitAccumulatorUpdateTelemetry emits telemetry for accumulator updates
// It detects whether an accumulator update does not occur when expected due to truncation or does occur and emits the appropriate telemetry
func emitAccumulatorUpdateTelemetry(truncatedPlaceholder string, rewardsPerUnitOfLiquidity, rewardsTotal osmomath.Dec, poolID uint64, liquidityInAccum osmomath.Dec, extraKeyVals ...string) {
	// If truncation occurs, we emit events to alert us of the issue.
	if rewardsPerUnitOfLiquidity.IsZero() && !rewardsTotal.IsZero() {
		labels := []metrics.Label{
			{
				Name:  "pool_id",
				Value: strconv.FormatUint(poolID, 10),
			},
			{
				Name:  "total_liq",
				Value: liquidityInAccum.String(),
			},
			{
				Name:  "per_unit_liq",
				Value: rewardsPerUnitOfLiquidity.String(),
			},
			{
				Name:  "total_amt",
				Value: rewardsTotal.String(),
			},
		}

		// Append additional labels
		for i := 0; i < len(extraKeyVals); i += 2 {
			// This might skip applying the last label pair if key or value is missing
			if i+1 > len(labels)-1 {
				break
			}

			key := extraKeyVals[i]
			value := extraKeyVals[i+1]

			labels = append(labels, metrics.Label{
				Name:  key,
				Value: value,
			})
		}

		telemetry.IncrCounterWithLabels([]string{truncatedPlaceholder}, 1, labels)
	}
}

// emitIncentiveOverflowTelemetry emits telemetry for incentive overflow in intermediaty calculations
func emitIncentiveOverflowTelemetry(poolID, incentiveRecordID uint64, timeElapsed, emissionRate osmomath.Dec, err error) {
	telemetry.IncrCounterWithLabels([]string{types.IncentiveOverflowTelemetryName}, 1, []metrics.Label{
		{
			Name:  "pool_id",
			Value: strconv.FormatUint(poolID, 10),
		},
		{
			Name:  "incentive_id",
			Value: strconv.FormatUint(incentiveRecordID, 10),
		},
		{
			Name:  "time_elapsed",
			Value: timeElapsed.String(),
		},
		{
			Name:  "emission_rate",
			Value: emissionRate.String(),
		},
		{
			Name:  "error",
			Value: err.Error(),
		},
	})
}
