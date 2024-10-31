package concentrated_liquidity

import (
	"errors"
	"fmt"

	"cosmossdk.io/store/prefix"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	errorsmod "cosmossdk.io/errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// InitializePool initializes a new concentrated liquidity pool with the given PoolI interface and creator address.
// It validates tick spacing, spread factor, and authorized quote denominations before creating and setting
// the pool's fee and uptime accumulators. If the pool is successfully created, it calls the AfterConcentratedPoolCreated
// listener function.
//
// Returns an error if any of the following conditions are met:
// - The poolI cannot be converted to a ConcentratedPool.
// - The tick spacing is invalid.
// - The spread factor is invalid.
// - The quote denomination is unauthorized.
// - There is an error creating the fee or uptime accumulator.
// - There is an error setting the pool in the keeper's state.
func (k Keeper) InitializePool(ctx sdk.Context, poolI poolmanagertypes.PoolI, creatorAddress sdk.AccAddress) error {
	concentratedPool, err := asConcentrated(poolI)
	if err != nil {
		return err
	}

	params := k.GetParams(ctx)
	tickSpacing := concentratedPool.GetTickSpacing()
	spreadFactor := concentratedPool.GetSpreadFactor(ctx)
	poolId := concentratedPool.GetId()
	quoteAsset := concentratedPool.GetToken1()
	poolManagerParams := k.poolmanagerKeeper.GetParams(ctx)

	bypassRestrictions := false

	poolmanagerModuleAcc := k.accountKeeper.GetModuleAccount(ctx, poolmanagertypes.ModuleName).GetAddress()

	// allow pool manager module account to bypass restrictions (i.e. gov prop)
	if creatorAddress.Equals(poolmanagerModuleAcc) {
		bypassRestrictions = true
	}

	// allow whitelisted pool creators to bypass restrictions
	if !bypassRestrictions {
		for _, addr := range params.UnrestrictedPoolCreatorWhitelist {
			// okay to use MustAccAddressFromBech32 because already validated in params
			if sdk.MustAccAddressFromBech32(addr).Equals(creatorAddress) {
				bypassRestrictions = true
			}
		}
	}

	if !bypassRestrictions {
		if !k.IsPermissionlessPoolCreationEnabled(ctx) {
			return types.ErrPermissionlessPoolCreationDisabled
		}

		if !k.validateTickSpacing(params, tickSpacing) {
			return types.UnauthorizedTickSpacingError{ProvidedTickSpacing: tickSpacing, AuthorizedTickSpacings: params.AuthorizedTickSpacing}
		}

		if !k.validateSpreadFactor(params, spreadFactor) {
			return types.UnauthorizedSpreadFactorError{ProvidedSpreadFactor: spreadFactor, AuthorizedSpreadFactors: params.AuthorizedSpreadFactors}
		}

		if !validateAuthorizedQuoteDenoms(quoteAsset, poolManagerParams.AuthorizedQuoteDenoms) {
			return types.UnauthorizedQuoteDenomError{ProvidedQuoteDenom: quoteAsset, AuthorizedQuoteDenoms: poolManagerParams.AuthorizedQuoteDenoms}
		}
	}

	if err := k.createSpreadRewardAccumulator(ctx, poolId); err != nil {
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
	poolI, err := asPoolI(concentratedPool)
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

// Return true if the Pool has a position. This is guaranteed to be equi-satisfiable to
// the current sqrt price being 0.
// We also check that the current tick is 0, which is also guaranteed in this situation.
// That derisks any edge-case where the sqrt price is 0 but the tick is not 0.
func (k Keeper) PoolHasPosition(ctx sdk.Context, pool types.ConcentratedPoolExtension) bool {
	if pool.GetCurrentSqrtPrice().IsZero() && pool.GetCurrentTick() == 0 {
		return false
	}
	return true
}

func (k Keeper) CalculateSpotPrice(
	ctx sdk.Context,
	poolId uint64,
	quoteAssetDenom string,
	baseAssetDenom string,
) (spotPrice osmomath.BigDec, err error) {
	concentratedPool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return osmomath.BigDec{}, err
	}

	hasPositions := k.PoolHasPosition(ctx, concentratedPool)

	if !hasPositions {
		return osmomath.BigDec{}, types.NoSpotPriceWhenNoLiquidityError{PoolId: poolId}
	}

	price, err := concentratedPool.SpotPrice(ctx, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return osmomath.BigDec{}, err
	}

	if price.IsZero() {
		return osmomath.BigDec{}, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPriceV2, MaxSpotPrice: types.MaxSpotPrice}
	}
	if price.GT(types.MaxSpotPriceBigDec) || price.LT(types.MinSpotPriceBigDec) {
		return osmomath.BigDec{}, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPriceBigDec, MaxSpotPrice: types.MaxSpotPrice}
	}

	return price, nil
}

