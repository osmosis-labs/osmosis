package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v15/app/keepers"
	"github.com/osmosis-labs/osmosis/v15/app/upgrades"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanager "github.com/osmosis-labs/osmosis/v15/x/poolmanager"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// fromVM[cltypes.ModuleName] = 0
		// poolId, err := createCLPool(ctx, keepers.PoolManagerKeeper)
		// if err != nil {
		// 	panic(err)
		// }

		// err = migrateBalancerSharesToCLPool(ctx, keepers.BankKeeper, keepers.GAMMKeeper, poolId)
		// if err != nil {
		// 	panic(err)
		// }
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func createCLPool(ctx sdk.Context, poolManagerKeeper *poolmanager.Keeper) (uint64, error) {
	// use faucet acccount to create pool and pay for pool creation fee
	poolId, err := poolManagerKeeper.CreatePool(ctx, clmodel.NewMsgCreateConcentratedPool(
		sdk.MustAccAddressFromBech32("osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj"),
		"uosmo",
		"uion",
		uint64(1),
		sdk.NewInt(-1),
		sdk.MustNewDecFromStr("0.01"),
	))
	if err != nil {
		return 0, err
	}

	return poolId, nil
}

// runs migrationfrom pool #1 to the new CL pool.
// All shares are migrated as full range shares
func migrateBalancerSharesToCLPool(ctx sdk.Context, bankKeeper bankkeeper.Keeper, gammKeeper *gammkeeper.Keeper, newCLPoolID uint64) error {
	// manually set migration info
	migratingPools := []gammtypes.BalancerToConcentratedPoolLink{
		{
			BalancerPoolId: 1,
			ClPoolId:       newCLPoolID,
		},
	}
	gammKeeper.SetMigrationInfo(ctx, gammtypes.MigrationRecords{BalancerToConcentratedPoolLinks: migratingPools})

	// get sender + share amount for all position in pool 1, then iterate through all accounts,
	// migrating them to CL full range positions.
	balancerPoolShareDenom := gammtypes.GetPoolShareDenom(1)
	accountsBalances := bankKeeper.GetAccountsBalances(ctx)
	for _, accountBalance := range accountsBalances {
		// accountBalance.Coins.DenomsSubsetOf(balancerPoolShareCoins)
		balancerPoolShareAmt := accountBalance.Coins.AmountOf(balancerPoolShareDenom)
		if balancerPoolShareAmt.GT(sdk.ZeroInt()) {
			_, _, _, _, _, err := gammKeeper.MigrateFromBalancerToConcentrated(ctx, accountBalance.GetAddress(), sdk.NewCoin(balancerPoolShareDenom, balancerPoolShareAmt))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
