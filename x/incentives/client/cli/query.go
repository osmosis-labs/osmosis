package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v13/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
)

// GetQueryCmd returns the query commands for this module.
func GetQueryCmd() *cobra.Command {
	// group incentives queries under a subcommand
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	cmd.AddCommand(
		GetCmdGauges(),
		GetCmdToDistributeCoins(),
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
	return osmocli.SimpleQueryCmd[*types.GaugesRequest](
		"gauges",
		"Query available gauges",
		`{{.Short}}`,
		types.ModuleName, types.NewQueryClient,
	)
}

// GetCmdToDistributeCoins returns coins that are going to be distributed.
func GetCmdToDistributeCoins() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.ModuleToDistributeCoinsRequest](
		"to-distribute-coins",
		"Query coins that is going to be distributed",
		`{{.Short}}`,
		types.ModuleName, types.NewQueryClient,
	)
}

// GetCmdGaugeByID returns a gauge by ID.
func GetCmdGaugeByID() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.GaugeByIDRequest](
		"gauge-by-id [id]",
		"Query gauge by id.",
		`{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} gauge-by-id 1
`, types.ModuleName, types.NewQueryClient)
}

// GetCmdActiveGauges returns active gauges.
func GetCmdActiveGauges() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.ActiveGaugesRequest](
		"active-gauges",
		"Query active gauges",
		`{{.Short}}`,
		types.ModuleName, types.NewQueryClient,
	)
}

// GetCmdActiveGaugesPerDenom returns active gauges for a specified denom.
func GetCmdActiveGaugesPerDenom() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.ActiveGaugesPerDenomRequest](
		"active-gauges-per-denom [denom]",
		"Query active gauges per denom",
		`{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} active-gauges-per-denom gamm/pool/1`,
		types.ModuleName, types.NewQueryClient,
	)
}

// GetCmdUpcomingGauges returns scheduled gauges.
func GetCmdUpcomingGauges() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.UpcomingGaugesRequest](
		"upcoming-gauges",
		"Query upcoming gauges",
		`{{.Short}}`,
		types.ModuleName, types.NewQueryClient,
	)
}

// GetCmdUpcomingGaugesPerDenom returns scheduled gauges for specified denom..
func GetCmdUpcomingGaugesPerDenom() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.UpcomingGaugesPerDenomRequest](
		"upcoming-gauges-per-denom [denom]",
		"Query scheduled gauges per denom",
		`{{.Short}}`,
		types.ModuleName, types.NewQueryClient,
	)
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
			var res *lockuptypes.AccountLockedLongerDurationResponse
			ownerLocks := []uint64{}
			lockIds := []uint64{}

			owner, err := cmd.Flags().GetString(FlagOwner)
			if err != nil {
				return err
			}

			lockIdsCombined, err := cmd.Flags().GetString(FlagLockIds)
			if err != nil {
				return err
			}
			lockIdStrs := strings.Split(lockIdsCombined, ",")

			endEpoch, err := cmd.Flags().GetInt64(FlagEndEpoch)
			if err != nil {
				return err
			}

			// if user doesn't provide at least one of the lock ids or owner, we don't have enough information to proceed.
			if lockIdsCombined == "" && owner == "" {
				return fmt.Errorf("either one of owner flag or lock IDs must be provided")

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

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			if owner != "" {
				queryClientLockup := lockuptypes.NewQueryClient(clientCtx)

				res, err = queryClientLockup.AccountLockedLongerDuration(cmd.Context(), &lockuptypes.AccountLockedLongerDurationRequest{Owner: owner, Duration: time.Millisecond})
				if err != nil {
					return err
				}
				for _, lockId := range res.Locks {
					ownerLocks = append(ownerLocks, lockId.ID)
				}
			}

			// TODO: Fix accumulation store bug. For now, we return a graceful error when attempting to query bugged gauges
			rewardsEstimateResult, err := queryClient.RewardsEst(cmd.Context(), &types.RewardsEstRequest{
				Owner:    owner, // owner is used only when lockIds are empty
				LockIds:  lockIds,
				EndEpoch: endEpoch,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(rewardsEstimateResult)
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
