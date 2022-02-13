package cli

import (
	flag "github.com/spf13/pflag"
)

// flags for lockup module tx commands
const (
	FlagDuration    = "duration"
	FlagMinDuration = "min-duration"
)

// FlagSetLockTokens returns flags for LockTokens msg builder
func FlagSetLockTokens() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagDuration, "24h", "The duration token to be locked. e.g. 24h, 168h, 336h")
	return fs
}

func FlagSetMinDuration() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagMinDuration, "336h", "The minimum duration of token bonded. e.g. 24h, 168h, 336h")
	return fs
}
