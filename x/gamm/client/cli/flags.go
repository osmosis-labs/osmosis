package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagPoolTokenCustomDenom = "pool-token-custom-denom"
	FlagPoolTokenDescription = "pool-token-description"
	FlagPoolBindTokens       = "bind-tokens"
	FlagPoolBindTokenWeights = "bind-tokens-weight"
	FlagSwapFee              = "swap-fee"

	FlagPoolId = "pool-id"
	// This is string, because it is parsed as sdk.Int
	FlagPoolAmountOut = "pool-amount-out"
	// List of coin
	FlagMaxAountsIn = "pool-max-amounts-in"
)

func FlagSetCreatePool() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagPoolTokenCustomDenom, "", "The custom denom for pool's liquidity token")
	fs.String(FlagPoolTokenDescription, "", "Description of the pool token")
	fs.StringArray(FlagPoolBindTokens, []string{""}, "The tokens to be provided to the pool initially")
	fs.StringArray(FlagPoolBindTokenWeights, []string{""}, "The weights of the tokens in the pool")
	fs.String(FlagSwapFee, "", "Swap fee of the pool")
	return fs
}

func FlagSetJoinPool() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Uint64(FlagPoolId, 0, "The id of pool")
	// This is string, because it is parsed as sdk.Int
	// TODO: 어떻게 설명해야하지...
	fs.String(FlagPoolAmountOut, "", "TODO: add description")
	// TODO: 어떻게 설명해야하지...
	fs.StringArray(FlagMaxAountsIn, []string{""}, "TODO: add description")

	return fs
}
