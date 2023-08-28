package v19

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	gammtypes "github.com/osmosis-labs/osmosis/v19/x/gamm/types"

	"github.com/osmosis-labs/osmosis/v19/app/keepers"
	"github.com/osmosis-labs/osmosis/v19/app/upgrades"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"

	v18 "github.com/osmosis-labs/osmosis/v19/app/upgrades/v18"
)

const lastPoolToCorrect = v18.FirstCLPoolId - 1

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

		for id := 1; id <= lastPoolToCorrect; id++ {
			resetSuperfluidSumtree(keepers, ctx, uint64(id))
		}

		// Move the current authorized quote denoms from the concentrated liquidity params to the pool manager params.
		// This needs to be moved because the pool manager requires access to these denoms to determine if the taker fee should
		// be swapped into OSMO or not. The concentrated liquidity module already requires access to the pool manager keeper,
		// so the right move in this case is to move this parameter upwards in order to prevent circular dependencies.
		// TODO: In v20 upgrade handler, delete this param from the concentrated liquidity params.
		currentConcentratedLiquidityParams := keepers.ConcentratedLiquidityKeeper.GetParams(ctx)
		defaultPoolManagerParams := poolmanagertypes.DefaultParams()
		defaultPoolManagerParams.AuthorizedQuoteDenoms = currentConcentratedLiquidityParams.AuthorizedQuoteDenoms
		defaultPoolManagerParams.TakerFeeParams.DefaultTakerFee = sdk.ZeroDec()
		keepers.PoolManagerKeeper.SetParams(ctx, defaultPoolManagerParams)

		return migrations, nil
	}
}

func resetSuperfluidSumtree(keepers *keepers.AppKeepers, ctx sdk.Context, id uint64) {
	denom := gammtypes.GetPoolShareDenom(id)
	keepers.LockupKeeper.RebuildSuperfluidAccumulationStoresForDenom(ctx, denom)
}
