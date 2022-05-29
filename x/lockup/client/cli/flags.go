package cli

import (
	flag "github.com/spf13/pflag"
)

// flags for lockup module tx commands.
const (
	FlagDuration    = "duration"
	FlagMinDuration = "min-duration"
	FlagAmount      = "amount"
)

// FlagSetLockTokens returns flags for LockTokens msg builder.
func FlagSetLockTokens() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagDuration, "24h", "The duration token to be locked. e.g. 24h, 168h, 336h")
	return fs
}

func FlagSetUnlockTokens() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagAmount, "", "The amount to be unlocked. e.g. 1osmo")
	return fs
}

func FlagSetMinDuration() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagMinDuration, "336h", "The minimum duration of token bonded. e.g. 24h, 168h, 336h")
	return fs
}
