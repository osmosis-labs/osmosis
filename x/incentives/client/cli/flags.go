package cli

import (
	flag "github.com/spf13/pflag"
)

// Flags for incentives module tx commands.
const (
	FlagStartTime = "start-time"
	FlagEpochs    = "epochs"
	FlagPerpetual = "perpetual"
	FlagTimestamp = "timestamp"
	FlagOwner     = "owner"
	FlagLockIds   = "lock-ids"
	FlagEndEpoch  = "end-epoch"
)

// FlagSetCreateGauge returns flags for creating gauges.
func FlagSetCreateGauge() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagStartTime, "", "Timestamp to begin distribution")
	fs.Uint64(FlagEpochs, 0, "Total epochs to distribute tokens")
	fs.Bool(FlagPerpetual, false, "Perpetual distribution")
	return fs
}
