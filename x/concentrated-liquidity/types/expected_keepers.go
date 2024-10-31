package types

import (
	context "context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

type AccountKeeper interface {
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// BankKeeper defines the banking contract that must be fulfilled when
// creating a x/concentrated-liquidity keeper.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	HasBalance(ctx context.Context, addr sdk.AccAddress, amt sdk.Coin) bool
	MintCoins(ctx context.Context, name string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error
}

// PoolManagerKeeper defines the interface needed to be fulfilled for
// the poolmanager keeper.
type PoolManagerKeeper interface {
	GetParams(ctx sdk.Context) (params poolmanagertypes.Params)
	SetParams(ctx sdk.Context, params poolmanagertypes.Params)
	CreatePool(ctx sdk.Context, msg poolmanagertypes.CreatePoolMsg) (uint64, error)
	GetNextPoolId(ctx sdk.Context) uint64
	CreateConcentratedPoolAsPoolManager(ctx sdk.Context, msg poolmanagertypes.CreatePoolMsg) (poolmanagertypes.PoolI, error)
}

type GAMMKeeper interface {
	GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error)
	GetLinkedBalancerPoolID(ctx sdk.Context, poolIdEntering uint64) (uint64, error)
	GetTotalPoolShares(ctx sdk.Context, poolId uint64) (osmomath.Int, error)
}

type PoolIncentivesKeeper interface {
	GetPoolGaugeId(ctx sdk.Context, poolId uint64, lockableDuration time.Duration) (uint64, error)
	GetLongestLockableDuration(ctx sdk.Context) (time.Duration, error)
}

type IncentivesKeeper interface {
	AddToGaugeRewards(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, gaugeID uint64) error
}

// LockupKeeper defines the expected interface needed to retrieve locks.
type LockupKeeper interface {
	GetLockByID(ctx sdk.Context, lockID uint64) (*lockuptypes.PeriodLock, error)
	// Despite the name, BeginForceUnlock is really BeginUnlock
	// TODO: Fix this in future code update
	BeginForceUnlock(ctx sdk.Context, lockID uint64, coins sdk.Coins) (uint64, error)
	ForceUnlock(ctx sdk.Context, lock lockuptypes.PeriodLock) error
	CreateLock(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (lockuptypes.PeriodLock, error)
	CreateLockNoSend(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (lockuptypes.PeriodLock, error)
	SlashTokensFromLockByID(ctx sdk.Context, lockID uint64, coins sdk.Coins) (*lockuptypes.PeriodLock, error)
	GetLockedDenom(ctx sdk.Context, denom string, duration time.Duration) osmomath.Int
}

// CommunityPoolKeeper defines the contract needed to be fulfilled for distribution keeper.
type CommunityPoolKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// ContractKeeper handles logic related to CosmWasm contract interactions.
type ContractKeeper interface {
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}
