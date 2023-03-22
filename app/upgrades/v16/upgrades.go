package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/osmosis-labs/osmosis/v15/app/keepers"
	"github.com/osmosis-labs/osmosis/v15/app/upgrades"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		err := mintManyManyCoins(ctx, keepers.BankKeeper, keepers.ConcentratedLiquidityKeeper)
		if err != nil {
			panic(err)
		}
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func mintManyManyCoins(ctx sdk.Context, bankKeeper bankkeeper.Keeper, clKeeper *cl.Keeper) error {
	ctx.Logger().Info("Starting creating single full range position")

	faucetAddress := sdk.MustAccAddressFromBech32("osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj")

	// first mint coins to the gamm module
	coinsToMint := sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(999999802118068)), sdk.NewCoin("ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", sdk.NewInt(999999802118068)), sdk.NewCoin("uion", sdk.NewInt(999999802118068)))
	err := bankKeeper.MintCoins(ctx, gammtypes.ModuleName, coinsToMint)
	if err != nil {
		return err
	}

	ctx.Logger().Info("Finsihed minting")
	// now send the minted coins from bank module to the faucet account
	err = bankKeeper.SendCoinsFromModuleToAccount(ctx, gammtypes.ModuleName, faucetAddress, coinsToMint)
	if err != nil {
		return err
	}

	return nil
}
