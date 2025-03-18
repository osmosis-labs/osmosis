package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagPoolId                     = "pool-id"
	FlagPoolIdToTickSpacingRecords = "pool-tick-spacing-records"
)

func FlagSetJustPoolId() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Uint64(FlagPoolId, 0, "The id of pool")
	return fs
}
