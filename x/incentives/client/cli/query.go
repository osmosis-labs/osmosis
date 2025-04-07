package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
)

// GetQueryCmd returns the query commands for this module.
func GetQueryCmd() *cobra.Command {
	// group incentives queries under a subcommand
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	qcGetter := types.NewQueryClient
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdGauges)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdToDistributeCoins)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdGaugeByID)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdActiveGauges)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdActiveGaugesPerDenom)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdUpcomingGauges)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdUpcomingGaugesPerDenom)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdAllGroups)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdAllGroupsGauges)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdAllGroupsWithGauge)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdGroupByGroupGaugeID)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdCurrentWeightByGroupGaugeID)
	cmd.AddCommand(
		osmocli.GetParams[*types.ParamsRequest](
			types.ModuleName, types.NewQueryClient),
	)
	cmd.AddCommand(GetCmdRewardsEst())

	return cmd
}

// GetCmdGauges returns all available gauges.
func GetCmdGauges() (*osmocli.QueryDescriptor, *types.GaugesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "gauges",
		Short: "Query all available gauges",
		Long:  "{{.Short}}",
	}, &types.GaugesRequest{}
}

// GetCmdToDistributeCoins returns coins that are going to be distributed.
func GetCmdToDistributeCoins() (*osmocli.QueryDescriptor, *types.ModuleToDistributeCoinsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "to-distribute-coins",
		Short: "Query coins that is going to be distributed",
		Long:  `{{.Short}}`,
	}, &types.ModuleToDistributeCoinsRequest{}
}

// GetCmdGaugeByID returns a gauge by ID.
func GetCmdGaugeByID() (*osmocli.QueryDescriptor, *types.GaugeByIDRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "gauge-by-id",
		Short: "Query gauge by id.",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} gauge-by-id 1
`,
	}, &types.GaugeByIDRequest{}
}

// GetCmdActiveGauges returns active gauges.
func GetCmdActiveGauges() (*osmocli.QueryDescriptor, *types.ActiveGaugesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "active-gauges",
		Short: "Query active gauges",
		Long:  `{{.Short}}`,
	}, &types.ActiveGaugesRequest{}
}

// GetCmdActiveGaugesPerDenom returns active gauges for a specified denom.
func GetCmdActiveGaugesPerDenom() (*osmocli.QueryDescriptor, *types.ActiveGaugesPerDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "active-gauges-per-denom",
		Short: "Query active gauges per denom",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} active-gauges-per-denom gamm/pool/1`,
	}, &types.ActiveGaugesPerDenomRequest{}
}

// GetCmdUpcomingGauges returns scheduled gauges.
func GetCmdUpcomingGauges() (*osmocli.QueryDescriptor, *types.UpcomingGaugesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "upcoming-gauges",
		Short: "Query upcoming gauges",
		Long:  `{{.Short}}`,
	}, &types.UpcomingGaugesRequest{}
}

// GetCmdUpcomingGaugesPerDenom returns scheduled gauges for specified denom..
func GetCmdUpcomingGaugesPerDenom() (*osmocli.QueryDescriptor, *types.UpcomingGaugesPerDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "upcoming-gauges-per-denom",
		Short: "Query scheduled gauges per denom",
		Long:  `{{.Short}}`,
	}, &types.UpcomingGaugesPerDenomRequest{}
}

// GetCmdCurrentWeightByGroupGaugeID returns current weight for each gauge respectively since the last epoch from a group gauge ID.
func GetCmdCurrentWeightByGroupGaugeID() (*osmocli.QueryDescriptor, *types.QueryCurrentWeightByGroupGaugeIDRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "current-weight-by-group-gauge-id",
		Short: "Query current incentives distribution weight since epoch for each gauge respectively from a group gauge ID",
		Long:  `{{.Short}}`,
	}, &types.QueryCurrentWeightByGroupGaugeIDRequest{}
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

func GetCmdAllGroups() (*osmocli.QueryDescriptor, *types.QueryAllGroupsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "all-groups",
		Short: "Query all groups",
		Long:  `{{.Short}}`,
	}, &types.QueryAllGroupsRequest{}
}

func GetCmdAllGroupsGauges() (*osmocli.QueryDescriptor, *types.QueryAllGroupsGaugesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "all-groups-gauges",
		Short: "Query all group gauges",
		Long:  `{{.Short}}`,
	}, &types.QueryAllGroupsGaugesRequest{}
}

func GetCmdAllGroupsWithGauge() (*osmocli.QueryDescriptor, *types.QueryAllGroupsWithGaugeRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "all-groups-with-gauge",
		Short: "Query all groups with their respective group gauge",
		Long:  `{{.Short}}`,
	}, &types.QueryAllGroupsWithGaugeRequest{}
}

func GetCmdGroupByGroupGaugeID() (*osmocli.QueryDescriptor, *types.QueryGroupByGroupGaugeIDRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "group-by-group-gauge-id [group-gauge-id]",
		Short: "Query a group it's respective group gauge ID",
		Long:  `{{.Short}}`,
	}, &types.QueryGroupByGroupGaugeIDRequest{}
}
