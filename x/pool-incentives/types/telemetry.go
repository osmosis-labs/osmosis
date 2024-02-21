package types

import "github.com/osmosis-labs/osmosis/osmoutils/observability"

var (
	// pool_incentives_no_pool_id_for_gauge
	//
	// counter that is increased if no pool ID is found for a given gauge ID and a lockable duration.
	//
	// Has the following labels:
	// * gauge_id - the ID of the gauge.
	// * duration - the lockable duration.
	NoPoolIdForGaugeTelemetryName = formatPoolIncentivesMetricName("no_pool_id_for_gauge")
)

// formatPoolIncentivesMetricName formats the pool incentives module metric name.
func formatPoolIncentivesMetricName(metricName string) string {
	return observability.FormatMetricName(ModuleName, metricName)
}