// GetTotalPoolLiquidity returns the coins in the pool owned by all LPs
func (k Keeper) GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return nil, err
	}

	addr := pool.GetAddress()
	token0Bal := k.bankKeeper.GetBalance(ctx, addr, pool.GetToken0())
	token1Bal := k.bankKeeper.GetBalance(ctx, addr, pool.GetToken1())

	return sdk.NewCoins(token0Bal, token1Bal), nil
}

// asPoolI takes a types.ConcentratedPoolExtension and attempts to convert it to a
// poolmanagertypes.PoolI. If the conversion is successful, the converted value is returned. If the conversion fails,
// an error is returned.
func asPoolI(concentratedPool types.ConcentratedPoolExtension) (poolmanagertypes.PoolI, error) {
	// Attempt to convert the concentratedPool to a poolmanagertypes.PoolI
	pool, ok := concentratedPool.(poolmanagertypes.PoolI)
	if !ok {
		// If the conversion fails, return an error
		return nil, fmt.Errorf("given pool does not implement CFMMPoolI, implements %T", pool)
	}
	// Return the converted value
	return pool, nil
}

// asConcentrated takes a poolmanagertypes.PoolI and attempts to convert it to a
// types.ConcentratedPoolExtension. If the conversion is successful, the converted value is returned. If the conversion fails,
// an error is returned.
func asConcentrated(poolI poolmanagertypes.PoolI) (types.ConcentratedPoolExtension, error) {
	// Attempt to convert poolmanagertypes.PoolI to a concentratedPool
	concentratedPool, ok := poolI.(types.ConcentratedPoolExtension)
	if !ok {
		// If the conversion fails, return an error
		return nil, fmt.Errorf("given pool does not implement ConcentratedPoolExtension, implements %T", poolI)
	}
	// Return the converted value
	return concentratedPool, nil
}

