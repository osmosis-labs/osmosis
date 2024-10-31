package keeper

import (
	"fmt"
	"strconv"
	"time"

	"cosmossdk.io/log"
	"github.com/hashicorp/go-metrics"

	"github.com/osmosis-labs/osmosis/osmoutils"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	storetypes "cosmossdk.io/store/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey

	paramSpace paramtypes.Subspace

	epochKeeper       types.EpochKeeper
	incentivesKeeper  types.IncentivesKeeper
	accountKeeper     types.AccountKeeper
	bankKeeper        types.BankKeeper
	distrKeeper       types.DistrKeeper
	poolmanagerKeeper types.PoolManagerKeeper
	gammKeeper        types.GAMMKeeper
}

func NewKeeper(storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, incentivesKeeper types.IncentivesKeeper, distrKeeper types.DistrKeeper, poolmanagerKeeper types.PoolManagerKeeper, epochKeeper types.EpochKeeper, gammKeeper types.GAMMKeeper) Keeper {
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
		gammKeeper:        gammKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// CreateLockablePoolGauges create multiple gauges based on lockableDurations.
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
			0,
		)
		if err != nil {
			return err
		}

		if err := k.SetPoolGaugeIdInternalIncentive(ctx, poolId, lockableDuration, gaugeId); err != nil {
			return err
		}
	}
	return nil
}

// CreateConcentratedLiquidityPoolGauge creates a gauge for concentrated liquidity pool.
func (k Keeper) CreateConcentratedLiquidityPoolGauge(ctx sdk.Context, poolId uint64) error {
	pool, err := k.poolmanagerKeeper.GetPool(ctx, poolId)
	if err != nil {
		return err
	}
	isCLPool := pool.GetType() == poolmanagertypes.Concentrated
	if !isCLPool {
		return fmt.Errorf("pool %d is not concentrated liquidity pool", poolId)
	}

	incentivesEpoch := k.incentivesKeeper.GetEpochInfo(ctx)
	incentivesEpochDuration := incentivesEpoch.Duration

	gaugeId, err := k.incentivesKeeper.CreateGauge(
		ctx,
		true,
		k.accountKeeper.GetModuleAddress(types.ModuleName),
		sdk.Coins{},
		lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.NoLock,
			Denom:         incentivestypes.NoLockInternalGaugeDenom(pool.GetId()),
			// We specify this duration so that we can query this duration in the IncentivizedPools() query.
			Duration: incentivesEpochDuration,
		},
		ctx.BlockTime(),
		1,
		poolId,
	)
	if err != nil {
		return err
	}

	// Although the pool id <> gauge "NoLock" link is created in CreateGauge,
	// we create an additional "ByDuration" link here for tracking
	// internal incentive "NoLock" gauges
	if err := k.SetPoolGaugeIdInternalIncentive(ctx, poolId, incentivesEpochDuration, gaugeId); err != nil {
		return err
	}

	return nil
}

// SetPoolGaugeIdInternalIncentive sets the gauge id for the pool and internally incentivized duration.
// Returns error if the incentivized duration is zero.
// CONTRACT: this link is created only for the internally incentivized gauges.
func (k Keeper) SetPoolGaugeIdInternalIncentive(ctx sdk.Context, poolId uint64, incentivizedDuration time.Duration, gaugeId uint64) error {
	if incentivizedDuration == 0 {
		return fmt.Errorf("incentivized duration cannot be zero, pool id: %d", poolId)
	}

	// Note: this index is used for internal incentive gauges only.
	key := types.GetPoolGaugeIdInternalStoreKey(poolId, incentivizedDuration)
	store := ctx.KVStore(k.storeKey)
	store.Set(key, sdk.Uint64ToBigEndian(gaugeId))

	// Note: this index is used for general linking.
	key = types.GetPoolIdFromGaugeIdStoreKey(gaugeId, incentivizedDuration)
	store.Set(key, sdk.Uint64ToBigEndian(poolId))

	return nil
}

// SetPoolGaugeIdNoLock sets the link between pool id and gauge id for "NoLock" gauges.
// CONTRACT: the gauge of the given id must be "NoLock" gauge.
func (k Keeper) SetPoolGaugeIdNoLock(ctx sdk.Context, poolId uint64, gaugeId uint64, incentivizedUptime time.Duration) {
	store := ctx.KVStore(k.storeKey)
	// maps pool id and gauge id to gauge id.
	// Note: this could be pool id and gauge id to empty byte array,
	// but is chosen this way for ease of implementation at the cost of space.
	// Note 2: this index is used for "NoLock" gauges only.
	key := types.GetPoolNoLockGaugeIdStoreKey(poolId, gaugeId)
	store.Set(key, sdk.Uint64ToBigEndian(gaugeId))

	// Note: this index is used for general linking.
	key = types.GetPoolIdFromGaugeIdStoreKey(gaugeId, incentivizedUptime)
	store.Set(key, sdk.Uint64ToBigEndian(poolId))
}

