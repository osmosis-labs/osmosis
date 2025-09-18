package types

import "github.com/osmosis-labs/osmosis/osmoutils/observability"

var (
	// takerfee_failed_staking_reward_update
	//
	// counter that is increased if taker fee staking distribution fails
	//
	// Has the following labels:
	// * coins - the coins that fail to be sent.
	// * err - the error occurred
	TakerFeeFailedNativeRewardUpdateMetricName = formatTxFeesMetricName("takerfee_failed_staking_reward_update")
	// txfees_takerfee_failed_community_pool_update
	//
	// counter that is increased if taker fee distribution to community pool fails
	// Has the following labels:
	// * coins - the coins that fail to be sent.
	// * err - the error occurred
	TakerFeeFailedCommunityPoolUpdateMetricName = formatTxFeesMetricName("takerfee_failed_community_pool_update")
	// txfees_takerfee_failed_burn_update
	//
	// counter that is increased if taker fee distribution to burn address fails
	// Has the following labels:
	// * coins - the coins that fail to be burnt.
	// * err - the error occurred
	TakerFeeFailedBurnUpdateMetricName = formatTxFeesMetricName("takerfee_failed_burn_update")
	// txfees_takerfee_swap_failed
	//
	// counter that is increased if taker fee swap to native denom fails
	//
	// Has the following labels:
	// * coin_in - the coin being swapped in.
	// * pool_id - the pool_id being swapped against.
	// * err - the error occurred
	TakerFeeSwapFailedMetricName = formatTxFeesMetricName("takerfee_swap_failed")
	// txfees_takerfee_no_skip_route
	//
	// counter that is increased if taker fee logic fails to find a route to swap between two denoms
	// Has the following labels:
	// * base_denom - the base denom to swap to.
	// * match_denom - the match denom to swap to.
	// * err - the error occurred
	TakerFeeNoSkipRouteMetricName = formatTxFeesMetricName("takerfee_no_skip_route")
)

// formatTxFeesMetricName formats the tx fees module metric name.
func formatTxFeesMetricName(metricName string) string {
	return observability.FormatMetricName(ModuleName, metricName)
}
