package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v9/x/pool-incentives/types"

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
		GetCmdGaugeIds(),
		GetCmdDistrInfo(),
		GetCmdParams(),
		GetCmdLockableDurations(),
		GetCmdIncentivizedPools(),
		GetCmdExternalIncentiveGauges(),
	)

	return cmd
}

// GetCmdGaugeIds takes the pool id and returns the matching gauge ids and durations.
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

// GetCmdDistrInfo takes the pool id and returns the matching gauge ids and weights.
func GetCmdDistrInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "distr-info",
		Short: "Query distribution info",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query distribution info.

Example:
$ %s query pool-incentives distr-info
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

			res, err := queryClient.DistrInfo(cmd.Context(), &types.QueryDistrInfoRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdParams returns module params.
func GetCmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query module params",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query module params.

Example:
$ %s query pool-incentives params
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

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdLockableDurations returns lockable durations.
func GetCmdLockableDurations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lockable-durations",
		Short: "Query lockable durations",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query lockable durations.

Example:
$ %s query pool-incentives lockable-durations
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

			res, err := queryClient.LockableDurations(cmd.Context(), &types.QueryLockableDurationsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdIncentivizedPools returns incentivized pools.
func GetCmdIncentivizedPools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incentivized-pools",
		Short: "Query incentivized pools",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query incentivized pools.

Example:
$ %s query pool-incentives incentivized-pools
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

			res, err := queryClient.IncentivizedPools(cmd.Context(), &types.QueryIncentivizedPoolsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdIncentivizedPools returns incentivized pools.
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
