package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/x/concentrated-liquidity/client/queryproto"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

// Chain keepers required for sqs ingest.
type SQSIngestKeepers struct {
	GammKeeper         PoolKeeper
	CosmWasmPoolKeeper CosmWasmPoolKeeper
	BankKeeper         BankKeeper
	ProtorevKeeper     ProtorevKeeper
	PoolManagerKeeper  PoolManagerKeeper
	ConcentratedKeeper ConcentratedKeeper
}

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

	RouteGetPoolDenoms(
		ctx sdk.Context,
		poolId uint64,
	) (denoms []string, err error)

	GetTradingPairTakerFee(ctx sdk.Context, denom0, denom1 string) (osmomath.Dec, error)
}

// ConcentratedKeeper is an interface for the concentrated keeper.
type ConcentratedKeeper interface {
	PoolKeeper
	GetTickLiquidityForFullRange(ctx sdk.Context, poolId uint64) ([]queryproto.LiquidityDepthWithRange, int64, error)
}
