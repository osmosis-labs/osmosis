package cli

import (
	flag "github.com/spf13/pflag"
)

// flags for lockup module tx commands
const (
	FlagLockQueryType = "lock_query_type"
	FlagDenom         = "denom"
	FlagDuration      = "duration"
	FlagTimestamp     = "timestamp"
)

// FlagSetCreatePot returns flags for creating pot
func FlagSetCreatePot() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagLockQueryType, "ByDuration", "ByDuration | ByTime")
	fs.String(FlagDenom, "stake", "locked denom to be queried")
	fs.String(FlagDuration, "86400s", "The duration token to be locked. e.g. 1h, 1m, 1s, 0.1s")
	fs.Int64(FlagTimestamp, 1615917475, "Timestamp to that started tokens lock")
	return fs
}
