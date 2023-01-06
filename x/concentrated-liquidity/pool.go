package concentrated_liquidity

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// InitializePool initializes a concentrated liquidity pool and sets it in state.
func (k Keeper) InitializePool(ctx sdk.Context, poolI swaproutertypes.PoolI, creatorAddress sdk.AccAddress) error {
	concentratedPool, err := convertPoolInterfaceToConcentrated(poolI)
	if err != nil {
		return err
	}

	if err := k.createFeeAccumulator(ctx, concentratedPool.GetId()); err != nil {
		return err
	}

	tickSpacing := concentratedPool.GetTickSpacing()

	if !k.validateTickSpacing(ctx, tickSpacing) {
		return fmt.Errorf("invalid tick spacing. Got %d", tickSpacing)
	}

	return k.setPool(ctx, concentratedPool)
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
	found, err := osmoutils.Get(store, key, &pool)
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
	found, err := osmoutils.Get(store, key, &pool)
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
		return nil, fmt.Errorf("given pool does not implement ConcentratedPoolExtension, implements %T", poolI)
	}
	// Return the converted value
	return concentratedPool, nil
}

// validateTickSpacing returns true if the given tick spacing is one of the authorized tick spacings set in the
func (k Keeper) validateTickSpacing(ctx sdk.Context, tickSpacing uint64) bool {
	params := k.GetParams(ctx)
	for _, authorizedTick := range params.AuthorizedTickSpacing {
		if tickSpacing == authorizedTick {
			return true
		}
	}
	return false
}
