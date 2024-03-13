package v24

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	cwpooltypes "github.com/osmosis-labs/osmosis/v23/x/cosmwasmpool/types"

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

		// We no longer use the base denoms array and instead use the repeated base denoms field for performance reasons.
		// We retrieve the old base denoms array from the KVStore, delete the array from the KVStore, and set them as a repeated field in the new KVStore.
		baseDenoms, err := keepers.ProtoRevKeeper.DeprecatedGetAllBaseDenoms(ctx)
		if err != nil {
			return nil, err
		}
		keepers.ProtoRevKeeper.DeprecatedDeleteBaseDenoms(ctx)
		err = keepers.ProtoRevKeeper.SetBaseDenoms(ctx, baseDenoms)
		if err != nil {
			return nil, err
		}

		// Now that the TWAP keys are refactored, we can delete all time indexed TWAPs
		// since we only need the pool indexed TWAPs.
		keepers.TwapKeeper.DeleteAllHistoricalTimeIndexedTWAPs(ctx)

		// White Whale uploaded a broken contract. They later migrated cwpool via the governance
		// proposal in x/cosmwasmpool
		// However, there was a problem in the migration logic where the CosmWasmpool state CodeId  did not get updated.
		// As a result, the CodeID for the contract that is tracked in x/wasmd  was migrated correctly. However, the code ID that we track in the x/cosmwasmpool  state did not.
		// Therefore, we should perform a migration for each of the harcoded white whale pools.
		poolIds := []uint64{1463, 1462, 1461}
		for _, poolId := range poolIds {
			pool, err := keepers.CosmwasmPoolKeeper.GetPool(ctx, poolId)
			if err != nil {
				return nil, err
			}
			cwPool, ok := pool.(cwpooltypes.CosmWasmExtension)
			if !ok {
				ctx.Logger().Error("Pool has incorrect type", "poolId", poolId, "pool", pool)
				return nil, cwpooltypes.InvalidPoolTypeError{
					ActualPool: pool,
				}
			}
			if cwPool.GetCodeId() != 503 {
				ctx.Logger().Error("Pool has incorrect code id", "poolId", poolId, "codeId", cwPool.GetCodeId())
				return nil, cwpooltypes.InvalidPoolTypeError{
					ActualPool: pool,
				}
			}
			cwPool.SetCodeId(572)
			keepers.CosmwasmPoolKeeper.SetPool(ctx, cwPool)
		}

		return migrations, nil
	}
}
