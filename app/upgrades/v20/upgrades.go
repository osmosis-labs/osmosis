package v20

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v19/app/keepers"
	"github.com/osmosis-labs/osmosis/v19/app/upgrades"
	cltypes "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v19/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Initialize the newly created param
		keepers.ConcentratedLiquidityKeeper.SetParam(ctx, cltypes.KeyUnrestrictedPoolCreatorWhitelist, []string{})

		// Initialize the new params in incentives for group creation.
		keepers.IncentivesKeeper.SetParam(ctx, incentivestypes.KeyGroupCreationFee, incentivestypes.DefaultGroupCreationFee)
		keepers.IncentivesKeeper.SetParam(ctx, incentivestypes.KeyCreatorWhitelist, []string{})

		migrationInfo, err := keepers.GAMMKeeper.GetAllMigrationInfo(ctx)
		if err != nil {
			return nil, err
		}

		// Create map from CL pools ID to CFMM pools ID
		clPoolsIDToCFMMPoolsID := make(map[uint64]uint64)
		for _, info := range migrationInfo.BalancerToConcentratedPoolLinks {
			clPoolsIDToCFMMPoolsID[info.ClPoolId] = info.BalancerPoolId
		}

		distrInfo := keepers.PoolIncentivesKeeper.GetDistrInfo(ctx)
		for i, distrRecord := range distrInfo.Records {
			gaugeID := distrRecord.GaugeId
			gauge, err := keepers.IncentivesKeeper.GetGaugeByID(ctx, gaugeID)
			if err != nil {
				return nil, err
			}

			if gauge.DistributeTo.LockQueryType != lockuptypes.NoLock {
				continue
			}

			incentivesEpochDuration := keepers.IncentivesKeeper.GetEpochInfo(ctx).Duration
			poolID, err := keepers.PoolIncentivesKeeper.GetPoolIdFromGaugeId(ctx, gaugeID, incentivesEpochDuration)
			if err != nil {
				return nil, err
			}

			associatedGammPoolID, ok := clPoolsIDToCFMMPoolsID[poolID]
			if !ok {
				return nil, errors.New("pool id not found in migration info")
			}

			// Get incentives module account
			incentivesModuleAcc := keepers.AccountKeeper.GetModuleAccount(ctx, incentivestypes.ModuleName)

			// Create group
			groupedPoolIDs := []uint64{poolID, associatedGammPoolID}
			groupGaugeID, err := keepers.IncentivesKeeper.CreateGroupNoWeightSync(ctx, sdk.NewCoins(), incentivestypes.PerpetualNumEpochsPaidOver, incentivesModuleAcc.GetAddress(), groupedPoolIDs)
			if err != nil {
				return nil, err
			}

			// Replace gauge ID in the distribution record
			distrInfo.Records[i].GaugeId = groupGaugeID
		}

		return migrations, nil
	}
}
