package cli

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v7/x/launchpad"
	"github.com/osmosis-labs/osmosis/v7/x/launchpad/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	// Group launchpad queries under a subcommand
	cmd := &cobra.Command{
		Use:                        launchpad.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", launchpad.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQuerySales(),
		GetCmdQuerySale(),
		GetCmdUserPosition(),
	)

	return cmd
}

// GetCmdQuerySales implements a command to fetch launchpad sales.
func GetCmdQuerySales() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sales",
		Short: "Query launchpad sales list",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query sales.
Example:
$ %s query launchpad sales
`,
				version.AppName,
			),
		),
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
			sales := &types.QuerySales{
				Pagination: pageReq,
			}
			res, err := queryClient.Sales(cmd.Context(), sales)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQuerySale implements a command to get launchpad sale by id.
func GetCmdQuerySale() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sale <sale-id>",
		Short: "Query a launchpad sale by it's id",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query sale.
Example:
$ %s query launchpad sale 1
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			saleID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			sale := &types.QuerySale{
				SaleId: saleID,
			}
			res, err := queryClient.Sale(cmd.Context(), sale)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdUserPosition implements a command to get user position in a launchpad sale.
func GetCmdUserPosition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-position <sale-id> <address>",
		Short: "Query user position from a launchpad sale",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query user position.
Example:
$ %s query launchpad user-position 1 osmo1r85gjuck87f9hw7l2c30w3zh696xrq0lus0kq6
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			saleID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			address, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			userPosition := &types.QueryUserPosition{
				SaleId: saleID,
				User:   address.String(),
			}
			res, err := queryClient.UserPosition(cmd.Context(), userPosition)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
