package v27

import (
	"context"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/osmosis-labs/osmosis/v27/app/keepers"
	"github.com/osmosis-labs/osmosis/v27/app/upgrades"
	stablestakingincentvicestypes "github.com/osmosis-labs/osmosis/v27/x/stable-staking-incentives/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v27/x/txfees/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		ctx.Logger().Warn("Run migrations")

		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		newVM, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			ctx.Logger().Error("❌ Migration failed:", "error", err)
			return nil, err
		}

		ctx.Logger().Warn("Run init genesis for TxFees module")
		initGenParams := txfeestypes.DefaultGenesis()
		keepers.TxFeesKeeper.InitGenesis(ctx, *initGenParams)

		ctx.Logger().Warn("Run init genesis for StableStakingIncentvice module")
		initGenStableParams := stablestakingincentvicestypes.NewGenesisState(stablestakingincentvicestypes.NewParams(DistributionContractAddress))
		keepers.StableStakingIncentivesKeeper.InitGenesis(ctx, initGenStableParams)

		ctx.Logger().Warn("Migration completed!")
		return newVM, nil
	}
}
