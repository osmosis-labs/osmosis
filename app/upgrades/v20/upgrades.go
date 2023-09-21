package v20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v19/app/keepers"
	"github.com/osmosis-labs/osmosis/v19/app/upgrades"
	incentivetypes "github.com/osmosis-labs/osmosis/v19/x/incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
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

		createPoolFee := sdk.NewCoins(sdk.NewCoin(USDCaxlDenom, sdk.NewInt(100000000))) // 100 USDC
		keepers.PoolManagerKeeper.SetParam(ctx, poolmanagertypes.KeyPoolCreationFee, createPoolFee)

		createGaugeFee := sdk.NewCoins(sdk.NewCoin(USDCaxlDenom, sdk.NewInt(50000000))) // 50 USDC
		keepers.IncentivesKeeper.SetParam(ctx, incentivetypes.KeyCreateGaugeFee, createGaugeFee)

		addToGaugeFee := sdk.NewCoins(sdk.NewCoin(USDCaxlDenom, sdk.NewInt(25000000))) // 25 USDC
		keepers.IncentivesKeeper.SetParam(ctx, incentivetypes.KeyAddToGaugeFee, addToGaugeFee)

		return migrations, nil
	}
}
