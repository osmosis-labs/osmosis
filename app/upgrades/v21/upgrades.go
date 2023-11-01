package v21

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v20/app/keepers"
	"github.com/osmosis-labs/osmosis/v20/app/upgrades"
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

		// Since we are now tracking all protocol rev, we set the accounting height to the current block height for each module
		// that generates protocol rev.
		keepers.PoolManagerKeeper.SetTakerFeeTrackerStartHeight(ctx, uint64(ctx.BlockHeight()))
		keepers.TxFeesKeeper.SetTxFeesTrackerStartHeight(ctx, uint64(ctx.BlockHeight()))
		// We start the cyclic arb tracker from the value it currently is at since it has been tracking since inception (without a start height).
		allCyclicArbProfits := keepers.ProtoRevKeeper.GetAllProfits(ctx)
		allCyclicArbProfitsCoins := osmoutils.ConvertCoinArrayToCoins(allCyclicArbProfits)
		keepers.ProtoRevKeeper.SetCyclicArbProfitTrackerValue(ctx, allCyclicArbProfitsCoins)
		keepers.ProtoRevKeeper.SetCyclicArbProfitTrackerStartHeight(ctx, uint64(ctx.BlockHeight()))

		return migrations, nil
	}
}
