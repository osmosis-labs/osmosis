package types

import (
	context "context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/gamm keeper.
type AccountKeeper interface {
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI

	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)

	GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string)
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// BankKeeper defines the banking contract that must be fulfilled when
// creating a x/gamm keeper.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error

	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error

	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)

	// Only needed for simulation interface matching
	// TODO: Look into golang syntax to make this "Everything in stakingtypes.bankkeeper + extra funcs"
	// I think it has to do with listing another interface as the first line here?
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// CommunityPoolKeeper defines the contract needed to be fulfilled for distribution keeper.
type CommunityPoolKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// ConcentratedLiquidityKeeper defines the contract needed to be fulfilled for the concentrated liquidity keeper.
type ConcentratedLiquidityKeeper interface {
	SetParams(ctx sdk.Context, params cltypes.Params)
	GetConcentratedPoolById(ctx sdk.Context, poolId uint64) (cltypes.ConcentratedPoolExtension, error)
	CreateFullRangePosition(ctx sdk.Context, clPoolId uint64, owner sdk.AccAddress, coins sdk.Coins) (positionData cltypes.CreateFullRangePositionData, err error)
}

// PoolManager defines the interface needed to be fulfilled for
// the pool manager.
type PoolManager interface {
	CreatePool(ctx sdk.Context, msg poolmanagertypes.CreatePoolMsg) (uint64, error)

	GetNextPoolId(ctx sdk.Context) uint64

	RouteExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		routes []poolmanagertypes.SwapAmountInRoute,
		tokenIn sdk.Coin,
		tokenOutMinAmount osmomath.Int) (tokenOutAmount osmomath.Int, err error)

	RouteExactAmountOut(ctx sdk.Context,
		sender sdk.AccAddress,
		routes []poolmanagertypes.SwapAmountOutRoute,
		tokenInMaxAmount osmomath.Int,
		tokenOut sdk.Coin,
	) (tokenInAmount osmomath.Int, err error)

	MultihopEstimateOutGivenExactAmountIn(
		ctx sdk.Context,
		routes []poolmanagertypes.SwapAmountInRoute,
		tokenIn sdk.Coin,
	) (tokenOutAmount osmomath.Int, err error)

	MultihopEstimateInGivenExactAmountOut(
		ctx sdk.Context,
		routes []poolmanagertypes.SwapAmountOutRoute,
		tokenOut sdk.Coin) (tokenInAmount osmomath.Int, err error)

	GetPoolModule(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolModuleI, error)

	GetPool(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolI, error)

	CreateConcentratedPoolAsPoolManager(ctx sdk.Context, msg poolmanagertypes.CreatePoolMsg) (poolmanagertypes.PoolI, error)

	GetTradingPairTakerFee(ctx sdk.Context, denom0, denom1 string) (osmomath.Dec, error)
}

type PoolIncentivesKeeper interface {
	GetGaugesForCFMMPool(ctx sdk.Context, poolId uint64) ([]incentivestypes.Gauge, error)
	GetPoolGaugeId(ctx sdk.Context, poolId uint64, lockableDuration time.Duration) (uint64, error)
	GetDistrInfo(ctx sdk.Context) types.DistrInfo
	SetDistrInfo(ctx sdk.Context, distrInfo types.DistrInfo)
}

type IncentivesKeeper interface {
	GetEpochInfo(ctx sdk.Context) epochtypes.EpochInfo
}
