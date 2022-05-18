package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v9/x/incentives/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
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
		GetCmdToDistributeCoins(),
		GetCmdDistributedCoins(),
		GetCmdGaugeByID(),
		GetCmdActiveGauges(),
		GetCmdActiveGaugesPerDenom(),
		GetCmdUpcomingGauges(),
		GetCmdUpcomingGaugesPerDenom(),
		GetCmdRewardsEst(),
	)

	return cmd
}

// GetCmdGauges returns full available gauges.
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

// GetCmdToDistributeCoins returns coins that is going to be distributed.
func GetCmdToDistributeCoins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "to-distribute-coins",
		Short: "Query coins that is going to be distributed",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query coins that is going to be distributed.

Example:
$ %s query incentives to-distribute-coins
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

			res, err := queryClient.ModuleToDistributeCoins(cmd.Context(), &types.ModuleToDistributeCoinsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdDistributedCoins returns coins that are distributed so far.
func GetCmdDistributedCoins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "distributed-coins",
		Short: "Query coins distributed so far",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query coins distributed so far.

Example:
$ %s query incentives distributed-coins
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

			res, err := queryClient.ModuleDistributedCoins(cmd.Context(), &types.ModuleDistributedCoinsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdGaugeByID returns Gauge by id.
func GetCmdGaugeByID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gauge-by-id [id]",
		Short: "Query gauge by id.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query gauge by id.

Example:
$ %s query incentives gauge-by-id 1
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

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.GaugeByID(cmd.Context(), &types.GaugeByIDRequest{Id: id})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdActiveGauges returns active gauges.
func GetCmdActiveGauges() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "active-gauges",
		Short: "Query active gauges",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query active gauges.

Example:
$ %s query incentives active-gauges
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

			res, err := queryClient.ActiveGauges(cmd.Context(), &types.ActiveGaugesRequest{Pagination: pageReq})
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

// GetCmdActiveGaugesPerDenom returns active gauges for specified denom.
func GetCmdActiveGaugesPerDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "active-gauges-per-denom [denom]",
		Short: "Query active gauges per denom",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query active gauges.

Example:
$ %s query incentives active-gauges-per-denom [denom]
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
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.ActiveGaugesPerDenom(cmd.Context(), &types.ActiveGaugesPerDenomRequest{Denom: args[0], Pagination: pageReq})
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

// GetCmdUpcomingGauges returns scheduled gauges.
func GetCmdUpcomingGauges() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upcoming-gauges",
		Short: "Query scheduled gauges",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query scheduled gauges.

Example:
$ %s query incentives upcoming-gauges
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

			res, err := queryClient.UpcomingGauges(cmd.Context(), &types.UpcomingGaugesRequest{Pagination: pageReq})
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

// GetCmdActiveGaugesPerDenom returns active gauges for specified denom.
func GetCmdUpcomingGaugesPerDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upcoming-gauges-per-denom [denom]",
		Short: "Query scheduled gauges per denom",
		Args:  cobra.ExactArgs(1),
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

			res, err := queryClient.UpcomingGaugesPerDenom(cmd.Context(), &types.UpcomingGaugesPerDenomRequest{Denom: args[0], Pagination: pageReq})
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

// GetCmdRewardsEst returns rewards estimation.
func GetCmdRewardsEst() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards-estimation",
		Short: "Query rewards estimation",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query rewards estimation.

Example:
$ %s query incentives rewards-estimation 
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

			owner, err := cmd.Flags().GetString(FlagOwner)
			if err != nil {
				return err
			}

			lockIdsCombined, err := cmd.Flags().GetString(FlagLockIds)
			if err != nil {
				return err
			}

			lockIdStrs := strings.Split(lockIdsCombined, ",")
			lockIds := []uint64{}
			for _, lockIdStr := range lockIdStrs {
				lockId, err := strconv.ParseUint(lockIdStr, 10, 64)
				if err != nil {
					return err
				}
				lockIds = append(lockIds, lockId)
			}

			endEpoch, err := cmd.Flags().GetInt64(FlagEndEpoch)
			if err != nil {
				return err
			}

			res, err := queryClient.RewardsEst(cmd.Context(), &types.RewardsEstRequest{
				Owner:    owner, // owner is used only when lockIds are empty
				LockIds:  lockIds,
				EndEpoch: endEpoch,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(FlagOwner, "", "Owner to receive rewards, optionally used when lock-ids flag is NOT set")
	cmd.Flags().String(FlagLockIds, "", "the lock ids to receive rewards, when it is empty, all lock ids of the owner are used")
	cmd.Flags().Int64(FlagEndEpoch, 0, "the end epoch number to participate in rewards calculation")

	return cmd
}
