package v5

import (
	ibcconnectiontypes "github.com/cosmos/ibc-go/v3/modules/core/03-connection/types"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v11/app/keepers"
	txfeestypes "github.com/osmosis-labs/osmosis/v11/x/txfees/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Set IBC updates from {inside SDK} to v1
		//
		// See: https://github.com/cosmos/ibc-go/blob/main/docs/migrations/ibc-migration-043.md#in-place-store-migrations
		keepers.IBCKeeper.ConnectionKeeper.SetParams(ctx, ibcconnectiontypes.DefaultParams())

		// Set all modules "old versions" to 1. Then the run migrations logic will
		// handle running their upgrade logics.
		fromVM := make(map[string]uint64)
		for moduleName := range mm.Modules {
			fromVM[moduleName] = 1
		}

		// EXCEPT Auth needs to run AFTER staking.
		//
		// See: https://github.com/cosmos/cosmos-sdk/issues/10591
		//
		// So we do this by making auth run last. This is done by setting auth's
		// consensus version to 2, running RunMigrations, then setting it back to 1,
		// and then running migrations again.
		fromVM[authtypes.ModuleName] = 2

		// Override versions for authz & bech32ibctypes module as to not skip their
		// InitGenesis for txfees module, we will override txfees ourselves.
		delete(fromVM, authz.ModuleName)
		delete(fromVM, bech32ibctypes.ModuleName)

		newVM, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Override txfees genesis here
		ctx.Logger().Info("Setting txfees module genesis with actual v5 desired genesis")
		feeTokens := InitialWhitelistedFeetokens(ctx, keepers.GAMMKeeper)
		keepers.TxFeesKeeper.InitGenesis(ctx, txfeestypes.GenesisState{
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
