package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group superfluid queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdAssetTwap(),
		GetCmdAllIntermediaryAccounts(),
	)

	return cmd
}

// GetCmdAssetTwap returns twap of an asset by denom
func GetCmdAssetTwap() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset-twap [denom]",
		Short: "Query asset twap by denom",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query asset twap by denom.

Example:
$ %s query superfluid asset-twap gamm/pool/1
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

			res, err := queryClient.AssetTwap(cmd.Context(), &types.AssetTwapRequest{
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

// GetCmdAllIntermediaryAccounts returns all superfluid intermediary accounts
func GetCmdAllIntermediaryAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all-intermediary-accounts",
		Short: "Query ll superfluid intermediary accounts",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query ll superfluid intermediary accounts.

Example:
$ %s query superfluid all-intermediary-accounts
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

			res, err := queryClient.AllIntermediaryAccounts(cmd.Context(), &types.AllIntermediaryAccountsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "superfluid")

	return cmd
}
