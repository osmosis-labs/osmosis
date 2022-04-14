package v7

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	epochskeeper "github.com/osmosis-labs/osmosis/v7/x/epochs/keeper"
	lockupkeeper "github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
	mintkeeper "github.com/osmosis-labs/osmosis/v7/x/mint/keeper"
	superfluidkeeper "github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
	superfluidtypes "github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	wasmKeeper *wasm.Keeper,
	superfluidKeeper *superfluidkeeper.Keeper,
	epochsKeeper *epochskeeper.Keeper,
	lockupKeeper *lockupkeeper.Keeper,
	mintKeeper *mintkeeper.Keeper,
	accountKeeper *authkeeper.AccountKeeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
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
		params := wasmKeeper.GetParams(ctx)
		params.CodeUploadAccess = wasmtypes.AllowNobody
		wasmKeeper.SetParams(ctx, params)

		// Merge similar duration lockups
		ctx.Logger().Info("Merging lockups for similar durations")
		lockupkeeper.MergeLockupsForSimilarDurations(
			ctx, *lockupKeeper, accountKeeper,
			lockupkeeper.BaselineDurations, lockupkeeper.HourDuration,
		)

		ctx.Logger().Info("Migration for superfluid staking")

		superfluidAsset := superfluidtypes.SuperfluidAsset{
			Denom:     "gamm/pool/1",
			AssetType: superfluidtypes.SuperfluidAssetTypeLPShare,
		}
		superfluidKeeper.AddNewSuperfluidAsset(ctx, superfluidAsset)

		// Set the supply offset from the developer vesting account
		mintkeeper.SetInitialSupplyOffsetDuringMigration(ctx, *mintKeeper)

		return newVM, err
	}
}