// GetConcentratedPoolById returns a concentrated pool interface associated with the given id.
// Returns error if fails to fetch the pool from the store.
func (k Keeper) GetConcentratedPoolById(ctx sdk.Context, poolId uint64) (types.ConcentratedPoolExtension, error) {
	poolI, err := k.GetPool(ctx, poolId)
	if err != nil {
		return nil, err
	}
	return asConcentrated(poolI)
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
		pool, err := k.GetConcentratedPoolById(ctx, poolIdToTickSpacingRecord.PoolId)
		if err != nil {
			return err
		}
		params := k.GetParams(ctx)

		if !k.validateTickSpacingUpdate(pool, params, poolIdToTickSpacingRecord.NewTickSpacing) {
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
func (k Keeper) validateTickSpacing(params types.Params, tickSpacing uint64) bool {
	for _, authorizedTick := range params.AuthorizedTickSpacing {
		if tickSpacing == authorizedTick {
			return true
		}
	}
	return false
}

// validateTickSpacingUpdate returns true if the given tick spacing is one of the authorized tick spacings set in the
// params and is less than the current tick spacing. False otherwise.
func (k Keeper) validateTickSpacingUpdate(pool types.ConcentratedPoolExtension, params types.Params, newTickSpacing uint64) bool {
	currentTickSpacing := pool.GetTickSpacing()
	for _, authorizedTick := range params.AuthorizedTickSpacing {
		// New tick spacing must be one of the authorized tick spacings and must be less than the current tick spacing
		if newTickSpacing == authorizedTick && newTickSpacing < currentTickSpacing {
			return true
		}
	}
	return false
}

// validateSpreadFactor returns true if the given spread factor is one of the authorized spread factors set in the
// params. False otherwise.
func (k Keeper) validateSpreadFactor(params types.Params, spreadFactor osmomath.Dec) bool {
	for _, authorizedSpreadFactor := range params.AuthorizedSpreadFactors {
		if spreadFactor.Equal(authorizedSpreadFactor) {
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
func validateAuthorizedQuoteDenoms(denom1 string, authorizedQuoteDenoms []string) bool {
	for _, authorizedQuoteDenom := range authorizedQuoteDenoms {
		if denom1 == authorizedQuoteDenom {
			return true
		}
	}
	return false
}

// GetLinkedBalancerPoolID is a wrapper function for gammKeeper.GetLinkedBalancerPoolID in order to allow
// the concentrated pool module to access the linked balancer pool id via query.
// Without this function, both pool link query functions would have to live in the gamm module which is unintuitive.
func (k Keeper) GetLinkedBalancerPoolID(ctx sdk.Context, concentratedPoolId uint64) (uint64, error) {
	return k.gammKeeper.GetLinkedBalancerPoolID(ctx, concentratedPoolId)
}

func (k Keeper) GetUserUnbondingPositions(ctx sdk.Context, address sdk.AccAddress) ([]model.PositionWithPeriodLock, error) {
	// Get the position IDs for the specified user address.
	positions, err := k.GetUserPositions(ctx, address, 0)
	if err != nil {
		return nil, err
	}

	// Query each position ID and determine if it has a lock ID associated with it.
	// Construct a response with the position as well as the lock's info.
	var userPositionsWithPeriodLocks []model.PositionWithPeriodLock
	for _, pos := range positions {
		lockId, err := k.GetLockIdFromPositionId(ctx, pos.PositionId)
		if errors.Is(err, types.PositionIdToLockNotFoundError{PositionId: pos.PositionId}) {
			continue
		} else if err != nil {
			return nil, err
		}
		// If we have hit this logic branch, it means that, at one point, the lockId provided existed. If we fetch it again
		// and it doesn't exist, that means that the lock has matured.
		lock, err := k.lockupKeeper.GetLockByID(ctx, lockId)
		if errors.Is(err, errorsmod.Wrap(lockuptypes.ErrLockupNotFound, fmt.Sprintf("lock with ID %d does not exist", lockId))) {
			continue
		}
		if err != nil {
			return nil, err
		}

		// Don't include locks that aren't unlocking
		if lock.EndTime.IsZero() {
			continue
		}

		userPositionsWithPeriodLocks = append(userPositionsWithPeriodLocks, model.PositionWithPeriodLock{
			Position: pos,
			Locks:    *lock,
		})
	}
	return userPositionsWithPeriodLocks, nil
}

// getPositionIDsByPoolID returns all position IDs for a given pool ID.
func (k Keeper) GetPositionIDsByPoolID(ctx sdk.Context, poolID uint64) ([]uint64, error) {
	key := types.KeyPoolPosition(poolID)
	key = append(key, types.KeySeparator...)
	positionIDs, err := osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), key, parsePositionIDFromPoolLink)
	if err != nil {
		return nil, err
	}

	return positionIDs, nil
}

// parsePositionIDFromPoolLink parses the position ID from the pool link key.
func parsePositionIDFromPoolLink(key []byte, _ []byte) (uint64, error) {
	if len(key) != types.PoolPositionIDFullPrefixLen {
		return 0, fmt.Errorf("length (%d) of key (%v) is not equal to expected (%d)", len(key), key, types.PoolPositionIDFullPrefixLen)
	}

	if key[types.PoolPositionIDKeySeparatorIndex] != types.KeySeparator[0] {
		return 0, fmt.Errorf("key (%v) is expected to have key separator (%v) at index (%d)", key, types.KeySeparator, types.PoolPositionIDKeySeparatorIndex)
	}

	positionID := sdk.BigEndianToUint64(key[types.PoolPositionIDKeySeparatorIndex+len(types.KeySeparator):])

	return positionID, nil
}

// MigrateIncentivesAccumulatorToScalingFactor multiplies the value of the uptime accumulator, respective position accumulators
// and tick uptime trackers by the per-unit liquidity scaling factor and overwrites the accumulators with the new values.
func (k Keeper) MigrateIncentivesAccumulatorToScalingFactor(ctx sdk.Context, poolId uint64) error {
	ctx.Logger().Info("migration start", "pool_id", poolId)
	// Get pool-global incentive accumulator
	uptimeAccums, err := k.GetUptimeAccumulators(ctx, poolId)
	if err != nil {
		return err
	}

	// Get all position IDs for the pool.
	positionIDs, err := k.GetPositionIDsByPoolID(ctx, poolId)
	if err != nil {
		return err
	}

	ctx.Logger().Info("num_positions", "count", len(positionIDs))

	// For each uptime accumulator, multiply the value by the per-unit liquidity scaling factor
	// and overwrite the accumulator with the new value.
	for uptimeIndex, uptimeAccum := range uptimeAccums {
		value := uptimeAccum.GetValue().MulDecTruncate(perUnitLiqScalingFactor)
		if err := accum.OverwriteAccumulatorUnsafe(ctx.KVStore(k.storeKey), types.KeyUptimeAccumulator(poolId, uint64(uptimeIndex)), value, uptimeAccum.GetTotalShares()); err != nil {
			return err
		}

		// For each position ID, multiply the value by the per-unit liquidity scaling factor
		// and overwrite the accumulator with the new value.
		for _, positionID := range positionIDs {
			positionPrefix := types.KeyPositionId(positionID)

			if !uptimeAccum.HasPosition(string(positionPrefix)) {
				return fmt.Errorf("position ID %d not found in uptime accumulator %d in pool %d", positionID, uptimeIndex, poolId)
			}

			positionSnapshot, err := uptimeAccum.GetPosition(string(positionPrefix))
			if err != nil {
				return err
			}

			positionSnapshotValue := positionSnapshot.GetAccumValuePerShare()

			// Multiply the value by the per-unit liquidity scaling factor
			newValue := positionSnapshotValue.MulDecTruncate(perUnitLiqScalingFactor)

			// Overwrite the position accumulator with the new value
			if err := uptimeAccum.SetPositionIntervalAccumulation(string(positionPrefix), newValue); err != nil {
				return err
			}
		}
	}

	// Retrieve all ticks for the pool
	ticks, err := k.GetAllInitializedTicksForPool(ctx, poolId)
	if err != nil {
		return err
	}

	ctx.Logger().Info("num_ticks", "count", len(ticks))

	// For each tick, multiply the value of the uptime accumulator tracker.
	for _, tick := range ticks {
		// Get the tick's accumulator
		uptimeTrackers := tick.Info.UptimeTrackers
		for i, uptimeTracker := range uptimeTrackers.List {
			uptimeTrackers.List[i].UptimeGrowthOutside = uptimeTracker.UptimeGrowthOutside.MulDecTruncate(perUnitLiqScalingFactor)
		}

		// Overwrite the tick's accumulator with the new value
		k.SetTickInfo(ctx, poolId, tick.TickIndex, &tick.Info)
	}

	ctx.Logger().Info("migration end", "pool_id", poolId)
	return nil
}

// MigrateSpreadFactorAccumulatorToScalingFactor multiplies the value of the spread reward accumulator, respective position accumulators
// and tick spread reward trackers by the per-unit liquidity scaling factor and overwrites the accumulators with the new values.
func (k Keeper) MigrateSpreadFactorAccumulatorToScalingFactor(ctx sdk.Context, poolId uint64) error {
	ctx.Logger().Info("migration start", "pool_id", poolId)
	// Get the spread reward accumulator for the pool.
	spreadRewardAccumulator, err := k.GetSpreadRewardAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	// Update the spread reward accumulator's value by multiplying it by the per-unit liquidity scaling factor.
	value := spreadRewardAccumulator.GetValue().MulDecTruncate(perUnitLiqScalingFactor)
	if err := accum.OverwriteAccumulatorUnsafe(ctx.KVStore(k.storeKey), types.KeySpreadRewardPoolAccumulator(poolId), value, spreadRewardAccumulator.GetTotalShares()); err != nil {
		return err
	}

	// Get all position IDs for the pool.
	positionIDs, err := k.GetPositionIDsByPoolID(ctx, poolId)
	if err != nil {
		return err
	}

	ctx.Logger().Info("num_positions", "count", len(positionIDs))

	// For each position ID, multiply the value by the per-unit liquidity scaling factor
	// and overwrite the accumulator with the new value.
	for _, positionId := range positionIDs {
		// Get the key for the position's accumulator in the spread reward accumulator.
		positionKey := types.KeySpreadRewardPositionAccumulator(positionId)

		// Check if the position exists in the spread reward accumulator.
		hasPosition := spreadRewardAccumulator.HasPosition(positionKey)
		if !hasPosition {
			return types.SpreadRewardPositionNotFoundError{PositionId: positionId}
		}

		// Get the position's current accumulator value per share from the spread reward accumulator.
		positionSnapshot, err := spreadRewardAccumulator.GetPosition(positionKey)
		if err != nil {
			return err
		}
		positionSnapshotValue := positionSnapshot.GetAccumValuePerShare()

		// Multiply the value by the per-unit liquidity scaling factor
		newValue := positionSnapshotValue.MulDecTruncate(perUnitLiqScalingFactor)

		// Overwrite the position accumulator with the new value
		if err := spreadRewardAccumulator.SetPositionIntervalAccumulation(positionKey, newValue); err != nil {
			return err
		}
	}

	// Retrieve all ticks for the pool
	ticks, err := k.GetAllInitializedTicksForPool(ctx, poolId)
	if err != nil {
		return err
	}

	ctx.Logger().Info("num_ticks", "count", len(ticks))

	// For each tick, scale the value of the spread reward accumulator tracker.
	for _, tick := range ticks {
		tick.Info.SpreadRewardGrowthOppositeDirectionOfLastTraversal = tick.Info.SpreadRewardGrowthOppositeDirectionOfLastTraversal.MulDecTruncate(perUnitLiqScalingFactor)

		// Overwrite the tick's accumulator with the new value
		k.SetTickInfo(ctx, poolId, tick.TickIndex, &tick.Info)
	}

	ctx.Logger().Info("migration end", "pool_id", poolId)
	return nil
}
