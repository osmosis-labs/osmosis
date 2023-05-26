package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	// Will be parsed to string.
	FlagPoolFile = "pool-file"
	FlagPoolType = "pool-type"

	// Names of fields in pool json file.
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
	// FlagScalingFactors represents the flag name for the scaling factors.
	FlagScalingFactors = "scaling-factors"

	FlagMigrationRecords = "migration-records"
)

type createBalancerPoolInputs struct {
	Weights                  string                         `json:"weights"`
	InitialDeposit           string                         `json:"initial-deposit"`
	SwapFee                  string                         `json:"swap-fee"`
	ExitFee                  string                         `json:"exit-fee"`
	FutureGovernor           string                         `json:"future-governor"`
	SmoothWeightChangeParams smoothWeightChangeParamsInputs `json:"lbp-params"`
}

type createStableswapPoolInputs struct {
	InitialDeposit          string `json:"initial-deposit"`
	SwapFee                 string `json:"swap-fee"`
	ExitFee                 string `json:"exit-fee"`
	FutureGovernor          string `json:"future-governor"`
	ScalingFactorController string `json:"scaling-factor-controller"`
	ScalingFactors          string `json:"scaling-factors"`
}

type smoothWeightChangeParamsInputs struct {
	StartTime         string `json:"start-time"`
	Duration          string `json:"duration"`
	TargetPoolWeights string `json:"target-pool-weights"`
}

func FlagSetMultihopSwapRoutes() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagSwapRoutePoolIds, "", "swap route pool id")
	fs.String(FlagSwapRouteDenoms, "", "swap route amount")
	return fs
}

func FlagSetCreatePoolFile() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagPoolFile, "", "Pool json file path (if this path is given, other create pool flags should not be used)")
	return fs
}

func FlagSetCreatePoolType() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagPoolType, "uniswap", "Pool type (either \"balancer\", \"uniswap\", or \"stableswap\"")
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

func FlagSetJustPoolId() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Uint64(FlagPoolId, 0, "The id of pool")
	return fs
}

func FlagSetAdjustScalingFactors() *flag.FlagSet {
	fs := FlagSetJustPoolId()
	fs.String(FlagScalingFactors, "", "The scaling factors")
	return fs
}

func FlagSetMigratePosition() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringArray(FlagMinAmountsOut, []string{""}, "Minimum tokens out")
	return fs
}
