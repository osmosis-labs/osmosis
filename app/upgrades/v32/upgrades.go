package v32

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v31/app/keepers"
	"github.com/osmosis-labs/osmosis/v31/app/upgrades"
	protorevtypes "github.com/osmosis-labs/osmosis/v31/x/protorev/types"
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

		if err := moveProtorevNonBaseDenomsToSubDAO(sdkCtx, keepers); err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

// moveProtorevNonBaseDenomsToSubDAO sends every protorev module balance that is not a
// protorev base denom to the liquidity sub-DAO. Base denoms are left in place because
// the daily epoch hook already distributes them (developer fee, burn, community pool);
// everything else accumulates in the module account with no way out.
//
// Each denom is sent in its own cache context so a single denom that refuses to move
// (e.g. a tokenfactory before-send hook that errors) is logged and skipped rather than
// halting the upgrade.
func moveProtorevNonBaseDenomsToSubDAO(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	recipient, err := sdk.AccAddressFromBech32(LiquiditySubDAOAddress)
	if err != nil {
		return err
	}

	baseDenoms, err := keepers.ProtoRevKeeper.GetAllBaseDenoms(ctx)
	if err != nil {
		return err
	}
	isBaseDenom := make(map[string]bool, len(baseDenoms))
	for _, baseDenom := range baseDenoms {
		isBaseDenom[baseDenom.Denom] = true
	}

	moduleAddr := keepers.AccountKeeper.GetModuleAddress(protorevtypes.ModuleName)
	balances := keepers.BankKeeper.GetAllBalances(ctx, moduleAddr)

	moved := sdk.NewCoins()
	for _, coin := range balances {
		if isBaseDenom[coin.Denom] {
			continue
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			return keepers.BankKeeper.SendCoinsFromModuleToAccount(cacheCtx, protorevtypes.ModuleName, recipient, sdk.NewCoins(coin))
		})
		if err != nil {
			ctx.Logger().Error("v32: failed to move protorev balance to liquidity sub-DAO", "coin", coin.String(), "error", err)
			continue
		}
		moved = moved.Add(coin)
	}

	ctx.Logger().Info("v32: moved protorev non-base-denom balances to liquidity sub-DAO",
		"recipient", LiquiditySubDAOAddress, "denoms_moved", len(moved), "coins", moved.String())
	return nil
}
