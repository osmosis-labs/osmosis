package v8

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gammkeeper "github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	superfluidkeeper "github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
)

// RegisterWhitelistedDirectUnbondPools registers pools that are allowed to unpool
// https://www.mintscan.io/osmosis/proposals/226
// osmosisd q gov proposal 226
func RegisterWhitelistedDirectUnbondPools(ctx sdk.Context, superfluid *superfluidkeeper.Keeper, gamm *gammkeeper.Keeper) {
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
	whitelistedPoolShares := []uint64{560, 562, 567, 578, 592, 610, 612, 615, 642, 679}

	// Consistency check that each whitelisted pool contains UST
	for _, whitelistedPool := range whitelistedPoolShares {
		if err := CheckPoolContainsUST(ctx, gamm, whitelistedPool); err != nil {
			panic(err)
		}
	}

	superfluid.SetUnpoolAllowedPools(ctx, whitelistedPoolShares)
}

// CheckPoolContainsUST looks up the pool from the gammkeeper and
// returns nil if the pool contains UST's ibc denom
// returns an error if the pool does not contain UST's ibc denom or on any other error case.
func CheckPoolContainsUST(ctx sdk.Context, gamm *gammkeeper.Keeper, poolID uint64) error {
	pool, err := gamm.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	assets, err := pool.GetPoolAssets()
	if err != nil {
		return err
	}

	for _, asset := range assets {
		if asset.Token.Denom == "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC" {
			return nil
		}
	}

	return fmt.Errorf("pool with ID %d does not contain UST", poolID)
}
