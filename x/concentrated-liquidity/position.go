package concentrated_liquidity

import (
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
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	joinTime time.Time,
	freezeDuration time.Duration,
	positionId uint64,
) (sdk.Dec, error) {
	if !k.poolExists(ctx, poolId) {
		return sdk.Dec{}, types.PoolNotFoundError{PoolId: poolId}
	}
	if k.hasFullPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, freezeDuration, positionId) {
		positionLiquidity, err := k.GetPositionLiquidity(ctx, poolId, owner, lowerTick, upperTick, joinTime, freezeDuration, positionId)
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
	liquidity, err := k.getOrInitPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, freezeDuration, positionId)
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

func (k Keeper) hasFullPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, joinTime time.Time, freezeDuration time.Duration, positionId uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyFullPosition(poolId, owner, lowerTick, upperTick, joinTime, freezeDuration, positionId)
	return store.Has(key)
}

// GetPositionLiquidity checks if a position exists at the provided upper and lower ticks and freezeDuration time for the given owner. Returns position if found.
func (k Keeper) GetPositionLiquidity(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, joinTime time.Time, freezeDuration time.Duration, positionId uint64) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyFullPosition(poolId, owner, lowerTick, upperTick, joinTime, freezeDuration, positionId)

	liquidityStruct := &sdk.DecProto{}
	found, err := osmoutils.Get(store, key, liquidityStruct)
	if err != nil {
		return sdk.Dec{}, err
	}

	if !found {
		return sdk.Dec{}, types.PositionNotFoundError{PoolId: poolId, LowerTick: lowerTick, UpperTick: upperTick, JoinTime: joinTime, FreezeDuration: freezeDuration}
	}

	return liquidityStruct.Dec, nil
}

// GetUserPositions gets all the existing user positions, with the option to filter by a specific pool.
func (k Keeper) GetUserPositions(ctx sdk.Context, addr sdk.AccAddress, poolId uint64) ([]model.Position, error) {
	if poolId == 0 {
		return osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), types.KeyUserPositions(addr), ParseFullPositionFromBytes)
	} else {
		return osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), types.KeyAddressAndPoolId(addr, poolId), ParseFullPositionFromBytes)
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
	key := types.KeyFullPosition(poolId, owner, lowerTick, upperTick, joinTime, freezeDuration, positionId)
	osmoutils.MustSetDec(store, key, liquidity)
}

func (k Keeper) deletePosition(ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	joinTime time.Time,
	freezeDuration time.Duration,
	positionId uint64,
) error {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyFullPosition(poolId, owner, lowerTick, upperTick, joinTime, freezeDuration, positionId)

	if !store.Has(key) {
		return types.PositionNotFoundError{PoolId: poolId, LowerTick: lowerTick, UpperTick: upperTick, JoinTime: joinTime, FreezeDuration: freezeDuration}
	}

	store.Delete(key)
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
