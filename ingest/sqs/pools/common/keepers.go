package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

// PoolKeeper is an interface for getting pools from a keeper.
type PoolKeeper interface {
	GetPools(ctx sdk.Context) ([]poolmanagertypes.PoolI, error)
}

// CosmWasmPoolKeeper is an interface for getting CosmWasm pools from a keeper.
type CosmWasmPoolKeeper interface {
	GetPoolsWithWasmKeeper(ctx sdk.Context) ([]poolmanagertypes.PoolI, error)
}

// BankKeeper is an interface for getting bank balances.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// ProtorevKeeper is an interface for getting the pool for a denom pair.
type ProtorevKeeper interface {
	GetPoolForDenomPair(ctx sdk.Context, baseDenom, denomToMatch string) (uint64, error)
}

// PoolManagerKeeper is an interface for the pool manager keeper.
type PoolManagerKeeper interface {
	RouteCalculateSpotPrice(
		ctx sdk.Context,
		poolId uint64,
		quoteAssetDenom string,
		baseAssetDenom string,
	) (price osmomath.BigDec, err error)

	SwapExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		poolId uint64,
		tokenIn sdk.Coin,
		tokenOutDenom string,
		tokenOutMinAmount osmomath.Int,
	) (tokenOutAmount osmomath.Int, err error)
}
