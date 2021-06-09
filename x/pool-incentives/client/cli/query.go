package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/osmosis-labs/osmosis/x/pool-incentives/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdGaugeIds(),
	)

	return cmd
}

// GetCmdGaugeIds takes the pool id and returns the matching gauge ids and durations
func GetCmdGaugeIds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gauge-ids [pool-id]",
		Short: "Query the matching gauge ids and durations by pool id",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the matching gauge ids and durations by pool id.

Example:
$ %s query pool-incentives gauge-ids 1
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

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.GaugeIds(cmd.Context(), &types.QueryGaugeIdsRequest{
				PoolId: poolId,
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

//   rpc DistrInfo(QueryDistrInfoRequest) returns (QueryDistrInfoResponse) {
//     option (google.api.http).get =
//         "/osmosis/pool-incentives/v1beta1/distr_info";
//   }

//   rpc Params(QueryParamsRequest) returns (QueryParamsResponse) {
//     option (google.api.http).get = "/osmosis/pool-incentives/v1beta1/params";
//   }

//   rpc LockableDurations(QueryLockableDurationsRequest)
//       returns (QueryLockableDurationsResponse) {
//     option (google.api.http).get =
//         "/osmosis/pool-incentives/v1beta1/lockable_durations";
//   }

//   rpc IncentivizedPools(QueryIncentivizedPoolsRequest)
//       returns (QueryIncentivizedPoolsResponse) {
//     option (google.api.http).get =
//         "/osmosis/pool-incentives/v1beta1/incentivized_pools";
//   }
