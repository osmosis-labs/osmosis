package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v9/x/txfees/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	// Group queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdFeeTokens(),
		GetCmdDenomPoolID(),
		GetCmdBaseDenom(),
	)

	return cmd
}

// GetCmdFeeTokens takes the pool id and returns the matching gauge ids and durations.
func GetCmdFeeTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee-tokens",
		Short: "Query the list of non-basedenom fee tokens and their associated pool ids",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the list of non-basedenom fee tokens and their associated pool ids

Example:
$ %s query txfees fee-tokens
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

			res, err := queryClient.FeeTokens(cmd.Context(), &types.QueryFeeTokensRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdDenomPoolID takes the pool id and returns the matching gauge ids and durations.
func GetCmdDenomPoolID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "denom-pool-id",
		Short: "Query the pool id associated with a specific whitelisted fee token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the pool id associated with a specific fee token

Example:
$ %s query txfees denom-pool-id [denom]
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

			res, err := queryClient.DenomPoolId(cmd.Context(), &types.QueryDenomPoolIdRequest{
				Denom: args[0],
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

// GetCmdBaseDenom takes the pool id and returns the matching gauge ids and weights.
func GetCmdBaseDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "base-denom",
		Short: "Query the base fee denom",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the base fee denom.

Example:
$ %s query txfees base-denom
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

			res, err := queryClient.BaseDenom(cmd.Context(), &types.QueryBaseDenomRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
