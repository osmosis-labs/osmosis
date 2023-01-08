package concentrated_liquidity

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

// GetLastFreezeID returns ID used last time.
func (k Keeper) GetLastFreezeID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.LastFreezeID)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetLastFreezeID sets the global last freeze ID.
func (k Keeper) SetLastFreezeID(ctx sdk.Context, ID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastFreezeID, sdk.Uint64ToBigEndian(ID))
}

// CreatePoolIncentive creates incentives for the specified pool with the given minimum freeze duration, and the
// `SecondsPerIncentivizedLiquidityGlobal` field is initialized to zero.
//
// If the pool with the given ID does not exist, a `PoolNotFoundError` is returned.
//
// Returns an error if there is any problem creating the pool incentive.
func (k Keeper) CreatePoolIncentive(ctx sdk.Context, poolId uint64, minimumFreezeDuration time.Duration) (err error) {
	// Check if pool exists
	if !k.poolExists(ctx, poolId) {
		return types.PoolNotFoundError{PoolId: poolId}
	}

	// Retrieve concentrated pool
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return err
	}

	// Get the current incentivized liquidity records for the pool
	poolIncentivizedLiquidityRecord := pool.GetPoolIncentivizedLiquidityRecords()

	// Append to that record the new pool incentive
	poolIncentivizedLiquidityRecord = append(poolIncentivizedLiquidityRecord, types.PoolIncentivizedLiquidityRecord{
		ID:                                    k.GetLastFreezeID(ctx) + 1,
		MinimumFreezeDuration:                 minimumFreezeDuration,
		SecondsPerIncentivizedLiquidityGlobal: sdk.ZeroDec(),
	})

	// Set the pool liquidity record then set the pool
	k.SetLastFreezeID(ctx, k.GetLastFreezeID(ctx)+1)
	pool.SetPoolIncentivizedLiquidityRecords(poolIncentivizedLiquidityRecord)
	err = k.setPool(ctx, pool)
	if err != nil {
		return err
	}
	return nil
}
