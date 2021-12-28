package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/app/keepers"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
)

func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator,
	keepers *keepers.AppKeepers) upgradetypes.UpgradeHandler {
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
		keepers.GAMMKeeper.SetParams(ctx, gammtypes.NewParams(sdk.Coins{sdk.NewInt64Coin("uosmo", 1)})) // 1 uOSMO
		// execute prop12. See implementation in
		Prop12(ctx, keepers.BankKeeper, keepers.DistrKeeper)
		return vm, nil
	}
}
