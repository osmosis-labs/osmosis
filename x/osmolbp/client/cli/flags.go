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

	FlagPoolId = "pool-id"
	FlagAmount = "amount"
)

func FlagSetCreateLBP() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagTokenIn, "", "denom used to buy LB tokens.")
	fs.String(FlagTokenOut, "", "token denom to be bootstrapped.")
	fs.String(FlagStartTime, "", "when the token sale starts.")
	fs.String(FlagDuration, "", "time that the sale takes place over.")
	fs.String(FlagInitialDeposit, "", "total number of `tokens_out` to be sold during the continuous sale.")
	fs.String(FlagTreasury, "", "account which provides the tokens to sale and receives")

	return fs
}