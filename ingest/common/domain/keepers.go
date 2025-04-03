package commondomain

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client/queryproto"
	concentratedtypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// Chain keepers required for extracting pool data.
type PoolExtractorKeepers struct {
	GammKeeper         PoolKeeper
	CosmWasmPoolKeeper CosmWasmPoolKeeper
	WasmKeeper         WasmKeeper
	BankKeeper         BankKeeper
	ProtorevKeeper     ProtorevKeeper
	PoolManagerKeeper  PoolManagerKeeper
	ConcentratedKeeper ConcentratedKeeper
}

type WriteListener interface {
	OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error
}

// PoolKeeper is an interface for getting pools from a keeper.
type PoolKeeper interface {
	GetPools(ctx sdk.Context) ([]poolmanagertypes.PoolI, error)
}

// CosmWasmPoolKeeper is an interface for getting CosmWasm pools from a keeper.
type CosmWasmPoolKeeper interface {
	GetPoolsWithWasmKeeper(ctx sdk.Context) ([]poolmanagertypes.PoolI, error)
}

// WasmKeeper is an interface for querying CosmWasm contract.
type WasmKeeper interface {
	QueryRaw(ctx context.Context, contractAddress sdk.AccAddress, key []byte) []byte
	QuerySmart(ctx context.Context, contractAddress sdk.AccAddress, req []byte) ([]byte, error)
}

// BankKeeper is an interface for getting bank balances.
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
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
	) (tokenOutAmount osmomath.Int, takerFeeTotal sdk.Coin, err error)

	RouteGetPoolDenoms(
		ctx sdk.Context,
		poolId uint64,
	) (denoms []string, err error)

	GetTradingPairTakerFee(ctx sdk.Context, denom0, denom1 string) (osmomath.Dec, error)

	MultihopEstimateInGivenExactAmountOut(
		ctx sdk.Context,
		route []poolmanagertypes.SwapAmountOutRoute,
		tokenOut sdk.Coin,
	) (tokenInAmount osmomath.Int, err error)
}

// ConcentratedKeeper is an interface for the concentrated keeper.
type ConcentratedKeeper interface {
	PoolKeeper
	GetTickLiquidityForFullRange(ctx sdk.Context, poolId uint64) ([]queryproto.LiquidityDepthWithRange, int64, error)
	GetConcentratedPoolById(ctx sdk.Context, poolId uint64) (concentratedtypes.ConcentratedPoolExtension, error)
}
