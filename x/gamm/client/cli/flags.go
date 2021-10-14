package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	// Will be parsed to string
	FlagPoolFile = "pool-file"

	// Names of fields in pool json file
	PoolFileWeights        = "weights"
	PoolFileInitialDeposit = "initial-deposit"
	PoolFileSwapFee        = "swap-fee"
	PoolFileExitFee        = "exit-fee"
	PoolFileFutureGovernor = "future-governor"

	PoolFileSmoothWeightChangeParams = "lbp-params"
	PoolFileStartTime                = "start-time"
	PoolFileDuration                 = "duration"
	PoolFileTargetPoolWeights        = "target-pool-weights"

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

type createPoolInputs struct {
	Weights                  string                         `json:"weights"`
	InitialDeposit           string                         `json:"initial-deposit"`
	SwapFee                  string                         `json:"swap-fee"`
	ExitFee                  string                         `json:"exit-fee"`
	FutureGovernor           string                         `json:"future-governor"`
	SmoothWeightChangeParams smoothWeightChangeParamsInputs `json:"lbp-params"`
}

type smoothWeightChangeParamsInputs struct {
	StartTime         string `json:"start-time"`
	Duration          string `json:"duration"`
	TargetPoolWeights string `json:"target-pool-weights"`
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
	return fs
}

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
