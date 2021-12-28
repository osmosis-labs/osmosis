package v5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/authz"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/app/keepers"

	"github.com/osmosis-labs/osmosis/x/txfees"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v2/modules/core/03-connection/types"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"
	txfeestypes "github.com/osmosis-labs/osmosis/x/txfees/types"
)

func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator,
	keepers *keepers.AppKeepers) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Set IBC updates from {inside SDK} to v1
		// https://github.com/cosmos/ibc-go/blob/main/docs/migrations/ibc-migration-043.md#in-place-store-migrations
		keepers.IBCKeeper.ConnectionKeeper.SetParams(ctx, ibcconnectiontypes.DefaultParams())

		totalLiquidity := keepers.GAMMKeeper.GetLegacyTotalLiquidity(ctx)
		keepers.GAMMKeeper.DeleteLegacyTotalLiquidity(ctx)
		keepers.GAMMKeeper.SetTotalLiquidity(ctx, totalLiquidity)

		// Set all modules "old versions" to 1.
		// Then the run migrations logic will handle running their upgrade logics
		fromVM := make(map[string]uint64)
		for moduleName := range mm.Modules {
			fromVM[moduleName] = 1
		}
		// EXCEPT Auth needs to run _after_ staking (https://github.com/cosmos/cosmos-sdk/issues/10591),
		// and it seems bank as well (https://github.com/provenance-io/provenance/blob/407c89a7d73854515894161e1526f9623a94c368/app/upgrades.go#L86-L122).
		// So we do this by making auth run last.
		// This is done by setting auth's consensus version to 2, running RunMigrations,
		// then setting it back to 1, and then running migrations again.
		fromVM[authtypes.ModuleName] = 2

		// override versions for authz & bech32ibctypes module as to not skip their InitGenesis
		// for txfees module, we will override txfees ourselves.
		delete(fromVM, authz.ModuleName)
		delete(fromVM, bech32ibctypes.ModuleName)

		newVM, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Override txfees genesis here
		ctx.Logger().Info("Setting txfees module genesis with actual v5 desired genesis")
		feeTokens := InitialWhitelistedFeetokens(ctx, keepers.GAMMKeeper)
		txfees.InitGenesis(ctx, *keepers.TxFeesKeeper, txfeestypes.GenesisState{
			Basedenom: keepers.StakingKeeper.BondDenom(ctx),
			Feetokens: feeTokens,
		})

		// now update auth version back to v1, to run auth migration last
		newVM[authtypes.ModuleName] = 1

		ctx.Logger().Info("Now running migrations just for auth, to get auth migration to be last. " +
			"(CC https://github.com/cosmos/cosmos-sdk/issues/10591)")
		return mm.RunMigrations(ctx, configurator, newVM)
	}
}
