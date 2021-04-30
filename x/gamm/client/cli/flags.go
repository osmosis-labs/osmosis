package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	// Will be parsed to string
	FlagPoolFile = "pool-file"

	// Will be parsed to []sdk.DecCoin
	FlagWeights = "weights"
	// Will be parsed to []sdk.Coin
	FlagInitialDeposit = "initial-deposit"
	// Will be parsed to sdk.Dec
	FlagSwapFee = "swap-fee"
	// Will be parsed to sdk.Dec
	FlagExitFee = "exit-fee"
	// FlagFutureGovernor can be an address, or a This LP Token, lockup time pair
	FlagFutureGovernor = "future-governor"

	FlagPoolId = "pool-id"
	// Will be parsed to sdk.Int
	FlagShareAmountOut = "share-amount-out"
	// Will be parsed to []sdk.Coin
	FlagMaxAmountsIn = "max-amounts-in"

	// Will be parsed to sdk.Int
	FlagShareAmountIn = "share-amount-in"
	// Will be parsed to []sdk.Coin
	FlagMinAmountsOut = "min-amounts-out"

	// Will be parsed to uint64
	FlagSwapRoutePoolIds = "swap-route-pool-ids"
	// Will be parsed to []sdk.Coin
	FlagSwapRouteAmounts = "swap-route-amounts"
	// Will be parsed to []string
	FlagSwapRouteDenoms = "swap-route-denoms"
)

// CreatePoolFlags defines the core required fields of creating a pool. It is used to
// verify that these values are not provided in conjunction with a JSON pool
// file.
var CreatePoolFlags = []string{
	FlagWeights,
	FlagInitialDeposit,
	FlagSwapFee,
	FlagExitFee,
	FlagFutureGovernor,
}

type createPoolInputs struct {
	Weights        string `json:"weights"`
	InitialDeposit string `json:"initial-deposit"`
	SwapFee        string `json:"swap-fee"`
	ExitFee        string `json:"exit-fee"`
	FutureGovernor string `json:"future-governor"`
}

func FlagSetQuerySwapRoutes() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.StringArray(FlagSwapRoutePoolIds, []string{""}, "swap route pool id")
	fs.StringArray(FlagSwapRouteDenoms, []string{""}, "swap route amount")
	return fs
}

func FlagSetSwapAmountOutRoutes() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.StringArray(FlagSwapRoutePoolIds, []string{""}, "swap route pool ids")
	fs.StringArray(FlagSwapRouteDenoms, []string{""}, "swap route denoms")
	return fs
}

func FlagSetCreatePool() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagPoolFile, "", "Pool json file path (if this path is given, other create pool flags should not be used)")
	fs.String(FlagWeights, "", "The amm weights of the tokens in the pool")
	fs.String(FlagInitialDeposit, "", "The tokens to be deposited to the pool initially")
	fs.String(FlagSwapFee, "", "Swap fee of the pool")
	fs.String(FlagExitFee, "", "Exit fee of the pool")
	fs.String(FlagFutureGovernor, "", "Future governor of the pool")
	return fs
}

func FlagSetJoinPool() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagPoolId, 0, "The id of pool")
	fs.String(FlagShareAmountOut, "", "TODO: add description")
	fs.StringArray(FlagMaxAmountsIn, []string{""}, "TODO: add description")

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
