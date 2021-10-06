package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
		GetCmdToDistributeCoins(),
		GetCmdDistributedCoins(),
		GetCmdGaugeByID(),
		GetCmdActiveGauges(),
		GetCmdUpcomingGauges(),
		GetCmdRewardsEst(),
		GetCmdRewards(),
		GetCurrentReward(),
		GetHistoricalReward(),
		GetPeriodLockReward(),
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

// GetCmdToDistributeCoins returns coins that is going to be distributed
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

// GetCmdDistributedCoins returns coins that are distributed so far
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

// GetCmdGaugeByID returns Gauge by id
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

// GetCmdActiveGauges returns active gauges
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

// GetCmdUpcomingGauges returns scheduled gauges
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

// GetCmdRewardsEst returns rewards estimation
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

// GetCmdRewards returns current estimate of accumulated rewards
func GetCmdRewards() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards",
		Short: "Query rewards estimation by combining both current and historical rewards",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query rewards estimation.

Example:
$ %s query incentives rewards [owner-addr] 
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

			owner := args[0]

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

			res, err := queryClient.Rewards(cmd.Context(), &types.RewardsRequest{
				Owner:    owner,
				LockIds:  lockIds,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(FlagLockIds, "", "the lock ids to receive rewards, when it is empty, all lock ids of the owner are used")

	return cmd
}

func GetCurrentReward() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current-reward",
		Short: "Query current reward",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query current reward.

Example:
$ %s query incentives current-reward [denom] [lockable-duration]
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			denom := args[0]

			duration, err := time.ParseDuration(args[1])
			if err != nil {
				return err
			}

			res, err := queryClient.CurrentReward(cmd.Context(), &types.CurrentRewardRequest{
				Denom:             denom,
				LockableDurations: duration,
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

func GetHistoricalReward() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "historical-reward",
		Short: "Query historical reward",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query historical reward.

Example:
$ %s query incentives historical-reward [denom] [lockable-duration] [period]
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			denom := args[0]

			duration, err := time.ParseDuration(args[1])
			if err != nil {
				return err
			}

			period, err := strconv.ParseInt(args[2], 10, 64)

			res, err := queryClient.HistoricalReward(cmd.Context(), &types.HistoricalRewardRequest{
				Denom:             denom,
				LockableDurations: duration,
				Period:            period,
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

func GetPeriodLockReward() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "period-lock-reward",
		Short: "Query period lock reward",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query period lock reward.

Example:
$ %s query incentives period-lock-reward [id]
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

			res, err := queryClient.PeriodLockReward(cmd.Context(), &types.PeriodLockRewardRequest{
				Id: id,
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
