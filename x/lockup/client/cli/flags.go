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

	fs.String(FlagDuration, "86400s", "The duration token to be locked. e.g. 86400s, 604800s, 1209600s")
	return fs
}

func FlagSetMinDuration() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagMinDuration, "86400s", "The minimum duration of token bonded. e.g. 86400s, 604800s, 1209600s")
	return fs
}
