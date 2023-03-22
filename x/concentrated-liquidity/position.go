package concentrated_liquidity

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var emptyOptions = &accum.Options{}

// getOrInitPosition retrieves the position's liquidity for the given tick range.
// If it doesn't exist, it returns zero.
func (k Keeper) getOrInitPosition(
	ctx sdk.Context,
	positionId uint64,
) (sdk.Dec, error) {
	if k.hasFullPosition(ctx, positionId) {
		positionLiquidity, err := k.GetPositionLiquidity(ctx, positionId)
		if err != nil {
			return sdk.Dec{}, err
		}
		return positionLiquidity, nil
	}
	return sdk.ZeroDec(), nil
}

// initOrUpdatePosition checks to see if the specified owner has an existing position at the given tick range.
// If a position is not present, it initializes the position with the provided liquidity delta.
// If a position is present, it combines the existing liquidity in that position with the provided liquidity delta. It also
// bumps up all uptime accumulators to current time, including the ones the new position isn't eligible for.
func (k Keeper) initOrUpdatePosition(
	ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	liquidityDelta sdk.Dec,
	joinTime time.Time,
	freezeDuration time.Duration,
	positionId uint64,
) (err error) {
	liquidity, err := k.getOrInitPosition(ctx, positionId)
	if err != nil {
		return err
	}

	// note that liquidityIn can be either positive or negative.
	// If negative, this would work as a subtraction from liquidityBefore
	liquidity = liquidity.Add(liquidityDelta)
	if liquidity.IsNegative() {
		return types.NegativeLiquidityError{Liquidity: liquidity}
	}

	err = k.initOrUpdatePositionUptime(ctx, poolId, liquidity, owner, lowerTick, upperTick, liquidityDelta, joinTime, freezeDuration, positionId)
	if err != nil {
		return err
	}

	k.setPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, freezeDuration, liquidity, positionId)
	return nil
}

func (k Keeper) hasFullPosition(ctx sdk.Context, positionId uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPositionId(positionId)
	return store.Has(key)
}

// GetPositionLiquidity checks if a position exists at the provided upper and lower ticks and freezeDuration time for the given owner. Returns position if found.
func (k Keeper) GetPositionLiquidity(ctx sdk.Context, positionId uint64) (sdk.Dec, error) {
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return sdk.Dec{}, err
	}

	return position.Liquidity, nil
}

// GetPosition checks if the given position id exists. Returns position if found.
func (k Keeper) GetPosition(ctx sdk.Context, positionId uint64) (model.Position, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPositionId(positionId)

	positionStruct := &model.Position{}
	found, err := osmoutils.Get(store, key, positionStruct)
	if err != nil {
		return model.Position{}, err
	}

	if !found {
		return model.Position{}, types.PositionIdNotFoundError{PositionId: positionId}
	}

	return *positionStruct, nil
}

// GetUserPositions gets all the existing user positions, with the option to filter by a specific pool.
func (k Keeper) GetUserPositions(ctx sdk.Context, addr sdk.AccAddress, poolId uint64) ([]model.Position, error) {
	if poolId == 0 {
		positions := []model.Position{}
		positionIds, err := osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), types.KeyUserPositions(addr), ParsePositionIdFromBz)
		if err != nil {
			return nil, err
		}
		for _, positionId := range positionIds {
			position, err := k.GetPosition(ctx, positionId)
			if err != nil {
				return nil, err
			}
			positions = append(positions, position)
		}
		return positions, nil
	} else {
		positions := []model.Position{}
		positionIds, err := osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), types.KeyAddressAndPoolId(addr, poolId), ParsePositionIdFromBz)
		if err != nil {
			return nil, err
		}
		for _, positionId := range positionIds {
			position, err := k.GetPosition(ctx, positionId)
			if err != nil {
				return nil, err
			}
			positions = append(positions, position)
		}
		return positions, nil
	}
}

