package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v3 "github.com/osmosis-labs/osmosis/v7/app/upgrades/v3"
	v6 "github.com/osmosis-labs/osmosis/v7/app/upgrades/v6"
	v8 "github.com/osmosis-labs/osmosis/v7/app/upgrades/v8"
)

// BeginBlockForks is intended to be ran in
func BeginBlockForks(ctx sdk.Context, app *OsmosisApp) {
	switch ctx.BlockHeight() {
	case v3.UpgradeHeight:
		v3.RunForkLogic(ctx, app.GovKeeper, app.StakingKeeper)
	case v6.UpgradeHeight:
		v6.RunForkLogic(ctx)
	case v8.UpgradeHeight:
		v8.RunForkLogic(ctx, app.SuperfluidKeeper, app.PoolIncentivesKeeper)
	default:
		// do nothing
		return
	}
}
