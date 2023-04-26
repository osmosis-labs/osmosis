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
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdEstimateSinglePoolSwapExactAmountIn)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdEstimateSinglePoolSwapExactAmountOut)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdSpotPrice)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdTotalPoolLiquidity)

	return cmd
}

// GetCmdEstimateSwapExactAmountIn returns estimation of output coin when amount of x token input.
func GetCmdEstimateSwapExactAmountIn() (*osmocli.QueryDescriptor, *queryproto.EstimateSwapExactAmountInRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "estimate-swap-exact-amount-in <poolID> <tokenIn>",
		Short: "Query estimate-swap-exact-amount-in",
		Long: `Query estimate-swap-exact-amount-in.{{.ExampleHeader}}
{{.CommandPrefix}} estimate-swap-exact-amount-in 1  1000stake --swap-route-pool-ids=2 --swap-route-pool-ids=3`,
		ParseQuery:          EstimateSwapExactAmountInParseArgs,
		Flags:               osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetMultihopSwapRoutes()}},
		QueryFnName:         "EstimateSwapExactAmountIn",
		CustomFlagOverrides: customRouterFlagOverride,
	}, &queryproto.EstimateSwapExactAmountInRequest{}
}

// GetCmdEstimateSwapExactAmountOut returns estimation of input coin to get exact amount of x token output.
func GetCmdEstimateSwapExactAmountOut() (*osmocli.QueryDescriptor, *queryproto.EstimateSwapExactAmountOutRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "estimate-swap-exact-amount-out <poolID> <tokenOut>",
		Short: "Query estimate-swap-exact-amount-out",
		Long: `Query estimate-swap-exact-amount-out.{{.ExampleHeader}}
{{.CommandPrefix}} estimate-swap-exact-amount-out 1 1000stake --swap-route-pool-ids=2 --swap-route-pool-ids=3`,
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

// GetCmdPool returns pool information.
func GetCmdPool() (*osmocli.QueryDescriptor, *queryproto.PoolRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pool [poolID]",
		Short: "Query pool",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pool 1`}, &queryproto.PoolRequest{}
}

func GetCmdSpotPrice() (*osmocli.QueryDescriptor, *queryproto.SpotPriceRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "spot-price <pool-ID> [quote-asset-denom] [base-asset-denom]",
		Short: "Query spot-price",
		Long: `Query spot-price
{{.CommandPrefix}} spot-price 1 uosmo ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
`}, &queryproto.SpotPriceRequest{}
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
		PoolId:  uint64(poolID), // TODO: is this poolId used?
		TokenIn: args[1],
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
		PoolId:   uint64(poolID), // TODO: is this poolId used?
		Routes:   routes,
		TokenOut: args[1],
	}, nil
}

// GetCmdEstimateSinglePoolSwapExactAmountIn returns estimation of output coin when amount of x token input.
func GetCmdEstimateSinglePoolSwapExactAmountIn() (*osmocli.QueryDescriptor, *queryproto.EstimateSinglePoolSwapExactAmountInRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "estimate-single-pool-swap-exact-amount-in <poolID> <tokenIn> <tokenOutDenom>",
		Short: "Query estimate-single-pool-swap-exact-amount-in",
		Long: `Query estimate-single-pool-swap-exact-amount-in.{{.ExampleHeader}}
{{.CommandPrefix}} estimate-single-pool-swap-exact-amount-in 1 1000stake uosmo`,
		QueryFnName: "EstimateSinglePoolSwapExactAmountIn",
	}, &queryproto.EstimateSinglePoolSwapExactAmountInRequest{}
}

// GetCmdEstimateSinglePoolSwapExactAmountOut returns estimation of input coin to get exact amount of x token output.
func GetCmdEstimateSinglePoolSwapExactAmountOut() (*osmocli.QueryDescriptor, *queryproto.EstimateSinglePoolSwapExactAmountOutRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "estimate-single-pool-swap-exact-amount-out <poolID> <tokenInDenom> <tokenOut>",
		Short: "Query estimate-single-pool-swap-exact-amount-out",
		Long: `Query estimate-single-pool-swap-exact-amount-out.{{.ExampleHeader}}
{{.CommandPrefix}} estimate-single-pool-swap-exact-amount-out 1 uosmo 1000stake`,
		QueryFnName: "EstimateSinglePoolSwapExactAmountOut",
	}, &queryproto.EstimateSinglePoolSwapExactAmountOutRequest{}
}

func GetCmdTotalPoolLiquidity() (*osmocli.QueryDescriptor, *queryproto.TotalPoolLiquidityRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "total-pool-liquidity [poolID]",
		Short: "Query total-pool-liquidity",
		Long: `{{.Short}} 
		{{.CommandPrefix}} total-pool-liquidity 1`,
	}, &queryproto.TotalPoolLiquidityRequest{}
}
