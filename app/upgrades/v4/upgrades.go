package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	gammkeeper "github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

// CreateUpgradeHandler returns an x/upgrade handler for the Osmosis v4 on-chain
// upgrade. Namely, it executes:
//
// 1. Setting x/gamm parameters for pool creation
// 2. Executing prop 12
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bank bankkeeper.Keeper,
	distr *distrkeeper.Keeper,
	gamm *gammkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// configure upgrade for x/gamm module pool creation fee param
		gamm.SetParams(ctx, gammtypes.NewParams(sdk.Coins{sdk.NewInt64Coin("uosmo", 1)})) // 1 uOSMO

		Prop12(ctx, bank, distr)

		return vm, nil
	}
}
