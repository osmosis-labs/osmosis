package cli

import (
	flag "github.com/spf13/pflag"
)

// flags for lockup module tx commands
const (
	FlagDuration = "duration"
)

// FlagSetLockTokens returns flags for LockTokens msg builder
func FlagSetLockTokens() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagDuration, "86400s", "The duration token to be locked. e.g. 1h, 1m, 1s, 0.1s")
	return fs
}
