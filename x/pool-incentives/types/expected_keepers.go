package types

import (
	context "context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	gammmigration "github.com/osmosis-labs/osmosis/v27/x/gamm/types/migration"
	incentivestypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// AccountKeeper interface contains functions for getting accounts and the module address
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
}

// BankKeeper sends tokens across modules and is able to get account balances.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// PoolManagerKeeper gets the pool interface from poolID.
type PoolManagerKeeper interface {
	GetNextPoolId(ctx sdk.Context) uint64
	GetPool(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolI, error)
}

type GAMMKeeper interface {
	GetAllMigrationInfo(ctx sdk.Context) (gammmigration.MigrationRecords, error)
}

// IncentivesKeeper creates and gets gauges, and also allows additions to gauge rewards.
type IncentivesKeeper interface {
	CreateGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64, poolId uint64) (uint64, error)
	GetGaugeByID(ctx sdk.Context, gaugeID uint64) (*incentivestypes.Gauge, error)
	GetGauges(ctx sdk.Context) []incentivestypes.Gauge
	GetParams(ctx sdk.Context) incentivestypes.Params
	GetEpochInfo(ctx sdk.Context) epochstypes.EpochInfo

	AddToGaugeRewards(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, gaugeID uint64) error
	GetGroupByGaugeID(ctx sdk.Context, gaugeID uint64) (incentivestypes.Group, error)
	GetPoolIdsAndDurationsFromGaugeRecords(ctx sdk.Context, gaugeRecords []incentivestypes.InternalGaugeRecord) ([]uint64, []time.Duration, error)
}

// DistrKeeper handles pool-fees functionality - setting / getting fees and funding the community pool.
type DistrKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
}

type SuperfluidKeeper interface {
	GetAllMigrationInfo(ctx sdk.Context) (MigrationRecords, error)
}
