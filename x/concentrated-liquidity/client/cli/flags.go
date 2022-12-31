package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagPoolId = "pool-id"
	// Will be parsed to uint64.
	FlagSwapRoutePoolIds = "swap-route-pool-ids"
	// Will be parsed to []string.
	FlagSwapRouteDenoms = "swap-route-denoms"
)

func FlagSetJustPoolId() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Uint64(FlagPoolId, 0, "The id of pool")
	return fs
}

func FlagSetMultihopSwapRoutes() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagSwapRoutePoolIds, "", "swap route pool id")
	fs.String(FlagSwapRouteDenoms, "", "swap route amount")
	return fs
}
