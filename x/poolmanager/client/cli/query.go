package cli

import (
	"strconv"

	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
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
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdAllPools)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdPool)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdTotalVolumeForPool)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdTradingPairTakerFee)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdEstimateTradeBasedOnPriceImpact)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdListPoolsByDenom)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetAllTakerFeeShareAgreements)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetTakerFeeShareAgreementFromDenom)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetTakerFeeShareDenomsToAccruedValue)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetAllTakerFeeShareAccumulators)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetRegisteredAlloyedPoolFromDenom)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetRegisteredAlloyedPoolFromPoolId)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetAllRegisteredAlloyedPools)
	cmd.AddCommand(
		osmocli.GetParams[*queryproto.ParamsRequest](
			types.ModuleName, queryproto.NewQueryClient),
	)

	return cmd
}

// GetCmdEstimateSwapExactAmountIn returns estimation of output coin when amount of x token input.
func GetCmdEstimateSwapExactAmountIn() (*osmocli.QueryDescriptor, *queryproto.EstimateSwapExactAmountInRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "estimate-swap-exact-amount-in",
		Short: "Query estimate-swap-exact-amount-in",
		Long: `Query estimate-swap-exact-amount-in.{{.ExampleHeader}}
{{.CommandPrefix}} estimate-swap-exact-amount-in 1000stake --swap-route-pool-ids=2 --swap-route-pool-ids=3`,
		ParseQuery:          EstimateSwapExactAmountInParseArgs,
		Flags:               osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetMultihopSwapRoutes()}},
		QueryFnName:         "EstimateSwapExactAmountIn",
		CustomFlagOverrides: customRouterFlagOverride,
	}, &queryproto.EstimateSwapExactAmountInRequest{}
}

// GetCmdEstimateSwapExactAmountOut returns estimation of input coin to get exact amount of x token output.
func GetCmdEstimateSwapExactAmountOut() (*osmocli.QueryDescriptor, *queryproto.EstimateSwapExactAmountOutRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "estimate-swap-exact-amount-out",
		Short: "Query estimate-swap-exact-amount-out",
		Long: `Query estimate-swap-exact-amount-out.{{.ExampleHeader}}
{{.CommandPrefix}} estimate-swap-exact-amount-out 1000stake --swap-route-pool-ids=2 --swap-route-pool-ids=3`,
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

// GetCmdAllPools return all pools available across Osmosis modules.
func GetCmdAllPools() (*osmocli.QueryDescriptor, *queryproto.AllPoolsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "all-pools",
		Short: "Query all pools on the Osmosis chain",
		Long:  "{{.Short}}",
	}, &queryproto.AllPoolsRequest{}
}

