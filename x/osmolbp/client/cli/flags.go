package cli

import (
	flag "github.com/spf13/pflag"
	"time"
)

const (
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
}
