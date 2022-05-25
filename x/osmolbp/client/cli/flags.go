package cli

import (
	flag "github.com/spf13/pflag"
<<<<<<< HEAD
	"time"
)

const (
	FlagTokenIn        = "token-in"
	FlagTokenOut       = "token-out"
	FlagStartTime      = "start-time"
	FlagDuration       = "duration"
	FlagInitialDeposit = "initial-deposit"
	FlagTreasury       = "treasury"

	FlagPoolId  = "pool-id"
	FlagAmount  = "amount"
	FlagLBPFile = "lbp-file"
)

type createLBPInputs struct {
	TokenIn        string    `json:"token-in"`
	TokenOut       string    `json:"token-out"`
	StartTime      time.Time `json:"start-time"`
	Duration       string    `json:"duration"`
	InitialDeposit string    `json:"initial-deposit"`
	Treasury       string    `json:"treasury"`
}

func FlagSetCreateLBP() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagLBPFile, "", "LBP json file path (if this path is given, other create lbp flags should not be used)")
	return fs
}

func FlagSetFinalizeLBP() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagPoolId, 0, "id of the pool.")
=======
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
>>>>>>> upstream/osmolbp

	return fs
}

func FlagSetSubscribe() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagPoolId, 0, "id of the pool.")
	fs.Int64(FlagAmount, 0, "amount to pool.")

	return fs
}

func FlagSetWithdraw() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagPoolId, 0, "id of the pool.")
	fs.Int64(FlagAmount, 0, "amount to pool.")

	return fs
}

func FlagSetExit() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagPoolId, 0, "id of the pool.")

	return fs
<<<<<<< HEAD
}
=======
}
>>>>>>> upstream/osmolbp
