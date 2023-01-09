package cli

import flag "github.com/spf13/pflag"

const (
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

	// Will be parsed to string.
	FlagPoolFile = "pool-file"
	// Will be parsed to uint64.
	FlagSwapRoutePoolIds = "swap-route-pool-ids"
	// Will be parsed to []string.
	FlagSwapRouteDenoms = "swap-route-denoms"
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
