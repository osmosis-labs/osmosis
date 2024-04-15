package types

import "github.com/osmosis-labs/osmosis/osmoutils/observability"

var (
	// incentives_group_gauge_sync_failure
	//
	// counter that is increased if group gauge fails to sync
	//
	// Has the following labels:
	// * group_gauge_id - the ID of the group gauge
	// * err - error
	SyncGroupGaugeFailureMetricName = formatIncentivesMetricName("incentives_group_gauge_sync_failure")
)

// formatIncentivesMetricName formats the incentives module metric name.
func formatIncentivesMetricName(metricName string) string {
	return observability.FormatMetricName(ModuleName, metricName)
}
