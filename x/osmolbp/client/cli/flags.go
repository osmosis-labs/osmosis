package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagTokenIn = "token-in"
	FlagTokenOut = "token-out"
	FlagStartTime = "start-time"
	FlagDuration = "duration"
	FlagInitialDeposit = "initial-deposit"
	FlagTreasury = "treasury"
)

func FlagSetCreateLBP() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagTokenIn, "", "")
	fs.String(FlagTokenOut, "", "")
	fs.String(FlagStartTime, "", "")
	fs.String(FlagDuration, "", "")
	fs.String(FlagInitialDeposit, "", "")
	fs.String(FlagTreasury, "", "")

	return fs
}