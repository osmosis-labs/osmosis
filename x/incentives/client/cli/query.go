package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
)

// GetQueryCmd returns the query commands for this module.
func GetQueryCmd() *cobra.Command {
	// group incentives queries under a subcommand
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

// GetCmdGauges returns all available gauges.
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

// GetCmdToDistributeCoins returns coins that are going to be distributed.
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

// GetCmdDistributedCoins returns coins that have been distributed so far.
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

// GetCmdGaugeByID returns a gauge by ID.
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

// GetCmdActiveGaugesPerDenom returns active gauges for a specified denom.
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

// GetCmdUpcomingGaugesPerDenom returns active gauges for a specified denom.
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

			var res *lockuptypes.AccountLockedLongerDurationResponse
			if owner != "" {
				queryClientLockup := lockuptypes.NewQueryClient(clientCtx)

				res, err = queryClientLockup.AccountLockedLongerDuration(cmd.Context(), &lockuptypes.AccountLockedLongerDurationRequest{Owner: owner, Duration: time.Millisecond})
				if err != nil {
					return err
				}
			} else {
				owner = ""
			}

			ownerLocks := []uint64{}

			for _, lockId := range res.Locks {
				ownerLocks = append(ownerLocks, lockId.ID)
			}

			lockIdStrs := strings.Split(lockIdsCombined, ",")
			lockIds := []uint64{}
			// if user doesn't provide at least one of the lock ids or owner, we don't have enough information to proceed.
			if lockIdsCombined == "" && owner == "" {
				return fmt.Errorf("if owner flag is not set, lock IDs must be provided")

				// if user provides lockIDs, use these lockIDs in our rewards estimation
			} else if owner == "" {
				for _, lockIdStr := range lockIdStrs {
					lockId, err := strconv.ParseUint(lockIdStr, 10, 64)
					if err != nil {
						return err
					}
					lockIds = append(lockIds, lockId)
				}

				// if no lockIDs are provided but an owner is provided, we query the rewards for all of the locks the owner has
			} else if lockIdsCombined == "" {
				lockIds = append(lockIds, ownerLocks...)
			}

			// if lockIDs are provided and an owner is provided, only query the lockIDs that are provided
			// if a lockID was provided and it doesn't belong to the owner, return an error
			if len(lockIds) != 0 && owner != "" {
				for _, lockId := range lockIds {
					validInputLockId := contains(ownerLocks, lockId)
					if !validInputLockId {
						return fmt.Errorf("lock-id %v does not belong to %v", lockId, owner)
					}
				}
			}

			endEpoch, err := cmd.Flags().GetInt64(FlagEndEpoch)
			if err != nil {
				return err
			}

			// TODO: Figure out why some lock ids are good and some causes "Error: rpc error: code = Unknown desc = panic message redacted to hide potentially sensitive system info: panic"
			// since owner checks have already been completed above, we switch the owner address to a random module account address since a blank owner panics.
			// we should find a better way to circumvent this address validity check
			if owner == "" {
				owner = "osmo14kjcwdwcqsujkdt8n5qwpd8x8ty2rys5rjrdjj"
			}
			res1, err1 := queryClient.RewardsEst(cmd.Context(), &types.RewardsEstRequest{
				Owner:    owner, // owner is used only when lockIds are empty
				LockIds:  lockIds,
				EndEpoch: endEpoch,
			})
			if err1 != nil {
				return err1
			}

			return clientCtx.PrintProto(res1)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(FlagOwner, "", "Owner to receive rewards, optionally used when lock-ids flag is NOT set")
	cmd.Flags().String(FlagLockIds, "", "the lock ids to receive rewards, when it is empty, all lock ids of the owner are used")
	cmd.Flags().Int64(FlagEndEpoch, 0, "the end epoch number to participate in rewards calculation")

	return cmd
}

func contains(s []uint64, value uint64) bool {
	for _, v := range s {
		if v == value {
			return true
		}
	}

	return false
}
