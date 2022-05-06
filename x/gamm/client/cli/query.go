package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group gamm queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdPool(),
		GetCmdPools(),
		GetCmdNumPools(),
		GetCmdPoolParams(),
		GetCmdTotalShares(),
		GetCmdPoolAssets(),
		GetCmdSpotPrice(),
		GetCmdQueryTotalLiquidity(),
		GetCmdEstimateSwapExactAmountIn(),
		GetCmdEstimateSwapExactAmountOut(),
	)

	return cmd
}

// GetCmdPool returns pool
func GetCmdPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool <poolID>",
		Short: "Query pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query pool.
Example:
$ %s query gamm pool 1
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

			res, err := queryClient.Pool(cmd.Context(), &types.QueryPoolRequest{
				PoolId: uint64(poolID),
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

// TODO: Push this to the SDK
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

// GetCmdPools return pools
func GetCmdPools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Short: "Query pools",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query pools.
Example:
$ %s query gamm pools
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Pools(cmd.Context(), &types.QueryPoolsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "pools")

	return cmd
}

// GetCmdNumPools return number of pools available
func GetCmdNumPools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "num-pools",
		Short: "Query number of pools",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query number of pools.
Example:
$ %s query gamm num-pools
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.NumPools(cmd.Context(), &types.QueryNumPoolsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdPoolParams return pool params
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

// GetCmdTotalShares return total share
func GetCmdTotalShares() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-share <poolID>",
		Short: "Query total-share",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query total-share.
Example:
$ %s query gamm total-share 1
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

			res, err := queryClient.TotalShares(cmd.Context(), &types.QueryTotalSharesRequest{
				PoolId: uint64(poolID),
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

// GetCmdQueryTotalLiquidity return total liquidity
func GetCmdQueryTotalLiquidity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-liquidity",
		Short: "Query total-liquidity",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query total-liquidity.
Example:
$ %s query gamm total-liquidity
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.TotalLiquidity(cmd.Context(), &types.QueryTotalLiquidityRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdPoolAssets return pool-assets for a pool
func GetCmdPoolAssets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool-assets <poolID>",
		Short: "Query pool-assets",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query pool assets.
Example:
$ %s query gamm pool-assets 1
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

			res, err := queryClient.PoolAssets(cmd.Context(), &types.QueryPoolAssetsRequest{
				PoolId: uint64(poolID),
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

// GetCmdSpotPrice returns spot price
func GetCmdSpotPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spot-price <poolID> <tokenInDenom> <tokenOutDenom>",
		Short: "Query spot-price",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query spot-price.
Example:
$ %s query gamm spot-price 1 stake stake2
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

			res, err := queryClient.SpotPrice(cmd.Context(), &types.QuerySpotPriceRequest{
				PoolId:        uint64(poolID),
				TokenInDenom:  args[1],
				TokenOutDenom: args[2],
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

// GetCmdEstimateSwapExactAmountIn returns estimation of output coin when amount of x token input
func GetCmdEstimateSwapExactAmountIn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate-swap-exact-amount-in <poolID> <sender> <tokenIn>",
		Short: "Query estimate-swap-exact-amount-in",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query estimate-swap-exact-amount-in.
Example:
$ %s query gamm estimate-swap-exact-amount-in 1 osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7 stake --swap-route-pool-ids=2 --swap-route-amounts=100stake2 --swap-route-pool-ids=3 --swap-route-amounts=100stake
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

// GetCmdEstimateSwapExactAmountOut returns estimation of input coin to get exact amount of x token output
func GetCmdEstimateSwapExactAmountOut() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate-swap-exact-amount-out <poolID> <sender> <tokenOut>",
		Short: "Query estimate-swap-exact-amount-out",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query estimate-swap-exact-amount-out.
Example:
$ %s query gamm estimate-swap-exact-amount-out 1 osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7 stake --swap-route-pool-ids=2 --swap-route-amounts=100stake2 --swap-route-pool-ids=3 --swap-route-amounts=100stake
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
