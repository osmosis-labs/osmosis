package v25

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	slashing "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	"github.com/osmosis-labs/osmosis/v24/app/keepers"
	"github.com/osmosis-labs/osmosis/v24/app/upgrades"
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

		// Now that all deprecated historical TWAPs have been pruned via v24, we can delete is isPruning state entry as well
		keepers.TwapKeeper.DeleteDeprecatedHistoricalTWAPsIsPruning(ctx)

		// Reset missed blocks counter for all validators
		resetMissedBlocksCounter(ctx, keepers.SlashingKeeper)

		return migrations, nil
	}
}

// resetMissedBlocksCounter resets the missed blocks counter for all validators back to zero.
// This corrects a mistake that was overlooked in v24, where we cleared all missedBlocks but did not reset the counter.
func resetMissedBlocksCounter(ctx sdk.Context, slashingKeeper *slashing.Keeper) {
	// Iterate over all validators signing info
	slashingKeeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		missedBlocks, err := slashingKeeper.GetValidatorMissedBlocks(ctx, address)
		if err != nil {
			panic(err)
		}

		// Reset missed blocks counter
		info.MissedBlocksCounter = int64(len(missedBlocks))
		slashingKeeper.SetValidatorSigningInfo(ctx, address, info)

		return false
	})
}
