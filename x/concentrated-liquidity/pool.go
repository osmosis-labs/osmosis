package concentrated_liquidity

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/osmoutils"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/model"
	types "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// GetPool returns a pool with a given id.
func (k Keeper) GetPool(ctx sdk.Context, poolId uint64) (swaproutertypes.PoolI, error) {
	return nil, errors.New("not implemented")
}

// TODO: spec and test
func (k Keeper) getPoolById(ctx sdk.Context, poolId uint64) (types.ConcentratedPoolExtension, error) {
	store := ctx.KVStore(k.storeKey)
	pool := model.Pool{}
	key := types.KeyPool(poolId)
	found, err := osmoutils.GetIfFound(store, key, &pool)
	if err != nil {
		panic(err)
	}
	if !found {
		return nil, types.PoolNotFoundError{PoolId: poolId}
	}
	return &pool, nil
}

// TODO: spec and test
func (k Keeper) setPool(ctx sdk.Context, pool types.ConcentratedPoolExtension) error {
	poolModel, ok := pool.(*model.Pool)
	if !ok {
		return errors.New("invalid pool type when setting concentreated pool")
	}
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPool(pool.GetId())
	osmoutils.MustSet(store, key, poolModel)
	return nil
}
