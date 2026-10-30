package concentrated_liquidity

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v32/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v32/x/concentrated-liquidity/types"
)

// ReplacePoolDenomUnsafe renames oldDenom to newDenom in the token pair of a concentrated liquidity
// pool. It is only for upgrade handlers.
//
// It moves no funds and leaves the spread reward and incentive accumulators alone, so spread rewards
// already earned in oldDenom stay claimable from the spread rewards address. The caller must already
// have exchanged the pool's whole oldDenom balance for the same amount of newDenom: the pool's
// liquidity and price are kept as they are, so the swap must be exactly 1:1.
func (k Keeper) ReplacePoolDenomUnsafe(ctx sdk.Context, poolId uint64, oldDenom, newDenom string) error {
	if oldDenom == "" || newDenom == "" || oldDenom == newDenom {
		return fmt.Errorf("invalid denom replacement %q -> %q", oldDenom, newDenom)
	}
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return err
	}
	clPool, ok := pool.(*model.Pool)
	if !ok {
		return fmt.Errorf("pool %d has unexpected type %T", poolId, pool)
	}
	switch {
	case clPool.Token0 == oldDenom && clPool.Token1 != newDenom:
		clPool.Token0 = newDenom
	case clPool.Token1 == oldDenom && clPool.Token0 != newDenom:
		clPool.Token1 = newDenom
	default:
		return fmt.Errorf("pool %d (%s, %s) cannot replace %s with %s", poolId, clPool.Token0, clPool.Token1, oldDenom, newDenom)
	}
	return k.setPool(ctx, clPool)
}

// GetPositionIdsForPool returns the IDs of the positions in poolId, in ascending order, read from the
// pool to position index. It returns an error if the pool has more than maxPositions positions.
func (k Keeper) GetPositionIdsForPool(ctx sdk.Context, poolId uint64, maxPositions int) ([]uint64, error) {
	prefix := types.KeyPoolPosition(poolId)
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), prefix)
	defer iterator.Close()

	positionIds := make([]uint64, 0, min(maxPositions, 1024))
	for ; iterator.Valid(); iterator.Next() {
		if len(positionIds) == maxPositions {
			return nil, fmt.Errorf("pool %d has more than %d positions", poolId, maxPositions)
		}
		// The key is prefix | KeySeparator | big-endian position ID (see KeyPoolPositionPositionId).
		key := iterator.Key()
		if len(key) != types.PoolPositionIDFullPrefixLen {
			return nil, fmt.Errorf("pool %d: invalid pool position key %x", poolId, key)
		}
		positionIds = append(positionIds, sdk.BigEndianToUint64(key[len(key)-types.Uint64ByteSize:]))
	}
	return positionIds, nil
}
