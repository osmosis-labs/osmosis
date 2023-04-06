package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

// PoolI defines an interface for pools that hold tokens.
type PoolI interface {
	proto.Message

	GetAddress() sdk.AccAddress
	String() string
	GetId() uint64
	// GetSwapFee returns the pool's swap fee, based on the current state.
	// Pools may choose to make their swap fees dependent upon state
	// (prior TWAPs, network downtime, other pool states, etc.)
	// hence Context is provided as an argument.
	GetSwapFee(ctx sdk.Context) sdk.Dec
	// Returns whether the pool has swaps enabled at the moment
	IsActive(ctx sdk.Context) bool

	// Returns the spot price of the 'base asset' in terms of the 'quote asset' in the pool,
	// errors if either baseAssetDenom, or quoteAssetDenom does not exist.
	// For example, if this was a UniV2 50-50 pool, with 2 ETH, and 8000 UST
	// pool.SpotPrice(ctx, "eth", "ust") = 4000.00
	SpotPrice(ctx sdk.Context, quoteAssetDenom string, baseAssetDenom string) (sdk.Dec, error)
	// GetType returns the type of the pool (Balancer, Stableswap, Concentrated, etc.)
	GetType() PoolType
}

// NewPoolAddress returns an address for a pool from a given id.
func NewPoolAddress(poolId uint64) sdk.AccAddress {
	return osmoutils.NewModuleAddressWithPrefix(ModuleName, "pool", sdk.Uint64ToBigEndian(poolId))
}