// GetInternalGaugeIdForPool returns the internally incentivized gauge ID associated with the pool ID.
// Contrary to GetPoolGaugeId, it determines the appropriate lockable duration based on the pool type.
// For balancer or stableswap pools that have 3 gauges, it assumes the longest lockable duration gauge.
// For CL pools, it assumes the gauge with the incentive module epoch duration.
// Returns gauge ID on success, returns error if:
// - fails to get pool
// - given pool type does not support incentives (e.g. CW pools at the time of writing)
// - fails to get the gauge ID for the given poolID and inferred lockable duration
func (k Keeper) GetInternalGaugeIDForPool(ctx sdk.Context, poolID uint64) (uint64, error) {
	pool, err := k.poolmanagerKeeper.GetPool(ctx, poolID)
	if err != nil {
		return 0, err
	}

	poolType := pool.GetType()
	var gaugeDuration time.Duration
	switch poolType {
	case poolmanagertypes.Concentrated:
		gaugeDuration = k.incentivesKeeper.GetEpochInfo(ctx).Duration
	case poolmanagertypes.Balancer, poolmanagertypes.Stableswap:
		gaugeDuration, err = k.GetLongestLockableDuration(ctx)
		if err != nil {
			return 0, err
		}
	default:
		return 0, types.UnsupportedPoolTypeError{PoolID: poolID, PoolType: poolType}
	}

	gaugeId, err := k.GetPoolGaugeId(ctx, poolID, gaugeDuration)
	if err != nil {
		return 0, err
	}

	return gaugeId, nil
}

// GetPoolGaugeId returns the gauge id associated with the pool id and lockable duration.
// This can only be used for the internally incentivized gauges.
// Externally incentivized gauges do not have such link created.
func (k Keeper) GetPoolGaugeId(ctx sdk.Context, poolId uint64, lockableDuration time.Duration) (uint64, error) {
	if lockableDuration == 0 {
		return 0, fmt.Errorf("cannot get gauge id from pool id without a lockable duration. There can be many gauges for pool id %d and duration 0", poolId)
	}

	key := types.GetPoolGaugeIdInternalStoreKey(poolId, lockableDuration)
	store := ctx.KVStore(k.storeKey)

	if !store.Has(key) {
		return 0, types.NoGaugeAssociatedWithPoolError{PoolId: poolId, Duration: lockableDuration}
	}

	bz := store.Get(key)
	gaugeId := sdk.BigEndianToUint64(bz)
	return gaugeId, nil
}

// GetNoLockGaugeIdsFromPool returns all the NoLock gauge ids associated with the pool id.
// This can only be used for the "NoLock" gauges. For "NoLock" gauges there are 2 kinds
// of links created:
// - general
// - by duration (for internal incentives)
//
// Every "NoLock" gauge has the first link. Only the internal incentives "NoLock" gauges
// have the second link.
func (k Keeper) GetNoLockGaugeIdsFromPool(ctx sdk.Context, poolId uint64) ([]uint64, error) {
	store := ctx.KVStore(k.storeKey)
	gaugeIds, err := osmoutils.GatherValuesFromStorePrefix(store, types.GetPoolNoLockGaugeIdIterationStoreKey(poolId), func(b []byte) (uint64, error) {
		return sdk.BigEndianToUint64(b), nil
	})
	if err != nil {
		return nil, err
	}
	return gaugeIds, nil
}

func (k Keeper) GetPoolIdFromGaugeId(ctx sdk.Context, gaugeId uint64, lockableDuration time.Duration) (uint64, error) {
	key := types.GetPoolIdFromGaugeIdStoreKey(gaugeId, lockableDuration)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)

	if len(bz) == 0 {
		telemetry.IncrCounterWithLabels([]string{types.NoPoolIdForGaugeTelemetryName}, 1, []metrics.Label{
			{
				Name:  "gauge_id",
				Value: strconv.FormatUint(gaugeId, 10),
			},
			{
				Name:  "duration",
				Value: lockableDuration.String(),
			},
		})
		return 0, types.NoPoolAssociatedWithGaugeError{GaugeId: gaugeId, Duration: lockableDuration}
	}

	return sdk.BigEndianToUint64(bz), nil
}

