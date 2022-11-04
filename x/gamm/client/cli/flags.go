package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagPoolId = "pool-id"
	// Will be parsed to sdk.Int.
	FlagShareAmountOut = "share-amount-out"
	// Will be parsed to []sdk.Coin.
	FlagMaxAmountsIn = "max-amounts-in"

	// Will be parsed to sdk.Int.
	FlagShareAmountIn = "share-amount-in"
	// Will be parsed to []sdk.Coin.
	FlagMinAmountsOut = "min-amounts-out"

	// Will be parsed to uint64.
	FlagSwapRoutePoolIds = "swap-route-pool-ids"
	// Will be parsed to []sdk.Coin.
	FlagSwapRouteAmounts = "swap-route-amounts"
	// Will be parsed to []string.
	FlagSwapRouteDenoms = "swap-route-denoms"
)

func FlagSetJoinPool() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagPoolId, 0, "The id of pool")
	fs.String(FlagShareAmountOut, "", "Minimum amount of Gamm tokens to receive")
	fs.StringArray(FlagMaxAmountsIn, []string{""}, "Maximum amount of each denom to send into the pool (specify multiple denoms with: --max-amounts-in=1uosmo --max-amounts-in=1uion)")

	return fs
}

func FlagSetExitPool() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagPoolId, 0, "The id of pool")
	fs.String(FlagShareAmountIn, "", "TODO: add description")
	fs.StringArray(FlagMinAmountsOut, []string{""}, "TODO: add description")

	return fs
}

func FlagSetJoinSwapExternAmount() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagPoolId, 0, "The id of pool")

	return fs
}
