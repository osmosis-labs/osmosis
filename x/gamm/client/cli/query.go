package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v2"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdPool)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdSpotPrice)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdPool)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdPools)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdEstimateSwapExactAmountIn)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdEstimateSwapExactAmountOut)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetConcentratedPoolIdLinkFromCFMMRequest)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCFMMConcentratedPoolLinksRequest)
	cmd.AddCommand(
		GetCmdNumPools(),
		GetCmdPoolParams(),
		GetCmdTotalShares(),
		GetCmdQueryTotalLiquidity(),
		GetCmdTotalPoolLiquidity(),
		GetCmdQueryPoolsWithFilter(),
		GetCmdPoolType(),
	)

	return cmd
}

var customRouterFlagOverride = map[string]string{
	"router": FlagSwapRouteDenoms,
}

// Deprecated: use x/poolmanager's Pool query.
// nolint: staticcheck
func GetCmdPool() (*osmocli.QueryDescriptor, *types.QueryPoolRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pool [poolID]",
		Short: "Query pool",
		// Deprecated: use x/poolmanager's Pool query.
		// nolint: staticcheck
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pool 1`,
	}, &types.QueryPoolRequest{}
}

// TODO: Push this to the SDK.
func writeOutputBoilerplate(ctx client.Context, out []byte) error {
	writer := ctx.Output
	if writer == nil {
		writer = os.Stdout
	}

	_, err := writer.Write(out)
	if err != nil {
		return err
	}

	if ctx.OutputFormat != "text" {
		// append new-line for formats besides YAML
		_, err = writer.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func GetCmdPools() (*osmocli.QueryDescriptor, *types.QueryPoolsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pools",
		Short: "Query pools",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pools`,
	}, &types.QueryPoolsRequest{}
}

// nolint: staticcheck
func GetCmdNumPools() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryNumPoolsRequest](
		"num-pools",
		"Query number of pools",
		"{{.Short}}",
		types.ModuleName, types.NewQueryClient,
	)
}

// GetCmdPoolParams return pool params.
func GetCmdPoolParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool-params <poolID>",
		Short: "Query pool-params",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query pool-params.
Example:
$ %s query gamm pool-params 1
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.PoolParams(cmd.Context(), &types.QueryPoolParamsRequest{
				PoolId: uint64(poolID),
			})
			if err != nil {
				return err
			}

			if clientCtx.OutputFormat == "text" {
				poolParams := &balancer.PoolParams{}
				if err := poolParams.Unmarshal(res.GetParams().Value); err != nil {
					return err
				}

				out, err := yaml.Marshal(poolParams)
				if err != nil {
					return err
				}
				return writeOutputBoilerplate(clientCtx, out)
			} else {
				out, err := clientCtx.Codec.MarshalJSON(res)
				if err != nil {
					return err
				}

				return writeOutputBoilerplate(clientCtx, out)
			}
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetCmdTotalShares() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryTotalSharesRequest](
		"total-share [poolID]",
		"Query total-share",
		`Query total-share.
Example:
{{.CommandPrefix}} total-share 1
`,
		types.ModuleName, types.NewQueryClient,
	)
}

func GetCmdQueryTotalLiquidity() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryTotalLiquidityRequest](
		"total-liquidity",
		"Query total-liquidity",
		`Query total-liquidity.
Example:
{{.CommandPrefix}} total-liquidity
`,
		types.ModuleName, types.NewQueryClient,
	)
}

// Deprecated: use alternate in x/poolmanager.
func GetCmdSpotPrice() (*osmocli.QueryDescriptor, *types.QuerySpotPriceRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "spot-price <pool-ID> [quote-asset-denom] [base-asset-denom]",
		Short: "Query spot-price (LEGACY, arguments are reversed!!)",
		Long: `Query spot price (Legacy).{{.ExampleHeader}}
{{.CommandPrefix}} spot-price 1 uosmo ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
`,
	}, &types.QuerySpotPriceRequest{}
}

// Deprecated: use alternate in x/poolmanager.
func GetCmdEstimateSwapExactAmountIn() (*osmocli.QueryDescriptor, *types.QuerySwapExactAmountInRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "estimate-swap-exact-amount-in <poolID> <sender> <tokenIn>",
		Short: "Query estimate-swap-exact-amount-in",
		Long: `Query estimate-swap-exact-amount-in.{{.ExampleHeader}}
{{.CommandPrefix}} estimate-swap-exact-amount-in 1 osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7 1000stake --swap-route-pool-ids=2 --swap-route-pool-ids=3`,
		ParseQuery:          EstimateSwapExactAmountInParseArgs,
		Flags:               osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetMultihopSwapRoutes()}},
		QueryFnName:         "EstimateSwapExactAmountIn",
		CustomFlagOverrides: customRouterFlagOverride,
	}, &types.QuerySwapExactAmountInRequest{}
}

