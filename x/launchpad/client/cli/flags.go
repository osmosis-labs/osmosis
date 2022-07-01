package cli

import (
	"time"

	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/v7/osmoutils/jsontypes"
)

const (
	FlagSaleId   = "sale-id"
	FlagAmount   = "amount"
	FlagSaleFile = "sale-file"
)

type createSaleInputs struct {
	TokenIn   string             `json:"token-in"`
	TokenOut  jsontypes.Coin     `json:"token-out"`
	MaxFee    []jsontypes.Coin   `json:"max-fee"`
	StartTime time.Time          `json:"start-time"`
	Duration  jsontypes.Duration `json:"duration"`
	Recipient string             `json:"recipient"`
	Name      string             `json:"name"`
	Url       string             `json:"url"`
}

func FlagSetCreateSale() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagSaleFile, "", "Sale json file path")
	return fs
}

func FlagSetFinalizeSale() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Uint64(FlagSaleId, 0, "id of the sale.")

	return fs
}

func FlagSetSubscribe() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagSaleId, 0, "id of the sale.")
	fs.Int64(FlagAmount, 0, "amount of sale token_in to deposit for sale.")

	return fs
}

func FlagSetWithdraw() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagSaleId, 0, "id of the sale.")
	fs.Int64(FlagAmount, 0, "amount of sale unspent token_in to withdraw.")

	return fs
}

func FlagSetExit() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagSaleId, 0, "id of the sale.")

	return fs
}
