package v22

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	tmtypes "github.com/cometbft/cometbft/types"

	"github.com/osmosis-labs/osmosis/v23/app/keepers"
	"github.com/osmosis-labs/osmosis/v23/app/upgrades"
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

		// Migrate legacy taker fee tracker to new taker fee tracker (for performance reasons)
		oldTakerFeeTrackerForStakers := keepers.PoolManagerKeeper.GetLegacyTakerFeeTrackerForStakers(ctx)
		for _, coin := range oldTakerFeeTrackerForStakers {
			err := keepers.PoolManagerKeeper.UpdateTakerFeeTrackerForStakersByDenom(ctx, coin.Denom, coin.Amount)
			if err != nil {
				return nil, err
			}
		}

		oldTakerFeeTrackerForCommunityPool := keepers.PoolManagerKeeper.GetLegacyTakerFeeTrackerForCommunityPool(ctx)
		for _, coin := range oldTakerFeeTrackerForCommunityPool {
			err := keepers.PoolManagerKeeper.UpdateTakerFeeTrackerForCommunityPoolByDenom(ctx, coin.Denom, coin.Amount)
			if err != nil {
				return nil, err
			}
		}

		// Properly register consensus params. In the process, change params as per:
		// https://www.mintscan.io/osmosis/proposals/705
		defaultConsensusParams := tmtypes.DefaultConsensusParams().ToProto()
		defaultConsensusParams.Block.MaxBytes = 5000000 // previously 10485760
		defaultConsensusParams.Block.MaxGas = 300000000 // previously 120000000
		keepers.ConsensusParamsKeeper.Set(ctx, &defaultConsensusParams)

		// Increase the tx size cost per byte to 20 to reduce the exploitability of bandwidth amplification problems.
		accountParams := keepers.AccountKeeper.GetParams(ctx)
		accountParams.TxSizeCostPerByte = 20 // Double from the default value of 10
		err = keepers.AccountKeeper.SetParams(ctx, accountParams)
		if err != nil {
			return nil, err
		}

		return migrations, nil
	}
}
