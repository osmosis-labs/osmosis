package types

import (
	time "time"

	"github.com/tidwall/btree"
)

const (
	ModuleName = "downtime-detector"
	StoreKey   = ModuleName
	RouterKey  = ModuleName

	QuerierRoute = ModuleName
)

var DowntimeToDuration = btree.NewMap[Downtime, time.Duration](16)
var DefaultLastDowntime = time.Unix(0, 0)

// init initializes the DowntimeToDuration map with mappings
// from the Duration enum values to their corresponding
// time.Duration values.
func init() {
	DowntimeToDuration.Set(Downtime_DURATION_30S, 30*time.Second)
	DowntimeToDuration.Set(Downtime_DURATION_1M, time.Minute)
	DowntimeToDuration.Set(Downtime_DURATION_2M, 2*time.Minute)
	DowntimeToDuration.Set(Downtime_DURATION_3M, 3*time.Minute)
	DowntimeToDuration.Set(Downtime_DURATION_4M, 4*time.Minute)
	DowntimeToDuration.Set(Downtime_DURATION_5M, 5*time.Minute)
	DowntimeToDuration.Set(Downtime_DURATION_10M, 10*time.Minute)
	DowntimeToDuration.Set(Downtime_DURATION_20M, 20*time.Minute)
	DowntimeToDuration.Set(Downtime_DURATION_30M, 30*time.Minute)
	DowntimeToDuration.Set(Downtime_DURATION_40M, 40*time.Minute)
	DowntimeToDuration.Set(Downtime_DURATION_50M, 50*time.Minute)
	DowntimeToDuration.Set(Downtime_DURATION_1H, time.Hour)
	DowntimeToDuration.Set(Downtime_DURATION_1_5H, time.Hour+30*time.Minute)
	DowntimeToDuration.Set(Downtime_DURATION_2H, 2*time.Hour)
	DowntimeToDuration.Set(Downtime_DURATION_2_5H, 2*time.Hour+30*time.Minute)
	DowntimeToDuration.Set(Downtime_DURATION_3H, 3*time.Hour)
	DowntimeToDuration.Set(Downtime_DURATION_4H, 4*time.Hour)
	DowntimeToDuration.Set(Downtime_DURATION_5H, 5*time.Hour)
	DowntimeToDuration.Set(Downtime_DURATION_6H, 6*time.Hour)
	DowntimeToDuration.Set(Downtime_DURATION_9H, 9*time.Hour)
	DowntimeToDuration.Set(Downtime_DURATION_12H, 12*time.Hour)
	DowntimeToDuration.Set(Downtime_DURATION_18H, 18*time.Hour)
	DowntimeToDuration.Set(Downtime_DURATION_24H, 24*time.Hour)
	DowntimeToDuration.Set(Downtime_DURATION_36H, 36*time.Hour)
	DowntimeToDuration.Set(Downtime_DURATION_48H, 48*time.Hour)
}
