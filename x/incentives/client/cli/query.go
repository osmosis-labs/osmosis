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
		GetCmdGauges(),
	)

	return cmd
}

// GetCmdGauges returns full available gauges
func GetCmdGauges() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gauges",
		Short: "Query available gauges",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query available gauges.

Example:
$ %s query incentives gauges
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

			res, err := queryClient.Gauges(cmd.Context(), &types.GaugesRequest{
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
// // returns Gauge by id
// rpc GaugeByID(GaugeByIDRequest) returns (GaugeByIDResponse);
// // returns gauges both upcoming and active
// rpc Gauges(GaugesRequest) returns (GaugesResponse);
// // returns active gauges
// rpc ActiveGauges(ActiveGaugesRequest) returns (ActiveGaugesResponse);
// // returns scheduled gauges
// rpc UpcomingGauges(UpcomingGaugesRequest) returns (UpcomingGaugesResponse);
// // returns rewards estimation at a future specific time
// rpc RewardsEst(RewardsEstRequest) returns (RewardsEstResponse);
