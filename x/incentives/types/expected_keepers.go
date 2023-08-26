package types

import (
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin

	HasSupply(ctx sdk.Context, denom string) bool

	SendCoinsFromModuleToManyAccounts(
		ctx sdk.Context, senderModule string, recipientAddrs []sdk.AccAddress, amts []sdk.Coins,
	) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// LockupKeeper defines the expected interface needed to retrieve locks.
type LockupKeeper interface {
	GetLocksLongerThanDurationDenom(ctx sdk.Context, denom string, duration time.Duration) []lockuptypes.PeriodLock
	GetPeriodLocksAccumulation(ctx sdk.Context, query lockuptypes.QueryCondition) sdk.Int
	GetAccountPeriodLocks(ctx sdk.Context, addr sdk.AccAddress) []lockuptypes.PeriodLock
	GetLockByID(ctx sdk.Context, lockID uint64) (*lockuptypes.PeriodLock, error)
}

// EpochKeeper defines the expected interface needed to retrieve epoch info.
type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
}

// CommunityPoolKeeper defines the contract needed to be fulfilled for distribution keeper.
type CommunityPoolKeeper interface {
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// TxFeesKeeper defines the expected interface needed to managing transaction fees.
type TxFeesKeeper interface {
	GetBaseDenom(ctx sdk.Context) (denom string, err error)
}

type ConcentratedLiquidityKeeper interface {
	CreateIncentive(ctx sdk.Context, poolId uint64, sender sdk.AccAddress, incentiveCoin sdk.Coin, emissionRate sdk.Dec, startTime time.Time, minUptime time.Duration) (cltypes.IncentiveRecord, error)
	GetConcentratedPoolById(ctx sdk.Context, poolId uint64) (cltypes.ConcentratedPoolExtension, error)
}

type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

type PoolIncentiveKeeper interface {
	GetPoolIdFromGaugeId(ctx sdk.Context, gaugeId uint64, lockableDuration time.Duration) (uint64, error)
	SetPoolGaugeIdNoLock(ctx sdk.Context, poolId uint64, gaugeId uint64)
}

type GAMMKeeper interface {
	GetPoolType(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolType, error)
}

type PoolManagerKeeper interface {
	GetPool(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolI, error)
}
