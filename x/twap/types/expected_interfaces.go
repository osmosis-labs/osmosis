package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// AmmInterface is the functionality needed from a given pool ID, in order to maintain records and serve TWAPs.
type PoolManagerInterface interface {
	RouteGetPoolDenoms(ctx sdk.Context, poolId uint64) (denoms []string, err error)
	// CalculateSpotPrice returns the spot price of the quote asset in terms of the base asset,
	// using the specified pool.
	// E.g. if pool 1 traded 2 atom for 3 osmo, the quote asset was atom, and the base asset was osmo,
	// this would return 1.5. (Meaning that 1 atom costs 1.5 osmo)
	RouteCalculateSpotPrice(
		ctx sdk.Context,
		poolID uint64,
		quoteAssetDenom string,
		baseAssetDenom string,
	) (price osmomath.Dec, err error)
}
