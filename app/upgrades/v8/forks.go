package v8

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/app/keepers"
)

// RunForkLogic executes height-gated on-chain fork logic for the Osmosis v3
// upgrade.
func RunForkLogic(ctx sdk.Context, appKeepers *keepers.AppKeepers) {
	ctx.Logger().Info("Applying Osmosis vx upgrade." +
		"Allowing direct unbonding for whitelisted pools")
	ctx.Logger().Info("Applying accelerated incentive updates per proposal 225")
	ApplyProp222Change(ctx, appKeepers.PoolIncentivesKeeper)
	ApplyProp223Change(ctx, appKeepers.PoolIncentivesKeeper)
	ApplyProp224Change(ctx, appKeepers.PoolIncentivesKeeper)
	RegisterWhitelistedDirectUnbondPools(ctx, appKeepers)
}

// RegisterWhitelistedDirectUnbondPools registers pools that are allowed to unpool
// https://www.mintscan.io/osmosis/proposals/226
// osmosisd q gov proposal 226
func RegisterWhitelistedDirectUnbondPools(ctx sdk.Context, appKeepers *keepers.AppKeepers) {
	// These are the pools listed in the proposal. Proposal raw text for the listing of UST pools:
	// 	The list of pools affected are defined below:
	// #560 (UST/OSMO)
	// #562 (UST/LUNA)
	// #567 (UST/EEUR)
	// #578 (UST/XKI)
	// #592 (UST/BTSG)
	// #610 (UST/CMDX)
	// #612 (UST/XPRT)
	// #615 (UST/LUM)
	// #642 (UST/UMEE)
	// #679 (4Pool)
	whitelistedPoolShares := []int64{560, 562, 567, 578, 592, 610, 612, 615, 642, 679}
	unpoolAllowedPools := appKeepers.SuperfluidKeeper.GetUnpoolAllowedPools(ctx)

	for _, whitelistedPool := range whitelistedPoolShares {
		unpoolAllowedPools = append(unpoolAllowedPools, uint64(whitelistedPool))
	}

	appKeepers.SuperfluidKeeper.SetUnpoolAllowedPools(ctx, unpoolAllowedPools)
}
