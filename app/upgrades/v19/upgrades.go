package v19

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	gammtypes "github.com/osmosis-labs/osmosis/v25/x/gamm/types"

	"github.com/osmosis-labs/osmosis/v25/app/keepers"
	"github.com/osmosis-labs/osmosis/v25/app/upgrades"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"

	v18 "github.com/osmosis-labs/osmosis/v25/app/upgrades/v18"
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

		defaultPoolManagerParams := poolmanagertypes.DefaultParams()
		defaultPoolManagerParams.TakerFeeParams.DefaultTakerFee = osmomath.ZeroDec()
		keepers.PoolManagerKeeper.SetParams(ctx, defaultPoolManagerParams)

		return migrations, nil
	}
}

func resetSuperfluidSumtree(keepers *keepers.AppKeepers, ctx sdk.Context, id uint64) {
	denom := gammtypes.GetPoolShareDenom(id)
	keepers.LockupKeeper.RebuildSuperfluidAccumulationStoresForDenom(ctx, denom)
}
