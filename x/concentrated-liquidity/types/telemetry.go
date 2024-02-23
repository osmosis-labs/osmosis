package types

import "github.com/osmosis-labs/osmosis/osmoutils/observability"

var (
	// concentrated_liquidity_incentive_truncation
	//
	// counter that is increased if the incentive accumulator update gets truncated due to division.
	//
	// Has the following labels:
	// * pool_id - the ID of the pool.
	// * total_liq - the liquidity amount in the denominator.
	// * per_unit_liq - the resulting rewards per unit of liquidity (zero if truncated).
	// * total_amt - the total reward amount before dividing by the liquidity value.
	IncentiveTruncationTelemetryName = formatConcentratedMetricName("incentive_truncation")
	// concentrated_liquidity_incentive_overflow
	//
	// counter that is increased if an intermediary math operation in the incentives flow overflows.
	//
	// Has the following labels:
	// * pool_id - the ID of the pool.
	// * incentive_id - the incentive record ID.
	// * time_elapsed - the time elapsed in seconds since the last pool update
	// * emission_rate - the emission rate per second from the incentive record.
	// * error - the error/panic from the failing math operation
	IncentiveOverflowTelemetryName = formatConcentratedMetricName("incentive_overflow")
	// concentrated_liquidity_sptread_factor_truncation
	//
	// counter that is increased if spread factor accumulator update gets truncated due to division by large
	// liquidity value.
	//
	// Has the following labels:
	// * pool_id - the ID of the pool.
	// * incentive_id - the incentive record ID.
	// * time_elapsed - the time elapsed in seconds since the last pool update
	// * emission_rate - the emission rate per second from the incentive record.
	// * error - the error/panic from the failing math operation
	// * is_out_given_in - boolean flag specifying which swap method caused the truncation.
	SpreadFactorTruncationTelemetryName = formatConcentratedMetricName("spread_factor_truncation")
)

// formatConcentratedMetricName formats the concentrated module metric name.
func formatConcentratedMetricName(metricName string) string {
	return observability.FormatMetricName(ModuleName, metricName)
}
