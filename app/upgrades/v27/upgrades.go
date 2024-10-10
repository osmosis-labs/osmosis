package v27

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/v26/app/keepers"
	"github.com/osmosis-labs/osmosis/v26/app/upgrades"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}
		bk := keepers.BankKeeper

		// Get the old offset from the old key: []byte{0x88}
		offsetOld := bk.GetSupplyOffsetOld(ctx, OsmoToken)

		// Add the old offset to the new key: collection.NewPrefix(88)
		bk.AddSupplyOffset(ctx, OsmoToken, offsetOld)

		// Remove the old key: []byte{0x88}
		bk.RemoveOldSupplyOffset(ctx, OsmoToken)

		return migrations, nil
	}
}
