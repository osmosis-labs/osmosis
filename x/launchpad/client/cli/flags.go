package cli

import (
	flag "github.com/spf13/pflag"
	"time"
)

const (
	FlagSaleId   = "sale-id"
	FlagAmount   = "amount"
	FlagSaleFile = "sale-file"
)

type createSaleInputs struct {
	TokenIn        string    `json:"token-in"`
	TokenOut       string    `json:"token-out"`
	StartTime      time.Time `json:"start-time"`
	Duration       string    `json:"duration"`
	InitialDeposit string    `json:"initial-deposit"`
	Treasury       string    `json:"treasury"`
}

func FlagSetCreateSale() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagSaleFile, "", "Sale json file path")
	return fs
}

func FlagSetFinalizeSale() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Uint64(FlagSaleId, 0, "id of the pool.")

	return fs
}

func FlagSetSubscribe() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagSaleId, 0, "id of the pool.")
	fs.Int64(FlagAmount, 0, "amount to pool.")

	return fs
}

func FlagSetWithdraw() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagSaleId, 0, "id of the pool.")
	fs.Int64(FlagAmount, 0, "amount to pool.")

	return fs
}

func FlagSetExit() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagSaleId, 0, "id of the pool.")

	return fs
}
