package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/protorev keeper.
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// BankKeeper defines the banking contract that must be fulfilled when
// creating a x/protorev keeper.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// GAMMKeeper defines the Gamm contract that must be fulfilled when
// creating a x/protorev keeper.
type GAMMKeeper interface {
	GetPoolAndPoke(ctx sdk.Context, poolId uint64) (gammtypes.CFMMPoolI, error)
}

// PoolManagerKeeper defines the PoolManager contract that must be fulfilled when
// creating a x/protorev keeper.
type PoolManagerKeeper interface {
	RouteExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		routes []poolmanagertypes.SwapAmountInRoute,
		tokenIn sdk.Coin,
		tokenOutMinAmount osmomath.Int) (tokenOutAmount osmomath.Int, err error)

	MultihopEstimateOutGivenExactAmountInNoTakerFee(
		ctx sdk.Context,
		routes []poolmanagertypes.SwapAmountInRoute,
		tokenIn sdk.Coin,
	) (tokenOutAmount osmomath.Int, err error)

	MultihopEstimateInGivenExactAmountOut(
		ctx sdk.Context,
		routes []poolmanagertypes.SwapAmountOutRoute,
		tokenOut sdk.Coin) (tokenInAmount osmomath.Int, err error)

	AllPools(
		ctx sdk.Context,
	) ([]poolmanagertypes.PoolI, error)
	GetPool(
		ctx sdk.Context,
		poolId uint64,
	) (poolmanagertypes.PoolI, error)
	GetPoolType(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolType, error)
	GetPoolModule(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolModuleI, error)
	GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error)
	RouteGetPoolDenoms(ctx sdk.Context, poolId uint64) ([]string, error)
	GetTakerFeeTrackerForStakers(ctx sdk.Context) []sdk.Coin
	GetTakerFeeTrackerForCommunityPool(ctx sdk.Context) []sdk.Coin
	GetTakerFeeTrackerStartHeight(ctx sdk.Context) int64
}

// EpochKeeper defines the Epoch contract that must be fulfilled when
// creating a x/protorev keeper.
type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochtypes.EpochInfo
}

// ConcentratedLiquidityKeeper defines the ConcentratedLiquidity contract that must be fulfilled when
// creating a x/protorev keeper.
type ConcentratedLiquidityKeeper interface {
	ComputeMaxInAmtGivenMaxTicksCrossed(
		ctx sdk.Context,
		poolId uint64,
		tokenInDenom string,
		maxTicksCrossed uint64,
	) (maxTokenIn, resultingTokenOut sdk.Coin, err error)
}

// DistributionKeeper defines the distribution contract that must be fulfilled when
// creating a x/protorev keeper.
type DistributionKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}
