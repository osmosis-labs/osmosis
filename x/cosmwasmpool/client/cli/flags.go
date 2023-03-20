package cli

import flag "github.com/spf13/pflag"

const (
	//TODO: change this, coode-generated
	FlagSwapRoutePoolIds = "test-flag"
)

func FlagCodeGenerated() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagSwapRoutePoolIds, "", "TODO: change this, coode-generated")
	return fs
}
