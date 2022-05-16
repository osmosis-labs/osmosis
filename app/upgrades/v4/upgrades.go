package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	gammkeeper "github.com/osmosis-labs/osmosis/v8/x/gamm/keeper"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	gammtypes "github.com/osmosis-labs/osmosis/v8/x/gamm/types"
)

func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator,
	bank bankkeeper.Keeper,
	distr *distrkeeper.Keeper,
	gamm *gammkeeper.Keeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// // Upgrade all of the lock storages
		// locks, err := app.LockupKeeper.GetLegacyPeriodLocks(ctx)
		// if err != nil {
		// 	panic(err)
		// }
		// // clear all lockup module locking / unlocking queue items
		// app.LockupKeeper.ClearAllLockRefKeys(ctx)
		// app.LockupKeeper.ClearAllAccumulationStores(ctx)

		// // reset all lock and references
		// if err := app.LockupKeeper.ResetAllLocks(ctx, locks); err != nil {
		// 	panic(err)
		// }

		// configure upgrade for gamm module's pool creation fee param add
		gamm.SetParams(ctx, gammtypes.NewParams(sdk.Coins{sdk.NewInt64Coin("uosmo", 1)})) // 1 uOSMO
		// execute prop12. See implementation in
		Prop12(ctx, bank, distr)
		return vm, nil
	}
}
