package v8

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammkeeper "github.com/osmosis-labs/osmosis/v8/x/gamm/keeper"
	poolincentiveskeeper "github.com/osmosis-labs/osmosis/v8/x/pool-incentives/keeper"
	superfluidkeeper "github.com/osmosis-labs/osmosis/v8/x/superfluid/keeper"
)

// RunForkLogic executes height-gated on-chain fork logic for the Osmosis v8
// upgrade.
func RunForkLogic(ctx sdk.Context, superfluid *superfluidkeeper.Keeper, poolincentives *poolincentiveskeeper.Keeper, gamm *gammkeeper.Keeper) {
	// Only proceed with v8 for mainnet, testnets need not adjust their pool incentives or unbonding.
	// https://github.com/osmosis-labs/osmosis/issues/1609
	if ctx.ChainID() != "osmosis-1" {
		return
	}

	for i := 0; i < 100; i++ {
		ctx.Logger().Info("I am upgrading to v8")
	}
	ctx.Logger().Info("Applying Osmosis v8 upgrade. Allowing direct unpooling for whitelisted pools")
	ctx.Logger().Info("Applying accelerated incentive updates per proposal 225")
	ApplyProp222Change(ctx, poolincentives)
	ApplyProp223Change(ctx, poolincentives)
	ApplyProp224Change(ctx, poolincentives)
	ctx.Logger().Info("Registering state change for whitelisted pools for unpooling per proposal 226 ")
	RegisterWhitelistedDirectUnbondPools(ctx, superfluid, gamm)
}
