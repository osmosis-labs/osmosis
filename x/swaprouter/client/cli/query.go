package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/client/queryproto"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdNumPools)
	cmd.AddCommand(
		GetCmdEstimateSwapExactAmountIn(),
		GetCmdEstimateSwapExactAmountOut(),
	)

	return cmd
}

// GetCmdEstimateSwapExactAmountIn returns estimation of output coin when amount of x token input.
func GetCmdEstimateSwapExactAmountIn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate-swap-exact-amount-in <poolID> <sender> <tokenIn>",
		Short: "Query estimate-swap-exact-amount-in",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query estimate-swap-exact-amount-in.
Example:
$ %s query swaprouter estimate-swap-exact-amount-in 1 osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7 stake --swap-route-pool-ids=2 --swap-route-pool-ids=3
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
			queryClient := queryproto.NewQueryClient(clientCtx)

			poolID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			routes, err := swapAmountInRoutes(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.EstimateSwapExactAmountIn(cmd.Context(), &queryproto.EstimateSwapExactAmountInRequest{
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
$ %s query swaprouter estimate-swap-exact-amount-out 1 osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7 stake --swap-route-pool-ids=2 --swap-route-pool-ids=3
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
			queryClient := queryproto.NewQueryClient(clientCtx)

			poolID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			routes, err := swapAmountOutRoutes(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.EstimateSwapExactAmountOut(cmd.Context(), &queryproto.EstimateSwapExactAmountOutRequest{
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

// GetCmdNumPools return number of pools available.
func GetCmdNumPools() (*osmocli.QueryDescriptor, *queryproto.NumPoolsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "num-pools",
		Short: "Query number of pools",
		Long:  "{{.Short}}",
	}, &queryproto.NumPoolsRequest{}
}
