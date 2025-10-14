package v20

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v31/app/keepers"
	"github.com/osmosis-labs/osmosis/v31/app/upgrades"
	cltypes "github.com/osmosis-labs/osmosis/v31/x/concentrated-liquidity/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v31/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v31/x/lockup/types"
	poolincenitvestypes "github.com/osmosis-labs/osmosis/v31/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v31/x/poolmanager/types"
)

type IncentivizedCFMMDirectWhenMigrationLinkPresentError struct {
	CFMMPoolID         uint64
	ConcentratedPoolID uint64
	CFMMGaugeID        uint64
}

var emptySlice = []string{}

func (e IncentivizedCFMMDirectWhenMigrationLinkPresentError) Error() string {
	return fmt.Sprintf("CFMM gauge ID (%d) incentivized CFMM pool (%d) directly when migration link is present with concentrated pool (%d)", e.CFMMGaugeID, e.CFMMPoolID, e.ConcentratedPoolID)
}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Initialize the newly created param
		keepers.ConcentratedLiquidityKeeper.SetParam(ctx, cltypes.KeyUnrestrictedPoolCreatorWhitelist, emptySlice)

		// Initialize the new params in incentives for group creation.
		keepers.IncentivesKeeper.SetParam(ctx, incentivestypes.KeyGroupCreationFee, incentivestypes.DefaultGroupCreationFee)
		keepers.IncentivesKeeper.SetParam(ctx, incentivestypes.KeyCreatorWhitelist, emptySlice)

		// Initialize new param in the poolmanager module with a whitelist allowing to bypass taker fees.
		keepers.PoolManagerKeeper.SetParam(ctx, poolmanagertypes.KeyReducedTakerFeeByWhitelist, emptySlice)

		// Converts pool incentive distribution records from concentrated gauges to group gauges.
		err = createGroupsForIncentivePairs(ctx, keepers)
		if err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

// createGroupsForIncentivePairs converts pool incentive distribution records from concentrated gauges to group gauges.
// The expected update is to convert concentrated gauges to group gauges iff
//   - migration record exists for the concentrated pool and another CFMM pool
//   - if migration between concentrated and CFMM exists, then the CFMM pool is not incentivized individually
//
// All other distribution records are not modified.
//
// The updated distribution records are saved in the store.
func createGroupsForIncentivePairs(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	// Create map from CL pools ID to CFMM pools ID
	// from migration records
	migrationInfo, err := keepers.GAMMKeeper.GetAllMigrationInfo(ctx)
	if err != nil {
		return err
	}

	poolIDMigrationRecordMap := make(map[uint64]uint64)
	for _, info := range migrationInfo.BalancerToConcentratedPoolLinks {
		poolIDMigrationRecordMap[info.ClPoolId] = info.BalancerPoolId
		poolIDMigrationRecordMap[info.BalancerPoolId] = info.ClPoolId
	}

	distrInfo := keepers.PoolIncentivesKeeper.GetDistrInfo(ctx)

	// For all incentive distribution records,
	// retrieve the gauge associated with the record
	// If gauge directs incentives to a concentrated pool AND the concentrated pool
	// is linked to balancer via migration map, create
	// a group gauge and replace it in the distribution record.
	// Note that if there is a concentrated pool that is not
	// linked to balancer, nothing is done.
	// Stableswap pools are expected to be silently ignored. We do not
	// expect any stableswap pool to be linked to concentrated.
	for i, distrRecord := range distrInfo.Records {
		gaugeID := distrRecord.GaugeId

		// Gauge with ID zero goes to community pool.
		if gaugeID == poolincenitvestypes.CommunityPoolDistributionGaugeID {
			continue
		}

		gauge, err := keepers.IncentivesKeeper.GetGaugeByID(ctx, gaugeID)
		if err != nil {
			return err
		}

		// At the time of v20 upgrade, we only have concentrated pools
		// that are linked to balancer. Concentrated gauges receive all rewards
		// and then retroactively update balancer.
		// Concentrated pools have NoLock Gauge associated with them.
		// As a result, we look for this specific type here.
		// If type mismatched, this is a CFMM pool gauge. In that case,
		// we continue to the next incentive record after validating
		// that there is no migration record present for this CFMM pool. That is
		// it is not incentivized individually when concentrated pool already retroactively
		// distributed rewards to it.
		if gauge.DistributeTo.LockQueryType != lockuptypes.NoLock {
			// Validate that if there is a migration record pair between a concentrated
			// and a cfmm pool, only concentrated is present in the distribution records.
			longestLockableDuration, err := keepers.PoolIncentivesKeeper.GetLongestLockableDuration(ctx)
			if err != nil {
				return err
			}
			cfmmPoolID, err := keepers.PoolIncentivesKeeper.GetPoolIdFromGaugeId(ctx, gaugeID, longestLockableDuration)
			if err != nil {
				return err
			}

			// If we had a migration record present and a balancer pool is still incentivized individually,
			// something went wrong. This is because the presence of migration record implies retroactive
			// incentive distribution from concentrated to balancer.
			linkedConcentratedPoolID, hasAssociatedConcentratedPoolLinked := poolIDMigrationRecordMap[cfmmPoolID]
			if hasAssociatedConcentratedPoolLinked {
				return IncentivizedCFMMDirectWhenMigrationLinkPresentError{
					CFMMPoolID:         cfmmPoolID,
					ConcentratedPoolID: linkedConcentratedPoolID,
					CFMMGaugeID:        gaugeID,
				}
			}

			// Validation passed. This was an individual CFMM pool with no link to concentrated
			// Silently skip it.
			continue
		}

		// Get PoolID associated with the given NoLock gauge ID
		// NoLock gauges are associated with an incentives epoch duration.
		incentivesEpochDuration := keepers.IncentivesKeeper.GetEpochInfo(ctx).Duration
		concentratedPoolID, err := keepers.PoolIncentivesKeeper.GetPoolIdFromGaugeId(ctx, gaugeID, incentivesEpochDuration)
		if err != nil {
			return err
		}

		associatedGammPoolID, ok := poolIDMigrationRecordMap[concentratedPoolID]
		if !ok {
			// There is no CFMM pool ID for the concentrated pool ID, continue to the next.
			continue
		}

		// Found concentrated and CFMM pools that are linked by
		// migration records. Create a Group for them
		groupedPoolIDs := []uint64{concentratedPoolID, associatedGammPoolID}
		groupGaugeID, err := keepers.IncentivesKeeper.CreateGroupAsIncentivesModuleAcc(ctx, incentivestypes.PerpetualNumEpochsPaidOver, groupedPoolIDs)
		if err != nil {
			return err
		}

		// Replace the gauge ID with the group gauge ID in the distribution records
		distrInfo.Records[i].GaugeId = groupGaugeID
	}

	keepers.PoolIncentivesKeeper.SetDistrInfo(ctx, distrInfo)

	return nil
}
