package concentrated_liquidity

import (
	"errors"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// InitializePool initializes a new concentrated liquidity pool with the given PoolI interface and creator address.
// It validates tick spacing, swap fee, and authorized quote denominations before creating and setting
// the pool's fee and uptime accumulators. If the pool is successfully created, it calls the AfterConcentratedPoolCreated
// listener function.
//
// Returns an error if any of the following conditions are met:
// - The poolI cannot be converted to a ConcentratedPool.
// - The tick spacing is invalid.
// - The swap fee is invalid.
// - The quote denomination is unauthorized.
// - There is an error creating the fee or uptime accumulator.
// - There is an error setting the pool in the keeper's state.
func (k Keeper) InitializePool(ctx sdk.Context, poolI poolmanagertypes.PoolI, creatorAddress sdk.AccAddress) error {
	concentratedPool, err := convertPoolInterfaceToConcentrated(poolI)
	if err != nil {
		return err
	}

	params := k.GetParams(ctx)
	tickSpacing := concentratedPool.GetTickSpacing()
	swapFee := concentratedPool.GetSwapFee(ctx)
	poolId := concentratedPool.GetId()
	quoteAsset := concentratedPool.GetToken1()

	if !k.validateTickSpacing(ctx, params, tickSpacing) {
		return types.UnauthorizedTickSpacingError{ProvidedTickSpacing: tickSpacing, AuthorizedTickSpacings: params.AuthorizedTickSpacing}
	}

	if !k.validateSwapFee(ctx, params, swapFee) {
		return types.UnauthorizedSwapFeeError{ProvidedSwapFee: swapFee, AuthorizedSwapFees: params.AuthorizedSwapFees}
	}

	if !validateAuthorizedQuoteDenoms(ctx, quoteAsset, params.AuthorizedQuoteDenoms) {
		return types.UnauthorizedQuoteDenomError{ProvidedQuoteDenom: quoteAsset, AuthorizedQuoteDenoms: params.AuthorizedQuoteDenoms}
	}

	if err := k.createFeeAccumulator(ctx, poolId); err != nil {
		return err
	}

	if err := k.createUptimeAccumulators(ctx, poolId); err != nil {
		return err
	}

	concentratedPool.SetLastLiquidityUpdate(ctx.BlockTime())

	if err := k.setPool(ctx, concentratedPool); err != nil {
		return err
	}

	k.listeners.AfterConcentratedPoolCreated(ctx, creatorAddress, poolId)

	return nil
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

func (k Keeper) GetPools(ctx sdk.Context) ([]poolmanagertypes.PoolI, error) {
	return osmoutils.GatherValuesFromStorePrefix(
		ctx.KVStore(k.storeKey), types.PoolPrefix, func(value []byte) (poolmanagertypes.PoolI, error) {
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

// setPool stores a ConcentratedPoolExtension in the Keeper's KVStore.
// It returns an error if the provided pool is not of type *model.Pool.
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

	hasPositions, err := k.HasAnyPositionForPool(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	if !hasPositions {
		return sdk.Dec{}, types.NoSpotPriceWhenNoLiquidityError{PoolId: poolId}
	}

	price, err := concentratedPool.SpotPrice(ctx, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	if price.IsZero() {
		return sdk.Dec{}, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice}
	}
	if price.GT(types.MaxSpotPrice) || price.LT(types.MinSpotPrice) {
		return sdk.Dec{}, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice}
	}

	return price, nil
}

// GetTotalPoolLiquidity returns the coins in the pool owned by all LPs
func (k Keeper) GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return nil, err
	}

	poolBalance := k.bankKeeper.GetAllBalances(ctx, pool.GetAddress())

	// This is to ensure that malicious actor cannot send dust to
	// a pool address.
	filteredPoolBalance := poolBalance.FilterDenoms([]string{pool.GetToken0(), pool.GetToken1()})

	return filteredPoolBalance, nil
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

func (k Keeper) GetSerializedPools(ctx sdk.Context, pagination *query.PageRequest) ([]*codectypes.Any, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	poolStore := prefix.NewStore(store, types.PoolPrefix)

	var anys []*codectypes.Any
	pageRes, err := query.Paginate(poolStore, pagination, func(key, _ []byte) error {
		pool := model.Pool{}
		// Get the next pool from the poolStore and pass it to the pool variable
		_, err := osmoutils.Get(poolStore, key, &pool)
		if err != nil {
			return err
		}

		// Retrieve the poolInterface from the respective pool
		poolI, err := k.GetPool(ctx, pool.GetId())
		if err != nil {
			return err
		}

		any, err := codectypes.NewAnyWithValue(poolI)
		if err != nil {
			return err
		}

		anys = append(anys, any)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return anys, pageRes, err
}

// DecreaseConcentratedPoolTickSpacing decreases the tick spacing of the given pools to the given tick spacings.
// This effectively increases the number of initializable ticks in the pool by reducing the number of ticks we skip over when traversing up and down.
// It returns an error if the tick spacing is not one of the authorized tick spacings or is not less than the current tick spacing of the respective pool.
func (k Keeper) DecreaseConcentratedPoolTickSpacing(ctx sdk.Context, poolIdToTickSpacingRecord []types.PoolIdToTickSpacingRecord) error {
	for _, poolIdToTickSpacingRecord := range poolIdToTickSpacingRecord {
		pool, err := k.GetPoolFromPoolIdAndConvertToConcentrated(ctx, poolIdToTickSpacingRecord.PoolId)
		if err != nil {
			return err
		}
		params := k.GetParams(ctx)

		if !k.validateTickSpacingUpdate(ctx, pool, params, poolIdToTickSpacingRecord.NewTickSpacing) {
			return fmt.Errorf("tick spacing %d is not valid", poolIdToTickSpacingRecord.NewTickSpacing)
		}

		pool.SetTickSpacing(poolIdToTickSpacingRecord.NewTickSpacing)
		err = k.setPool(ctx, pool)
		if err != nil {
			return err
		}
	}
	return nil
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

// validateTickSpacingUpdate returns true if the given tick spacing is one of the authorized tick spacings set in the
// params and is less than the current tick spacing. False otherwise.
func (k Keeper) validateTickSpacingUpdate(ctx sdk.Context, pool types.ConcentratedPoolExtension, params types.Params, newTickSpacing uint64) bool {
	currentTickSpacing := pool.GetTickSpacing()
	for _, authorizedTick := range params.AuthorizedTickSpacing {
		// New tick spacing must be one of the authorized tick spacings and must be less than the current tick spacing
		if newTickSpacing == authorizedTick && newTickSpacing < currentTickSpacing {
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

// validateAuthorizedQuoteDenoms validates if a given denom1 is present in the authorized quote denoms list
// It returns a boolean indicating if the denom1 is authorized or not.
//
// Parameters:
// - ctx: sdk.Context - The context object
// - denom1: string - The denom1 string to be checked
// - authorizedQuoteDenoms: []string - The list of authorized quote denoms
//
// Returns:
// - bool: A boolean indicating if the denom1 is authorized or not.
func validateAuthorizedQuoteDenoms(ctx sdk.Context, denom1 string, authorizedQuoteDenoms []string) bool {
	for _, authorizedQuoteDenom := range authorizedQuoteDenoms {
		if denom1 == authorizedQuoteDenom {
			return true
		}
	}
	return false
}
