package cli

import (
	"strconv"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/client/queryproto"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

var customRouterFlagOverride = map[string]string{
	"router": FlagSwapRouteDenoms,
}

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdNumPools)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdEstimateSwapExactAmountIn)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdEstimateSwapExactAmountOut)

	return cmd
}

// GetCmdEstimateSwapExactAmountIn returns estimation of output coin when amount of x token input.
func GetCmdEstimateSwapExactAmountIn() (*osmocli.QueryDescriptor, *queryproto.EstimateSwapExactAmountInRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "estimate-swap-exact-amount-in <poolID> <sender> <tokenIn>",
		Short: "Query estimate-swap-exact-amount-in",
		Long: `Query estimate-swap-exact-amount-in.{{.ExampleHeader}}
{{.CommandPrefix}} estimate-swap-exact-amount-in 1 osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7 1000stake --swap-route-pool-ids=2 --swap-route-pool-ids=3`,
		ParseQuery:          EstimateSwapExactAmountInParseArgs,
		Flags:               osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetMultihopSwapRoutes()}},
		QueryFnName:         "EstimateSwapExactAmountIn",
		CustomFlagOverrides: customRouterFlagOverride,
	}, &queryproto.EstimateSwapExactAmountInRequest{}
}

// GetCmdEstimateSwapExactAmountOut returns estimation of input coin to get exact amount of x token output.
func GetCmdEstimateSwapExactAmountOut() (*osmocli.QueryDescriptor, *queryproto.EstimateSwapExactAmountOutRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "estimate-swap-exact-amount-out <poolID> <sender> <tokenOut>",
		Short: "Query estimate-swap-exact-amount-out",
		Long: `Query estimate-swap-exact-amount-out.{{.ExampleHeader}}
{{.CommandPrefix}} estimate-swap-exact-amount-out 1 osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7 1000stake --swap-route-pool-ids=2 --swap-route-pool-ids=3`,
		ParseQuery:          EstimateSwapExactAmountOutParseArgs,
		Flags:               osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetMultihopSwapRoutes()}},
		QueryFnName:         "EstimateSwapExactAmountOut",
		CustomFlagOverrides: customRouterFlagOverride,
	}, &queryproto.EstimateSwapExactAmountOutRequest{}
}

// GetCmdNumPools return number of pools available.
func GetCmdNumPools() (*osmocli.QueryDescriptor, *queryproto.NumPoolsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "num-pools",
		Short: "Query number of pools",
		Long:  "{{.Short}}",
	}, &queryproto.NumPoolsRequest{}
}

func EstimateSwapExactAmountInParseArgs(args []string, fs *flag.FlagSet) (proto.Message, error) {
	poolID, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, err
	}

	routes, err := swapAmountInRoutes(fs)
	if err != nil {
		return nil, err
	}

	return &queryproto.EstimateSwapExactAmountInRequest{
		Sender:  args[1],        // TODO: where sender is used?
		PoolId:  uint64(poolID), // TODO: is this poolId used?
		TokenIn: args[2],
		Routes:  routes,
	}, nil
}

func EstimateSwapExactAmountOutParseArgs(args []string, fs *flag.FlagSet) (proto.Message, error) {
	poolID, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, err
	}

	routes, err := swapAmountOutRoutes(fs)
	if err != nil {
		return nil, err
	}

	return &queryproto.EstimateSwapExactAmountOutRequest{
		Sender:   args[1],        // TODO: where sender is used?
		PoolId:   uint64(poolID), // TODO: is this poolId used?
		Routes:   routes,
		TokenOut: args[2],
	}, nil
}
