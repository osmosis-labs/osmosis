package v27

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

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

		sdkCtx := sdk.UnwrapSDKContext(ctx)

		err = InitializeConstitutionCollection(sdkCtx, *keepers.GovKeeper)
		if err != nil {
			sdkCtx.Logger().Error("Error initializing Constitution Collection:", "message", err.Error())
		}

		return migrations, nil
	}
}

// setting the default constitution for the chain
// this is in line with cosmos-sdk v5 gov migration: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/gov/migrations/v5/store.go#L57
func InitializeConstitutionCollection(ctx sdk.Context, govKeeper govkeeper.Keeper) error {
	return govKeeper.Constitution.Set(ctx, "This chain has no constitution.")
}
