package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/osmoutils"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v15/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v15/x/pool-incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	appparams "github.com/osmosis-labs/osmosis/v15/app/params"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

type Keeper struct {
	storeKey sdk.StoreKey

	paramSpace paramtypes.Subspace

	epochKeeper       types.EpochKeeper
	incentivesKeeper  types.IncentivesKeeper
	accountKeeper     types.AccountKeeper
	bankKeeper        types.BankKeeper
	distrKeeper       types.DistrKeeper
	poolmanagerKeeper types.PoolManagerKeeper
}

func NewKeeper(storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, incentivesKeeper types.IncentivesKeeper, distrKeeper types.DistrKeeper, poolmanagerKeeper types.PoolManagerKeeper, epochKeeper types.EpochKeeper) Keeper {
	// ensure pool-incentives module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey: storeKey,

		paramSpace: paramSpace,

		accountKeeper:     accountKeeper,
		bankKeeper:        bankKeeper,
		incentivesKeeper:  incentivesKeeper,
		distrKeeper:       distrKeeper,
		poolmanagerKeeper: poolmanagerKeeper,
		epochKeeper:       epochKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// CreateLockablePoolGauges create multiple gauges based on lockableDurations.
// TODO: test this function
func (k Keeper) CreateLockablePoolGauges(ctx sdk.Context, poolId uint64) error {
	// Create the same number of gauges as there are LockableDurations
	for _, lockableDuration := range k.GetLockableDurations(ctx) {
		gaugeId, err := k.incentivesKeeper.CreateGauge(
			ctx,
			true,
			k.accountKeeper.GetModuleAddress(types.ModuleName),
			sdk.Coins{},
			lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         gammtypes.GetPoolShareDenom(poolId),
				Duration:      lockableDuration,
				Timestamp:     time.Time{},
			},
			ctx.BlockTime(),
			1,
		)
		if err != nil {
			return err
		}

		k.SetPoolGaugeId(ctx, poolId, lockableDuration, gaugeId)
	}
	return nil
}

// CreateConcentratedLiquidityPoolGauges creates a gauge for concentrated liquidity pool.
// TODO: test this function
func (k Keeper) CreateConcentratedLiquidityPoolGauges(ctx sdk.Context, poolId uint64) error {
	pool, err := k.poolmanagerKeeper.GetPool(ctx, poolId)
	if err != nil {
		return err
	}
	isCLPool := pool.GetType() == poolmanagertypes.Concentrated
	if !isCLPool {
		return fmt.Errorf("pool %d is not concentrated liquidity pool", poolId)
	}

	gaugeIdCL, err := k.incentivesKeeper.CreateGauge(
		ctx,
		true,
		k.accountKeeper.GetModuleAddress(types.ModuleName),
		sdk.Coins{},
		// dummy variable so that the existing logic does not break
		// CreateGauge checks if LockQueryType is `ByDuration` or not, we bypass this check by passing
		// lockQueryType as byTime. Although we donot need this check, we still cannot pass empty struct.
		lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByTime,
			Denom:         appparams.BaseCoinUnit,
		},
		ctx.BlockTime(),
		1,
	)
	if err != nil {
		return err
	}

	incParams := k.incentivesKeeper.GetEpochInfo(ctx)
	// lockable duration is 1day because we create incentive_record on every epoch
	k.SetPoolGaugeId(ctx, poolId, incParams.Duration, gaugeIdCL)

	return nil
}

func (k Keeper) SetPoolGaugeId(ctx sdk.Context, poolId uint64, lockableDuration time.Duration, gaugeId uint64) {
	key := types.GetPoolGaugeIdStoreKey(poolId, lockableDuration)
	store := ctx.KVStore(k.storeKey)
	store.Set(key, sdk.Uint64ToBigEndian(gaugeId))

	key = types.GetPoolIdFromGaugeIdStoreKey(gaugeId, lockableDuration)
	store.Set(key, sdk.Uint64ToBigEndian(poolId))
}

func (k Keeper) GetPoolGaugeId(ctx sdk.Context, poolId uint64, lockableDuration time.Duration) (uint64, error) {
	key := types.GetPoolGaugeIdStoreKey(poolId, lockableDuration)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)

	if len(bz) == 0 {
		return 0, sdkerrors.Wrapf(types.ErrNoGaugeIdExist, "gauge id for pool (%d) with duration (%s) not exist", poolId, lockableDuration.String())
	}

	return sdk.BigEndianToUint64(bz), nil
}

func (k Keeper) GetPoolIdFromGaugeId(ctx sdk.Context, gaugeId uint64, lockableDuration time.Duration) (uint64, error) {
	key := types.GetPoolIdFromGaugeIdStoreKey(gaugeId, lockableDuration)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)

	if len(bz) == 0 {
		return 0, types.NoPoolAssociatedWithGaugeError{GaugeId: gaugeId, Duration: lockableDuration}
	}

	return sdk.BigEndianToUint64(bz), nil
}

func (k Keeper) SetLockableDurations(ctx sdk.Context, lockableDurations []time.Duration) {
	store := ctx.KVStore(k.storeKey)
	info := types.LockableDurationsInfo{LockableDurations: lockableDurations}
	osmoutils.MustSet(store, types.LockableDurationsKey, &info)
}

func (k Keeper) GetLockableDurations(ctx sdk.Context) []time.Duration {
	store := ctx.KVStore(k.storeKey)
	info := types.LockableDurationsInfo{}
	osmoutils.MustGet(store, types.LockableDurationsKey, &info)
	return info.LockableDurations
}

func (k Keeper) GetAllGauges(ctx sdk.Context) []incentivestypes.Gauge {
	gauges := k.incentivesKeeper.GetGauges(ctx)
	return gauges
}

func (k Keeper) IsPoolIncentivized(ctx sdk.Context, poolId uint64) bool {
	lockableDurations := k.GetLockableDurations(ctx)
	distrInfo := k.GetDistrInfo(ctx)

	candidateGaugeIds := []uint64{}
	for _, lockableDuration := range lockableDurations {
		gaugeId, err := k.GetPoolGaugeId(ctx, poolId, lockableDuration)
		if err == nil {
			candidateGaugeIds = append(candidateGaugeIds, gaugeId)
		}
	}

	for _, record := range distrInfo.Records {
		for _, gaugeId := range candidateGaugeIds {
			if record.GaugeId == gaugeId {
				return true
			}
		}
	}
	return false
}
