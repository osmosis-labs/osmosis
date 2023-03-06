package concentrated_liquidity

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// InitializePool initializes a concentrated liquidity pool and sets it in state.
func (k Keeper) InitializePool(ctx sdk.Context, poolI poolmanagertypes.PoolI, creatorAddress sdk.AccAddress) error {
	concentratedPool, err := convertPoolInterfaceToConcentrated(poolI)
	if err != nil {
		return err
	}

	if err := k.createFeeAccumulator(ctx, concentratedPool.GetId()); err != nil {
		return err
	}

	if err := k.createUptimeAccumulators(ctx, concentratedPool.GetId()); err != nil {
		return err
	}

	concentratedPool.SetLastLiquidityUpdate(ctx.BlockTime())

	params := k.GetParams(ctx)
	tickSpacing := concentratedPool.GetTickSpacing()
	swapFee := concentratedPool.GetSwapFee(ctx)

	if !k.validateTickSpacing(ctx, params, tickSpacing) {
		return fmt.Errorf("invalid tick spacing. Got %d", tickSpacing)
	}

	if !k.validateSwapFee(ctx, params, swapFee) {
		return fmt.Errorf("invalid swap fee. Got %s", swapFee)
	}

	return k.setPool(ctx, concentratedPool)
}

// GetPool returns a pool with a given id.
func (k Keeper) GetPool(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolI, error) {
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

func (k Keeper) GetAllPools(ctx sdk.Context) ([]types.ConcentratedPoolExtension, error) {
	return osmoutils.GatherValuesFromStorePrefix(
		ctx.KVStore(k.storeKey), types.PoolPrefix, func(value []byte) (types.ConcentratedPoolExtension, error) {
			pool := model.Pool{}
			err := k.cdc.Unmarshal(value, &pool)
			if err != nil {
				return nil, err
			}
			return &pool, nil
		},
	)
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

func (k Keeper) GetPoolDenoms(ctx sdk.Context, poolId uint64) ([]string, error) {
	concentratedPool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return nil, err
	}

	denoms := []string{concentratedPool.GetToken0(), concentratedPool.GetToken1()}
	return denoms, nil
}

func (k Keeper) CalculateSpotPrice(
	ctx sdk.Context,
	poolId uint64,
	quoteAssetDenom string,
	baseAssetDenom string,
) (spotPrice sdk.Dec, err error) {
	concentratedPool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	price := concentratedPool.GetCurrentSqrtPrice().Power(2)
	if price.IsZero() {
		return sdk.Dec{}, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice}
	}

	if quoteAssetDenom == concentratedPool.GetToken1() {
		price = sdk.OneDec().Quo(price)
	}

	if price.GT(types.MaxSpotPrice) || price.LT(types.MinSpotPrice) {
		return sdk.Dec{}, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice}
	}

	return price, nil
}

// convertConcentratedToPoolInterface takes a types.ConcentratedPoolExtension and attempts to convert it to a
// poolmanagertypes.PoolI. If the conversion is successful, the converted value is returned. If the conversion fails,
// an error is returned.
func convertConcentratedToPoolInterface(concentratedPool types.ConcentratedPoolExtension) (poolmanagertypes.PoolI, error) {
	// Attempt to convert the concentratedPool to a poolmanagertypes.PoolI
	pool, ok := concentratedPool.(poolmanagertypes.PoolI)
	if !ok {
		// If the conversion fails, return an error
		return nil, fmt.Errorf("given pool does not implement CFMMPoolI, implements %T", pool)
	}
	// Return the converted value
	return pool, nil
}

// convertPoolInterfaceToConcentrated takes a poolmanagertypes.PoolI and attempts to convert it to a
// types.ConcentratedPoolExtension. If the conversion is successful, the converted value is returned. If the conversion fails,
// an error is returned.
func convertPoolInterfaceToConcentrated(poolI poolmanagertypes.PoolI) (types.ConcentratedPoolExtension, error) {
	// Attempt to convert poolmanagertypes.PoolI to a concentratedPool
	concentratedPool, ok := poolI.(types.ConcentratedPoolExtension)
	if !ok {
		// If the conversion fails, return an error
		return nil, fmt.Errorf("given pool does not implement ConcentratedPoolExtension, implements %T", poolI)
	}
	// Return the converted value
	return concentratedPool, nil
}

func (k Keeper) GetPoolFromPoolIdAndConvertToConcentrated(ctx sdk.Context, poolId uint64) (types.ConcentratedPoolExtension, error) {
	poolI, err := k.GetPool(ctx, poolId)
	if err != nil {
		return nil, err
	}
	return convertPoolInterfaceToConcentrated(poolI)
}

// validateTickSpacing returns true if the given tick spacing is one of the authorized tick spacings set in the
// params. False otherwise.
func (k Keeper) validateTickSpacing(ctx sdk.Context, params types.Params, tickSpacing uint64) bool {
	for _, authorizedTick := range params.AuthorizedTickSpacing {
		if tickSpacing == authorizedTick {
			return true
		}
	}
	return false
}

// validateSwapFee returns true if the given swap fee is one of the authorized swap fees set in the
// params. False otherwise.
func (k Keeper) validateSwapFee(ctx sdk.Context, params types.Params, swapFee sdk.Dec) bool {
	for _, authorizedSwapFee := range params.AuthorizedSwapFees {
		if swapFee.Equal(authorizedSwapFee) {
			return true
		}
	}
	return false
}