// GetCmdPool returns pool information.
func GetCmdPool() (*osmocli.QueryDescriptor, *queryproto.PoolRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pool",
		Short: "Query pool",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pool 1`,
	}, &queryproto.PoolRequest{}
}

func GetCmdSpotPrice() (*osmocli.QueryDescriptor, *queryproto.SpotPriceRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "spot-price",
		Short: "Query spot-price",
		Long: `Query spot-price
{{.CommandPrefix}} spot-price 1 uosmo ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
`,
	}, &queryproto.SpotPriceRequest{}
}
func GetCmdListPoolsByDenom() (*osmocli.QueryDescriptor, *queryproto.ListPoolsByDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "list-pools-by-denom",
		Short: "Query list-pools-by-denom",
		Long: `Query list-pools-by-denom
{{.CommandPrefix}} list-pools-by-denom uosmo 
`,
	}, &queryproto.ListPoolsByDenomRequest{}
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
		Use:   "estimate-single-pool-swap-exact-amount-in",
		Short: "Query estimate-single-pool-swap-exact-amount-in",
		Long: `Query estimate-single-pool-swap-exact-amount-in.{{.ExampleHeader}}
{{.CommandPrefix}} estimate-single-pool-swap-exact-amount-in 1 1000stake uosmo`,
		QueryFnName: "EstimateSinglePoolSwapExactAmountIn",
	}, &queryproto.EstimateSinglePoolSwapExactAmountInRequest{}
}

// GetCmdEstimateSinglePoolSwapExactAmountOut returns estimation of input coin to get exact amount of x token output.
func GetCmdEstimateSinglePoolSwapExactAmountOut() (*osmocli.QueryDescriptor, *queryproto.EstimateSinglePoolSwapExactAmountOutRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "estimate-single-pool-swap-exact-amount-out",
		Short: "Query estimate-single-pool-swap-exact-amount-out",
		Long: `Query estimate-single-pool-swap-exact-amount-out.{{.ExampleHeader}}
{{.CommandPrefix}} estimate-single-pool-swap-exact-amount-out 1 uosmo 1000stake`,
		QueryFnName: "EstimateSinglePoolSwapExactAmountOut",
	}, &queryproto.EstimateSinglePoolSwapExactAmountOutRequest{}
}

func GetCmdTotalPoolLiquidity() (*osmocli.QueryDescriptor, *queryproto.TotalPoolLiquidityRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "total-pool-liquidity",
		Short: "Query total-pool-liquidity",
		Long: `{{.Short}}
		{{.CommandPrefix}} total-pool-liquidity 1`,
	}, &queryproto.TotalPoolLiquidityRequest{}
}

func GetCmdTotalVolumeForPool() (*osmocli.QueryDescriptor, *queryproto.TotalVolumeForPoolRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "total-volume-for-pool",
		Short: "Query total-volume-for-pool",
		Long: `{{.Short}}
		{{.CommandPrefix}} total-volume-for-pool 1`,
	}, &queryproto.TotalVolumeForPoolRequest{}
}

func GetCmdTradingPairTakerFee() (*osmocli.QueryDescriptor, *queryproto.TradingPairTakerFeeRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "trading-pair-taker-fee",
		Short: "Query trading pair taker fee",
		Long: `{{.Short}}
		{{.CommandPrefix}} trading-pair-taker-fee uosmo uatom`,
	}, &queryproto.TradingPairTakerFeeRequest{}
}

func GetCmdEstimateTradeBasedOnPriceImpact() (
	*osmocli.QueryDescriptor, *queryproto.EstimateTradeBasedOnPriceImpactRequest,
) {
	return &osmocli.QueryDescriptor{
		Use:   "estimate-trade-based-on-price-impact",
		Short: "Query estimate-trade-based-on-price-impact",
		Long: `{{.Short}}
		{{.CommandPrefix}} estimate-trade-based-on-price-impact 100uosmo stosmo  833 0.001 1.00`,
		QueryFnName: "EstimateTradeBasedOnPriceImpact",
	}, &queryproto.EstimateTradeBasedOnPriceImpactRequest{}
}

func GetAllTakerFeeShareAgreements() (*osmocli.QueryDescriptor, *queryproto.AllTakerFeeShareAgreementsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "all-taker-fee-share-agreements",
		Short: "Query all taker fee share agreements",
		Long:  "{{.Short}}",
	}, &queryproto.AllTakerFeeShareAgreementsRequest{}
}

func GetTakerFeeShareAgreementFromDenom() (*osmocli.QueryDescriptor, *queryproto.TakerFeeShareAgreementFromDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "taker-fee-share-agreement-from-denom",
		Short: "Query taker fee share agreement from denom",
		Long: `{{.Short}}
		{{.CommandPrefix}} taker-fee-share-agreement-from-denom uosmo`,
	}, &queryproto.TakerFeeShareAgreementFromDenomRequest{}
}

func GetTakerFeeShareDenomsToAccruedValue() (*osmocli.QueryDescriptor, *queryproto.TakerFeeShareDenomsToAccruedValueRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "taker-fee-share-denoms-to-accrued-value",
		Short: "Query taker fee share denoms to accrued value",
		Long: `{{.Short}}
		{{.CommandPrefix}} taker-fee-share-denoms-to-accrued-value uosmo`,
	}, &queryproto.TakerFeeShareDenomsToAccruedValueRequest{}
}

func GetAllTakerFeeShareAccumulators() (*osmocli.QueryDescriptor, *queryproto.AllTakerFeeShareAccumulatorsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "all-taker-fee-share-accumulators",
		Short: "Query all taker fee share accumulators",
		Long:  "{{.Short}}",
	}, &queryproto.AllTakerFeeShareAccumulatorsRequest{}
}

func GetRegisteredAlloyedPoolFromDenom() (*osmocli.QueryDescriptor, *queryproto.RegisteredAlloyedPoolFromDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "registered-alloyed-pool-from-denom",
		Short: "Query registered alloyed pool from the alloyed pool denom",
		Long: `{{.Short}}
		{{.CommandPrefix}} registered-alloyed-pool-from-denom factory/osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3/alloyed/allBTC`,
	}, &queryproto.RegisteredAlloyedPoolFromDenomRequest{}
}

func GetRegisteredAlloyedPoolFromPoolId() (*osmocli.QueryDescriptor, *queryproto.RegisteredAlloyedPoolFromPoolIdRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "registered-alloyed-pool-from-pool-id",
		Short: "Query registered alloyed pool from pool id",
		Long: `{{.Short}}
		{{.CommandPrefix}} registered-alloyed-pool-from-pool-id 1868`,
	}, &queryproto.RegisteredAlloyedPoolFromPoolIdRequest{}
}

func GetAllRegisteredAlloyedPools() (*osmocli.QueryDescriptor, *queryproto.AllRegisteredAlloyedPoolsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "all-registered-alloyed-pools",
		Short: "Query all registered alloyed pools",
		Long:  "{{.Short}}",
	}, &queryproto.AllRegisteredAlloyedPoolsRequest{}
}
