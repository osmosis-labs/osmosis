package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// CalculateSpotPrice returns the spot price of the quote asset in terms of the base asset,
// using the specified pool.
// E.g. if pool 1 traded 2 atom for 3 osmo, the quote asset was atom, and the base asset was osmo,
// this would return 1.5. (Meaning that 1 atom costs 1.5 osmo)
type AmmInterface interface {
	GetPoolDenoms(ctx sdk.Context, poolId uint64) (denoms []string, err error)
	GetSpotPrice(ctx sdk.Context,
		poolID uint64,
		baseAssetDenom string,
		quoteAssetDenom string) (price sdk.Dec, err error)
}