// GetGaugesForCFMMPool returns the gauges associated with the given CFMM pool ID, by first retrieving
// the lockable durations for the pool, then using them to query the pool incentives keeper for the
// gauge IDs associated with each duration, and finally using the incentives keeper to retrieve the
// actual gauges from the retrieved gauge IDs.
// CONTRACT: pool id must be associated with a CFMM pool.
func (k Keeper) GetGaugesForCFMMPool(ctx sdk.Context, poolId uint64) ([]incentivestypes.Gauge, error) {
	lockableDurations := k.GetLockableDurations(ctx)
	cfmmGauges := make([]incentivestypes.Gauge, 0, len(lockableDurations))
	for _, duration := range lockableDurations {
		gaugeId, err := k.GetPoolGaugeId(ctx, poolId, duration)
		if err != nil {
			return nil, err
		}
		gauge, err := k.incentivesKeeper.GetGaugeByID(ctx, gaugeId)
		if err != nil {
			return nil, err
		}

		cfmmGauges = append(cfmmGauges, *gauge)
	}

	return cfmmGauges, nil
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

func (k Keeper) GetLongestLockableDuration(ctx sdk.Context) (time.Duration, error) {
	lockableDurations := k.GetLockableDurations(ctx)
	if len(lockableDurations) == 0 {
		return 0, fmt.Errorf("Lockable Durations does not exist")
	}
	longestDuration := time.Duration(0)

	for _, duration := range lockableDurations {
		if duration > longestDuration {
			longestDuration = duration
		}
	}

	return longestDuration, nil
}

func (k Keeper) GetAllGauges(ctx sdk.Context) []incentivestypes.Gauge {
	gauges := k.incentivesKeeper.GetGauges(ctx)
	return gauges
}

// IsPoolIncentivized checks if a given pool is internally incentivized or not.
// It first retrieves the epoch duration and lockable durations from the keeper.
// Then it iterates over the distribution records, retrieving the gauge for each record.
// If the gauge's DistributeTo.LockQueryType is ByGroup, it retrieves the group and checks if it's perpetual.
// If it is, it retrieves the pool IDs from the group and checks if the provided pool ID is in the list.
// If the gauge's DistributeTo.LockQueryType is NoLock, it retrieves the pool ID from the gauge and checks if it matches the provided pool ID.
// If the gauge's DistributeTo.LockQueryType is ByDuration, it iterates over the lockableDurations, retrieves the pool ID for each duration, and checks if it matches the provided pool ID.
// If none of the checks pass, it returns false, indicating that the pool is not incentivized.
func (k Keeper) IsPoolIncentivized(ctx sdk.Context, providedPoolId uint64) (bool, error) {
	epochDuration := k.incentivesKeeper.GetEpochInfo(ctx).Duration
	lockableDurations := k.GetLockableDurations(ctx)

	distrInfo := k.GetDistrInfo(ctx)

	for _, record := range distrInfo.Records {
		// Skip community pool gauge
		if record.GaugeId == 0 {
			continue
		}
		gauge, err := k.incentivesKeeper.GetGaugeByID(ctx, record.GaugeId)
		if err != nil {
			return false, err
		}
		if gauge.DistributeTo.LockQueryType == lockuptypes.ByGroup {
			group, err := k.incentivesKeeper.GetGroupByGaugeID(ctx, record.GaugeId)
			if err != nil {
				return false, err
			}
			groupGauge, err := k.incentivesKeeper.GetGaugeByID(ctx, group.GroupGaugeId)
			if err != nil {
				return false, err
			}
			if !groupGauge.IsPerpetual {
				// if the group is not perpetual, it is an externally incentivized gauge so we skip it
				continue
			}
			poolIds, _, err := k.incentivesKeeper.GetPoolIdsAndDurationsFromGaugeRecords(ctx, group.InternalGaugeInfo.GaugeRecords)
			if err != nil {
				return false, err
			}
			for _, poolId := range poolIds {
				if poolId == providedPoolId {
					return true, nil
				}
			}
		} else if gauge.DistributeTo.LockQueryType == lockuptypes.NoLock {
			poolId, err := k.GetPoolIdFromGaugeId(ctx, record.GaugeId, epochDuration)
			if err != nil {
				return false, err
			}
			if poolId == providedPoolId {
				return true, nil
			}
		} else if gauge.DistributeTo.LockQueryType == lockuptypes.ByDuration {
			gauge, err := k.incentivesKeeper.GetGaugeByID(ctx, record.GaugeId)
			if err != nil {
				return false, err
			}
			// Ensure the gauge's duration matches one of the lockable durations.
			matchFound := false
			for _, duration := range lockableDurations {
				if gauge.DistributeTo.Duration == duration {
					matchFound = true
					break
				}
			}
			if !matchFound {
				return false, types.IncentiveRecordContainsNonLockableDurationError{GaugeId: gauge.Id, Duration: gauge.DistributeTo.Duration, LockableDurations: lockableDurations}
			}
			poolId, err := k.GetPoolIdFromGaugeId(ctx, record.GaugeId, gauge.DistributeTo.Duration)
			if err != nil {
				return false, err
			}
			if poolId == providedPoolId {
				return true, nil
			}
		} else {
			return false, fmt.Errorf("unknown lock query type: %s", gauge.DistributeTo.LockQueryType)
		}
	}
	return false, nil
}
