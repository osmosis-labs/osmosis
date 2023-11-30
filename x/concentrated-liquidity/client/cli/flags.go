package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagPoolId                     = "pool-id"
	FlagPoolIdToTickSpacingRecords = "pool-tick-spacing-records"
	FlagPoolRecords                = "pool-records"
	FlagHookActions                = "hook-actions"
	FlagContractAddressBech32      = "contract-address-bech32"
)

func FlagSetJustPoolId() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Uint64(FlagPoolId, 0, "The id of pool")
	return fs
}