// Deprecated: use alternate in x/poolmanager.
func GetCmdEstimateSwapExactAmountOut() (*osmocli.QueryDescriptor, *types.QuerySwapExactAmountOutRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "estimate-swap-exact-amount-out <poolID> <sender> <tokenOut>",
		Short: "Query estimate-swap-exact-amount-out",
		Long: `Query estimate-swap-exact-amount-out.{{.ExampleHeader}}
{{.CommandPrefix}} estimate-swap-exact-amount-out 1 osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7 1000stake --swap-route-pool-ids=2 --swap-route-pool-ids=3`,
		ParseQuery:          EstimateSwapExactAmountOutParseArgs,
		Flags:               osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetMultihopSwapRoutes()}},
		QueryFnName:         "EstimateSwapExactAmountOut",
		CustomFlagOverrides: customRouterFlagOverride,
	}, &types.QuerySwapExactAmountOutRequest{}
}

// nolint: staticcheck
func EstimateSwapExactAmountInParseArgs(args []string, fs *flag.FlagSet) (proto.Message, error) {
	poolID, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, err
	}

	routes, err := swapAmountInRoutes(fs)
	if err != nil {
		return nil, err
	}

	return &types.QuerySwapExactAmountInRequest{
		Sender:  args[1],        // TODO: where sender is used?
		PoolId:  uint64(poolID), // TODO: is this poolId used?
		TokenIn: args[2],
		Routes:  routes,
	}, nil
}

// nolint: staticcheck
func EstimateSwapExactAmountOutParseArgs(args []string, fs *flag.FlagSet) (proto.Message, error) {
	poolID, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, err
	}

	routes, err := swapAmountOutRoutes(fs)
	if err != nil {
		return nil, err
	}

	return &types.QuerySwapExactAmountOutRequest{
		Sender:   args[1],        // TODO: where sender is used?
		PoolId:   uint64(poolID), // TODO: is this poolId used?
		Routes:   routes,
		TokenOut: args[2],
	}, nil
}

// GetCmdQueryPoolsWithFilter returns pool with filter
func GetCmdQueryPoolsWithFilter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools-with-filter <min_liquidity> <pool_type>",
		Short: "Query pools with filter",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query pools with filter. The possible filter options are:

1. By Pool type: Either "Balancer" or "Stableswap"
2. By min pool liquidity by providing min coins

Note that if both filters are to be applied, "min_liquidity" always needs to be provided as the first argument.

Example:
$ %s query gamm pools-with-filter <min_liquidity> <pool_type> 
`,
				version.AppName,
			),
		),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			var pool_type string
			min_liquidity := args[0]
			if len(args) > 1 {
				pool_type = args[1]
			}
			res, err := queryClient.PoolsWithFilter(cmd.Context(), &types.QueryPoolsWithFilterRequest{
				MinLiquidity: min_liquidity,
				PoolType:     pool_type,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdPoolType returns pool type given pool id.
func GetCmdPoolType() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryPoolTypeRequest](
		"pool-type <pool_id>",
		"Query pool type",
		`Query pool type
Example:
{{.CommandPrefix}} pool-type <pool_id>
`,
		types.ModuleName, types.NewQueryClient,
	)
}

// GetConcentratedPoolIdLinkFromCFMMRequest returns concentrated pool id that is linked to the given cfmm pool id.
func GetConcentratedPoolIdLinkFromCFMMRequest() (*osmocli.QueryDescriptor, *types.QueryConcentratedPoolIdLinkFromCFMMRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "cl-pool-link-from-cfmm [poolID]",
		Short: "Query concentrated pool id link from cfmm pool id",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} cl-pool-link-from-cfmm 1`,
	}, &types.QueryConcentratedPoolIdLinkFromCFMMRequest{}
}

// GetCmdTotalPoolLiquidity returns total liquidity in pool.
// Deprecated: please use the alternative in x/poolmanager
// nolint: staticcheck
func GetCmdTotalPoolLiquidity() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryTotalPoolLiquidityRequest](
		"total-pool-liquidity [poolID]",
		"Query total-pool-liquidity",
		`Query total-pool-liquidity.
Example:
{{.CommandPrefix}} total-pool-liquidity 1
`,
		types.ModuleName, types.NewQueryClient,
	)
}

// GetConcentratedPoolIdLinkFromCFMMRequest returns all concentrated pool id to cfmm pool id links.
func GetCFMMConcentratedPoolLinksRequest() (*osmocli.QueryDescriptor, *types.QueryCFMMConcentratedPoolLinksRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "cfmm-cl-pool-links",
		Short: "Query all concentrated pool and cfmm pool id links",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} cfmm-cl-pool-links`,
	}, &types.QueryCFMMConcentratedPoolLinksRequest{}
}
