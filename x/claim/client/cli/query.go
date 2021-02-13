package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/c-osmosis/osmosis/x/claim/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	claimQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	claimQueryCmd.AddCommand(
		GetCmdQueryClaimable(),
	)

	return claimQueryCmd
}

// GetCmdQueryClaimable implements the query claimables command.
func GetCmdQueryClaimable() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claimable [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the claimable amount per account.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the claimable amount for the account.
Example:
$ %s query claim claimable <address>
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
			// Query store
			res, err := queryClient.Claimable(context.Background(), &types.ClaimableRequest{Sender: args[0]})
			if err != nil {
				return err
			}
			return clientCtx.PrintObjectLegacy(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
