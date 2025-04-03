package types

import (
	context "context"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	epochstypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin

	HasSupply(ctx context.Context, denom string) bool

	SendCoinsFromModuleToManyAccounts(ctx context.Context, senderModule string, recipientAddrs []sdk.AccAddress, amts []sdk.Coins) error

	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// LockupKeeper defines the expected interface needed to retrieve locks.
type LockupKeeper interface {
	GetLocksLongerThanDurationDenom(ctx sdk.Context, denom string, duration time.Duration) []lockuptypes.PeriodLock
	GetPeriodLocksAccumulation(ctx sdk.Context, query lockuptypes.QueryCondition) osmomath.Int
	GetAccountPeriodLocks(ctx sdk.Context, addr sdk.AccAddress) []lockuptypes.PeriodLock
	GetLockByID(ctx sdk.Context, lockID uint64) (*lockuptypes.PeriodLock, error)
}

// EpochKeeper defines the expected interface needed to retrieve epoch info.
type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
}

// CommunityPoolKeeper defines the contract needed to be fulfilled for distribution keeper.
type CommunityPoolKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// TxFeesKeeper defines the expected interface needed to managing transaction fees.
type TxFeesKeeper interface {
	GetBaseDenom(ctx sdk.Context) (denom string, err error)
}

type ConcentratedLiquidityKeeper interface {
	CreateIncentive(ctx sdk.Context, poolId uint64, sender sdk.AccAddress, incentiveCoin sdk.Coin, emissionRate osmomath.Dec, startTime time.Time, minUptime time.Duration) (cltypes.IncentiveRecord, error)
	GetConcentratedPoolById(ctx sdk.Context, poolId uint64) (cltypes.ConcentratedPoolExtension, error)
	GetParams(ctx sdk.Context) (params cltypes.Params)
}

type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

type PoolIncentiveKeeper interface {
	GetPoolIdFromGaugeId(ctx sdk.Context, gaugeId uint64, lockableDuration time.Duration) (uint64, error)
	GetInternalGaugeIDForPool(ctx sdk.Context, poolID uint64) (uint64, error)
	SetPoolGaugeIdNoLock(ctx sdk.Context, poolId uint64, gaugeId uint64, uptime time.Duration)
	GetLongestLockableDuration(ctx sdk.Context) (time.Duration, error)
}

type GAMMKeeper interface {
	GetPoolType(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolType, error)
}

type PoolManagerKeeper interface {
	GetPool(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolI, error)
	GetOsmoVolumeForPool(ctx sdk.Context, poolId uint64) osmomath.Int
	GetPoolModuleAndPool(ctx sdk.Context, poolId uint64) (swapModule poolmanagertypes.PoolModuleI, pool poolmanagertypes.PoolI, err error)
}

type ProtorevKeeper interface {
	GetPoolForDenomPairNoOrder(ctx sdk.Context, denom1, denom2 string) (uint64, error)
}
