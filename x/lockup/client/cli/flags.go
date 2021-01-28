package cli

import (
	"time"

	flag "github.com/spf13/pflag"
)

// flags for lockup module tx commands
const (
	FlagDuration = "duration"
)

// FlagSetLockTokens returns flags for LockTokens msg builder
func FlagSetLockTokens() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Int64(FlagDuration, int64(24*time.Hour), "The duration token to be locked")
	return fs
}