// ParsePositionFromBz parses bytes into a position struct. Returns a parsed position and nil on success.
// Returns error if bytes length is zero or if fails to parse the given bytes into the position struct.
func (k Keeper) setPosition(ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	joinTime time.Time,
	freezeDuration time.Duration,
	liquidity sdk.Dec,
	positionId uint64,
) {
	store := ctx.KVStore(k.storeKey)
	position := model.Position{PositionId: positionId, PoolId: poolId, Address: owner.String(), LowerTick: lowerTick, UpperTick: upperTick, JoinTime: joinTime, FreezeDuration: freezeDuration, Liquidity: liquidity}
	key := types.KeyPositionId(positionId)
	osmoutils.MustSet(store, key, &position)

	key = types.KeyAddressPoolIdPositionId(owner, poolId, positionId)
	store.Set(key, sdk.Uint64ToBigEndian(positionId))

	// key = types.KeyPoolPositionId(poolId, positionId)
	// store.Set(key, sdk.Uint64ToBigEndian(positionId))
}

func (k Keeper) deletePosition(ctx sdk.Context,
	positionId uint64,
	owner sdk.AccAddress,
	poolId uint64,
) error {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPositionId(positionId)

	if !store.Has(key) {
		return types.PositionIdNotFoundError{PositionId: positionId}
	}

	store.Delete(key)

	storeTwo := ctx.KVStore(k.storeKey)
	keyTwo := types.KeyAddressPoolIdPositionId(owner, poolId, positionId)
	if !storeTwo.Has(keyTwo) {
		return fmt.Errorf("position id %d not found for address %s and pool id %d", positionId, owner.String(), poolId)
	}

	storeTwo.Delete(keyTwo)
	return nil
}

// CreateFullRangePosition creates a full range (min to max tick) concentrated liquidity position for the given pool ID, owner, coins, and frozen until time.
// The function returns the amounts of token 0 and token 1, and the liquidity created from the position.
func (k Keeper) CreateFullRangePosition(ctx sdk.Context, concentratedPool types.ConcentratedPoolExtension, owner sdk.AccAddress, coins sdk.Coins, freezeDuration time.Duration) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, err error) {
	// Determine the max and min ticks for the concentrated pool we are migrating to.
	minTick, maxTick := GetMinAndMaxTicksFromExponentAtPriceOne(concentratedPool.GetPrecisionFactorAtPriceOne())

	// Create a full range (min to max tick) concentrated liquidity position.
	positionId, amount0, amount1, liquidity, joinTime, err = k.createPosition(ctx, concentratedPool.GetId(), owner, coins.AmountOf(concentratedPool.GetToken0()), coins.AmountOf(concentratedPool.GetToken1()), sdk.ZeroInt(), sdk.ZeroInt(), minTick, maxTick, freezeDuration)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, err
	}

	return positionId, amount0, amount1, liquidity, joinTime, nil
}

func CalculateUnderlyingAssetsFromPosition(ctx sdk.Context, position model.Position, pool types.ConcentratedPoolExtension) (sdk.Dec, sdk.Dec, error) {
	// Transform the provided ticks into their corresponding sqrtPrices.
	sqrtPriceLowerTick, sqrtPriceUpperTick, err := math.TicksToSqrtPrice(position.LowerTick, position.UpperTick, pool.GetPrecisionFactorAtPriceOne())
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}

	// Calculate the amount of underlying assets in the position
	asset0, asset1 := pool.CalcActualAmounts(ctx, position.LowerTick, position.UpperTick, sqrtPriceLowerTick, sqrtPriceUpperTick, position.Liquidity)
	return asset0, asset1, nil
}

// getNextPositionIdAndIncrement returns the next position Id, and increments the corresponding state entry.
func (k Keeper) getNextPositionIdAndIncrement(ctx sdk.Context) uint64 {
	nextPositionId := k.GetNextPositionId(ctx)
	k.SetNextPositionId(ctx, nextPositionId+1)
	return nextPositionId
}
