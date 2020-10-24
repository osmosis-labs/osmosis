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
