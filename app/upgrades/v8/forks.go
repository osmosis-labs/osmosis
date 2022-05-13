package v8

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	superfluidkeeper "github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
)

// RunForkLogic executes height-gated on-chain fork logic for the Osmosis v3
// upgrade.
func RunForkLogic(ctx sdk.Context, superfluid *superfluidkeeper.Keeper) {
	ctx.Logger().Info("Applying Osmosis vx upgrade." +
		"Allowing direct unbonding for whitelisted pools")
	RegisterWhitelistedDirectUnbondPools(ctx, superfluid)
}

// RegisterWhitelistedDirectUnbondPools registers pools that are allowed to unpool
// https://www.mintscan.io/osmosis/proposals/226
func RegisterWhitelistedDirectUnbondPools(ctx sdk.Context, superfluid *superfluidkeeper.Keeper) {
	// TODO: Get from proposal
	whitelistedPoolShares := []int64{1, 2, 3, 4}
	unpoolAllowedPools := superfluid.GetUnpoolAllowedPools(ctx)

	for _, whitelistedPool := range whitelistedPoolShares {
		unpoolAllowedPools = append(unpoolAllowedPools, uint64(whitelistedPool))
	}

	superfluid.SetUnpoolAllowedPools(ctx, unpoolAllowedPools)
}
