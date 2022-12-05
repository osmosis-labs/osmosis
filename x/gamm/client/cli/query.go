package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	cmd.AddCommand(
		GetCmdPool(),
		GetCmdPools(),
		GetCmdNumPools(),
		GetCmdPoolParams(),
		GetCmdTotalShares(),
		GetCmdSpotPrice(),
		GetCmdQueryTotalLiquidity(),
		GetCmdEstimateSwapExactAmountIn(),
		GetCmdEstimateSwapExactAmountOut(),
		GetCmdTotalPoolLiquidity(),
		GetCmdQueryPoolsWithFilter(),
		GetCmdPoolType(),
	)

	return cmd
}

func GetCmdPool() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryPoolRequest](
		"pool [poolID]",
		"Query pool",
		`Query pool.
Example:
{{.CommandPrefix}} pool 1
`,
		types.ModuleName, types.NewQueryClient,
	)
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

func GetCmdPools() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryPoolsRequest](
		"pools",
		"Query pools",
		`{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pools`,
		types.ModuleName, types.NewQueryClient,
	)
}

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

func GetCmdSpotPrice() *cobra.Command {
	//nolint:staticcheck
	return osmocli.SimpleQueryCmd[*types.QuerySpotPriceRequest](
		"spot-price <pool-ID> [quote-asset-denom] [base-asset-denom]",
		"Query spot-price (LEGACY, arguments are reversed!!)",
		`Query spot price (Legacy).
Example:
{{.CommandPrefix}} spot-price 1 uosmo ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
`,
		types.ModuleName, types.NewQueryClient,
	)
}

// GetCmdEstimateSwapExactAmountIn returns estimation of output coin when amount of x token input.
func GetCmdEstimateSwapExactAmountIn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate-swap-exact-amount-in <poolID> <sender> <tokenIn>",
		Short: "Query estimate-swap-exact-amount-in",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query estimate-swap-exact-amount-in.
Example:
$ %s query gamm estimate-swap-exact-amount-in 1 osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7 stake --swap-route-pool-ids=2 --swap-route-pool-ids=3
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(3),
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

			routes, err := swapAmountInRoutes(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.EstimateSwapExactAmountIn(cmd.Context(), &types.QuerySwapExactAmountInRequest{
				Sender:  args[1],        // TODO: where sender is used?
				PoolId:  uint64(poolID), // TODO: is this poolId used?
				TokenIn: args[2],
				Routes:  routes,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetQuerySwapRoutes())
	flags.AddQueryFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(FlagSwapRoutePoolIds)
	_ = cmd.MarkFlagRequired(FlagSwapRouteDenoms)

	return cmd
}

// GetCmdEstimateSwapExactAmountOut returns estimation of input coin to get exact amount of x token output.
func GetCmdEstimateSwapExactAmountOut() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate-swap-exact-amount-out <poolID> <sender> <tokenOut>",
		Short: "Query estimate-swap-exact-amount-out",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query estimate-swap-exact-amount-out.
Example:
$ %s query gamm estimate-swap-exact-amount-out 1 osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7 stake --swap-route-pool-ids=2 --swap-route-pool-ids=3
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(3),
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

			routes, err := swapAmountOutRoutes(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.EstimateSwapExactAmountOut(cmd.Context(), &types.QuerySwapExactAmountOutRequest{
				Sender:   args[1],        // TODO: where sender is used?
				PoolId:   uint64(poolID), // TODO: is this poolId used?
				Routes:   routes,
				TokenOut: args[2],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetQuerySwapRoutes())
	flags.AddQueryFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(FlagSwapRoutePoolIds)
	_ = cmd.MarkFlagRequired(FlagSwapRouteDenoms)

	return cmd
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

			var min_liquidity sdk.Coins
			var pool_type string
			if len(args) == 1 {
				coins, err := sdk.ParseCoinsNormalized(args[0])
				if err != nil {
					pool_type = args[0]
				}
				min_liquidity = coins
			} else {
				coins, err := sdk.ParseCoinsNormalized(args[0])
				if err != nil {
					return status.Errorf(codes.InvalidArgument, "invalid token: %s", err.Error())
				}

				min_liquidity = coins
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
