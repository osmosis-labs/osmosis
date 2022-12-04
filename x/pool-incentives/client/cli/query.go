package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/pool-incentives/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	cmd.AddCommand(
		GetCmdGaugeIds(),
		GetCmdDistrInfo(),
		osmocli.GetParams[*types.QueryParamsRequest, *types.QueryParamsResponse](
			types.ModuleName, types.NewQueryClient),
		GetCmdLockableDurations(),
		GetCmdIncentivizedPools(),
		GetCmdExternalIncentiveGauges(),
	)

	return cmd
}

// GetCmdGaugeIds takes the pool id and returns the matching gauge ids and durations.
func GetCmdGaugeIds() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryGaugeIdsRequest](
		"gauge-ids [pool-id]",
		"Query the matching gauge ids and durations by pool id",
		`Query the matching gauge ids and durations by pool id.

Example:
{{.CommandPrefix}} gauge-ids 1
`, types.ModuleName, types.NewQueryClient)
}

// GetCmdDistrInfo takes the pool id and returns the matching gauge ids and weights.
func GetCmdDistrInfo() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryDistrInfoRequest](
		"distr-info",
		"Query distribution info",
		`Query distribution info.

Example:
{{.CommandPrefix}} distr-info
`, types.ModuleName, types.NewQueryClient)
}

// GetCmdLockableDurations returns lockable durations.
func GetCmdLockableDurations() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryLockableDurationsRequest](
		"lockable-durations",
		"Query lockable durations",
		`Query distribution info.

Example:
{{.CommandPrefix}} lockable-durations
`, types.ModuleName, types.NewQueryClient)
}

func GetCmdIncentivizedPools() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryIncentivizedPoolsRequest](
		"incentivized-pools",
		"Query incentivized pools",
		`Query incentivized pools.

Example:
{{.CommandPrefix}} incentivized-pools
`, types.ModuleName, types.NewQueryClient)
}

func GetCmdExternalIncentiveGauges() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "external-incentivized-gauges",
		Short: "Query external incentivized gauges",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query incentivized gauges.

Example:
$ %s query pool-incentives external-incentivized-gauges
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

			res, err := queryClient.ExternalIncentiveGauges(cmd.Context(), &types.QueryExternalIncentiveGaugesRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
