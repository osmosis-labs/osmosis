package v7

import (
	"context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/v30/app/keepers"
	"github.com/osmosis-labs/osmosis/v30/app/upgrades"
	lockupkeeper "github.com/osmosis-labs/osmosis/v30/x/lockup/keeper"
	superfluidtypes "github.com/osmosis-labs/osmosis/v30/x/superfluid/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		// Set wasm old version to 1 if we want to call wasm's InitGenesis ourselves
		// in this upgrade logic ourselves.
		//
		// vm[wasm.ModuleName] = wasm.ConsensusVersion
		//
		// Otherwise we run this, which will run wasm.InitGenesis(wasm.DefaultGenesis())
		// and then override it after.
		newVM, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return newVM, err
		}

		// Since we provide custom DefaultGenesis (privileges StoreCode) in
		// app/genesis.go rather than the wasm module, we need to set the params
		// here when migrating (is it is not customized).
		params := keepers.WasmKeeper.GetParams(ctx)
		params.CodeUploadAccess = wasmtypes.AllowNobody

		//nolint:errcheck
		keepers.WasmKeeper.SetParams(ctx, params)

		// Merge similar duration lockups
		ctx.Logger().Info("Merging lockups for similar durations")
		lockupkeeper.MergeLockupsForSimilarDurations(
			ctx, *keepers.LockupKeeper, keepers.AccountKeeper,
			lockupkeeper.BaselineDurations, lockupkeeper.HourDuration,
		)

		ctx.Logger().Info("Migration for superfluid staking")

		superfluidAsset := superfluidtypes.SuperfluidAsset{
			Denom:     "gamm/pool/1",
			AssetType: superfluidtypes.SuperfluidAssetTypeLPShare,
		}
		if err := keepers.SuperfluidKeeper.AddNewSuperfluidAsset(ctx, superfluidAsset); err != nil {
			return newVM, err
		}

		// N.B.: This is left for historic reasons.
		// After the v7 upgrade, there was no need for this function anymore so it was removed.
		// // Set the supply offset from the developer vesting account
		// if err := keepers.MintKeeper.SetInitialSupplyOffsetDuringMigration(ctx); err != nil {
		// 	panic(err)
		// }

		return newVM, err
	}
}
