package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd builds query command group for the module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the callback module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		getQueryParamsCmd(),
		getQueryEstimateCallbackFeesCmd(),
		getQueryCallbacksCmd(),
	)
	return cmd
}

func getQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query module parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func getQueryEstimateCallbackFeesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "estimate-callback-fees [block-height]",
		Aliases: []string{"estimate-fees"},
		Args:    cobra.ExactArgs(1),
		Short:   "Query callback registration fees for a given block height",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			blockHeight, err := ParseInt64Arg("block-height", args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.EstimateCallbackFees(cmd.Context(), &types.QueryEstimateCallbackFeesRequest{
				BlockHeight: blockHeight,
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

func getQueryCallbacksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "callbacks [block-height]",
		Args:  cobra.ExactArgs(1),
		Short: "Query callbacks for a given block height",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			blockHeight, err := ParseInt64Arg("block-height", args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Callbacks(cmd.Context(), &types.QueryCallbacksRequest{
				BlockHeight: blockHeight,
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
