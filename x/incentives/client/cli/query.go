package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group incentives queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdPots(),
	)

	return cmd
}

// GetCmdPots returns full available pots
func GetCmdPots() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pots",
		Short: "Query available pots",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query available pots.

Example:
$ %s query incentives pots
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

			res, err := queryClient.Pots(cmd.Context(), &types.PotsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "incentives")

	return cmd
}

// TODO: implement all the queries in proto
// // returns coins that is going to be distributed
// rpc ModuleToDistributeCoins(ModuleToDistributeCoinsRequest) returns (ModuleToDistributeCoinsResponse);
// // returns coins that are distributed by module so far
// rpc ModuleDistributedCoins(ModuleDistributedCoinsRequest) returns (ModuleDistributedCoinsResponse);
// // returns Pot by id
// rpc PotByID(PotByIDRequest) returns (PotByIDResponse);
// // returns pots both upcoming and active
// rpc Pots(PotsRequest) returns (PotsResponse);
// // returns active pots
// rpc ActivePots(ActivePotsRequest) returns (ActivePotsResponse);
// // returns scheduled pots
// rpc UpcomingPots(UpcomingPotsRequest) returns (UpcomingPotsResponse);
// // returns rewards estimation at a future specific time
// rpc RewardsEst(RewardsEstRequest) returns (RewardsEstResponse);
