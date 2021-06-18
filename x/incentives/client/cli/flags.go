package cli

import (
	"time"

	flag "github.com/spf13/pflag"
)

// flags for lockup module tx commands
const (
	FlagDuration  = "duration"
	FlagStartTime = "start-time"
	FlagEpochs    = "epochs"
	FlagPerpetual = "perpetual"

	FlagTimestamp = "timestamp"
	FlagOwner     = "owner"
	FlagLockIds   = "lock-ids"
	FlagEndEpoch  = "end-epoch"
)

// FlagSetCreateGauge returns flags for creating gauge
func FlagSetCreateGauge() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	dur, _ := time.ParseDuration("168h")
	fs.Duration(FlagDuration, dur, "The duration token to be locked, default 1w(168h). Other examples are 1h, 1m, 1s, 0.1s. Maximum unit is hour.")
	fs.String(FlagStartTime, "", "Timestamp to begin distribution")
	fs.Uint64(FlagEpochs, 0, "Total epochs to distribute tokens")
	fs.Bool(FlagPerpetual, false, "Perpetual distribution")
	return fs
}
