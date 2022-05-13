package vx

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/app/keepers"
	gammkeeper "github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
)

// RunForkLogic executes height-gated on-chain fork logic for the Osmosis v3
// upgrade.
func RunForkLogic(ctx sdk.Context, keepers *keepers.AppKeepers) {
	ctx.Logger().Info("Applying Osmosis vx upgrade." +
		"Allowing direct unbonding for whitelisted pools")

}

// RegisterWhitelistedDirectUnbondPools registers pools that are allowed to unpool
func RegisterWhitelistedDirectUnbondPools(ctx sdk.Context, gamm *gammkeeper.Keeper) {
	whitelistedPoolShares := []int64{1, 2, 3, 4}
	unpoolAllowedPools := gamm.GetUnpoolAllowedPools(ctx)

	for _, whitelistedPool := range whitelistedPoolShares {
		unpoolAllowedPools = append(unpoolAllowedPools, uint64(whitelistedPool))
	}

	gamm.SetUnpoolAllowedPools(ctx, unpoolAllowedPools)
}

// FinishCurrentUnbondingLocks directly unbonds current unbonding locks in the whitelisted pools
func FinishCurrentUnbondingLocks(ctx sdk.Context, gamm *gammkeeper.Keeper) {

}
