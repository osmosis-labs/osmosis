package concentrated_liquidity

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/osmoutils"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/model"
	types "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// CreateNewConcentratedLiquidityPool creates a new concentrated liquidity pool with the given parameters.
// The pool tokens are denom0 and denom1, and are ordered such that denom0 is lexicographically smaller than denom1.
// The pool is created with zero liquidity and the initial sqrt price and current tick set to zero.
// The given token denominations are ordered to ensure that the first token is the numerator of the price, and the second token is the denominator of the price.
// The pool is added to the pool store, and an error is returned if the operation fails.
func (k Keeper) CreateNewConcentratedLiquidityPool(
	ctx sdk.Context,
	poolId uint64,
	denom0, denom1 string,
	tickSpacing uint64,
) (types.ConcentratedPoolExtension, error) {
	// Order the initial pool denoms so that denom0 is lexicographically smaller than denom1.
	denom0, denom1, err := types.OrderInitialPoolDenoms(denom0, denom1)
	if err != nil {
		return nil, err
	}

	// Create a new concentrated liquidity pool with the given parameters.
	poolI, err := model.NewConcentratedLiquidityPool(poolId, denom0, denom1, tickSpacing)
	if err != nil {
		return nil, err
	}

	// Add the pool to the pool store.
	err = k.setPool(ctx, &poolI)
	if err != nil {
		return nil, err
	}

	return &poolI, nil
}

// GetPool returns a pool with a given id.
func (k Keeper) GetPool(ctx sdk.Context, poolId uint64) (swaproutertypes.PoolI, error) {
	concentratedPool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return nil, types.PoolNotFoundError{PoolId: poolId}
	}
	poolI, err := convertConcentratedToPoolInterface(concentratedPool)
	if err != nil {
		return nil, err
	}
	return poolI, nil
}

// getPoolById returns a concentratedPoolExtension that corresponds to the requested pool id. Returns error if pool id is not found.
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

// poolExists returns true if a pool with the given id exists. False otherwise.
func (k Keeper) poolExists(ctx sdk.Context, poolId uint64) bool {
	store := ctx.KVStore(k.storeKey)
	pool := model.Pool{}
	key := types.KeyPool(poolId)
	found, err := osmoutils.GetIfFound(store, key, &pool)
	if err != nil {
		panic(err)
	}
	return found
}

// TODO: spec and test
func (k Keeper) setPool(ctx sdk.Context, pool types.ConcentratedPoolExtension) error {
	poolModel, ok := pool.(*model.Pool)
	if !ok {
		return errors.New("invalid pool type when setting concentrated pool")
	}
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPool(pool.GetId())
	osmoutils.MustSet(store, key, poolModel)
	return nil
}

// convertConcentratedToPoolInterface takes a types.ConcentratedPoolExtension and attempts to convert it to a
// swaproutertypes.PoolI. If the conversion is successful, the converted value is returned. If the conversion fails,
// an error is returned.
func convertConcentratedToPoolInterface(concentratedPool types.ConcentratedPoolExtension) (swaproutertypes.PoolI, error) {
	// Attempt to convert the concentratedPool to a swaproutertypes.PoolI
	pool, ok := concentratedPool.(swaproutertypes.PoolI)
	if !ok {
		// If the conversion fails, return an error
		return nil, fmt.Errorf("given pool does not implement CFMMPoolI, implements %T", pool)
	}
	// Return the converted value
	return pool, nil
}

// convertPoolInterfaceToConcentrated takes a swaproutertypes.PoolI and attempts to convert it to a
// types.ConcentratedPoolExtension. If the conversion is successful, the converted value is returned. If the conversion fails,
// an error is returned.
func convertPoolInterfaceToConcentrated(poolI swaproutertypes.PoolI) (types.ConcentratedPoolExtension, error) {
	// Attempt to convert swaproutertypes.PoolI to a concentratedPool
	concentratedPool, ok := poolI.(types.ConcentratedPoolExtension)
	if !ok {
		// If the conversion fails, return an error
		return nil, fmt.Errorf("given pool does not implement ConcentratedPoolExtension, implements %T", concentratedPool)
	}
	// Return the converted value
	return concentratedPool, nil
}
