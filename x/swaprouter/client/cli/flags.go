package cli

import flag "github.com/spf13/pflag"

const (
	// Will be parsed to uint64.
	FlagSwapRoutePoolIds = "swap-route-pool-ids"
	// Will be parsed to []string.
	FlagSwapRouteDenoms = "swap-route-denoms"
)

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
