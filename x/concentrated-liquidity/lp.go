package concentrated_liquidity

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/swapstrategy"
	types "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

// createPosition creates a concentrated liquidity position in range between lowerTick and upperTick
// in a given `PoolId with the desired amount of each token. Since LPs are only allowed to provide
// liquidity proportional to the existing reserves, the actual amount of tokens used might differ from requested.
// As a result, LPs may also provide the minimum amount of each token to be used so that the system fails
// to create position if the desired amounts cannot be satisfied.
// On success, returns an actual amount of each token used and liquidity created.
// Returns error if:
// - the provided ticks are out of range / invalid
// - the pool provided does not exist
// - the liquidity delta is zero
// - the amount0 or amount1 returned from the position update is less than the given minimums
// - the pool or user does not have enough tokens to satisfy the requested amount
func (k Keeper) createPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, amount0Desired, amount1Desired, amount0Min, amount1Min sdk.Int, lowerTick, upperTick int64) (sdk.Int, sdk.Int, sdk.Dec, error) {
	// Retrieve the pool associated with the given pool ID.
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	// Check if the provided tick range is valid according to the pool's tick spacing and module parameters.
	if err := validateTickRangeIsValid(pool.GetTickSpacing(), lowerTick, upperTick); err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	// Transform the provided ticks into their corresponding square root prices.
	sqrtPriceLowerTick, sqrtPriceUpperTick, err := math.TicksToSqrtPrice(lowerTick, upperTick)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	// Create a cache context for the current transaction.
	// This allows us to make changes to the context without persisting it until later.
	// We only write the cache context (i.e. persist the changes) if the actual amounts returned
	// are greater than the given minimum amounts.
	cacheCtx, writeCacheCtx := ctx.CacheContext()
	initialSqrtPrice := pool.GetCurrentSqrtPrice()
	initialTick := pool.GetCurrentTick()

	// If the current square root price and current tick are zero, then this is the first position to be created for this pool.
	// In this case, we calculate the square root price and current tick based on the inputs of this position.
	if k.isInitialPosition(initialSqrtPrice, initialTick) {
		err := k.initializeInitialPosition(cacheCtx, pool, amount0Desired, amount1Desired)
		if err != nil {
			return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
		}
	}

	// Calculate the amount of liquidity that will be added to the pool by creating this position.
	liquidityDelta := math.GetLiquidityFromAmounts(pool.GetCurrentSqrtPrice(), sqrtPriceLowerTick, sqrtPriceUpperTick, amount0Desired, amount1Desired)
	if liquidityDelta.IsZero() {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, errors.New("liquidityDelta calculated equals zero")
	}

	// Update the position in the pool based on the provided tick range and liquidity delta.
	actualAmount0, actualAmount1, err := k.updatePosition(cacheCtx, poolId, owner, lowerTick, upperTick, liquidityDelta)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}
	// Check if the actual amounts of tokens 0 and 1 are greater than or equal to the given minimum amounts.
	if actualAmount0.LT(amount0Min) {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, types.InsufficientLiquidityCreatedError{Actual: actualAmount0, Minimum: amount0Min, IsTokenZero: true}
	}
	if actualAmount1.LT(amount1Min) {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, types.InsufficientLiquidityCreatedError{Actual: actualAmount1, Minimum: amount1Min}
	}

	// Transfer the actual amounts of tokens 0 and 1 from the position owner to the pool.
	err = k.sendCoinsBetweenPoolAndUser(cacheCtx, pool.GetToken0(), pool.GetToken1(), actualAmount0, actualAmount1, owner, pool.GetAddress())
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	// Persist the changes made to the cache context if the actual amounts of tokens 0 and 1 are greater than or equal to the given minimum amounts.
	writeCacheCtx()

	return actualAmount0, actualAmount1, liquidityDelta, nil
}

// withdrawPosition attempts to withdraw liquidityAmount from a position with the given pool id in the given tick range.
// On success, returns a positive amount of each token withdrawn.
// Returns error if
// - there is no position in the given tick ranges
// - if tick ranges are invalid
// - if attempts to withdraw an amount higher than originally provided in createPosition for a given range.
func (k Keeper) withdrawPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, requestedLiqudityAmountToWithdraw sdk.Dec) (amtDenom0, amtDenom1 sdk.Int, err error) {
	// Retrieve the pool associated with the given pool ID.
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// Check if the provided tick range is valid according to the pool's tick spacing and module parameters.
	if err := validateTickRangeIsValid(pool.GetTickSpacing(), lowerTick, upperTick); err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// Retrieve the position in the pool for the provided owner and tick range.
	position, err := k.getPosition(ctx, poolId, owner, lowerTick, upperTick)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// Check if the requested liquidity amount to withdraw is less than or equal to the available liquidity for the position.
	// If it is greater than the available liquidity, return an error.
	availableLiquidity := position.Liquidity
	if requestedLiqudityAmountToWithdraw.GT(availableLiquidity) {
		return sdk.Int{}, sdk.Int{}, types.InsufficientLiquidityError{Actual: requestedLiqudityAmountToWithdraw, Available: availableLiquidity}
	}

	// Calculate the change in liquidity for the pool based on the requested amount to withdraw.
	// This amount is negative because that liquidity is being withdrawn from the pool.
	liquidityDelta := requestedLiqudityAmountToWithdraw.Neg()

	// Update the position in the pool based on the provided tick range and liquidity delta.
	actualAmount0, actualAmount1, err := k.updatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// Transfer the actual amounts of tokens 0 and 1 from the pool to the position owner.
	err = k.sendCoinsBetweenPoolAndUser(ctx, pool.GetToken0(), pool.GetToken1(), actualAmount0, actualAmount1, pool.GetAddress(), owner)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	return actualAmount0.Neg(), actualAmount1.Neg(), nil
}

// updatePosition updates the position in the given pool id and in the given tick range and liquidityAmount.
// Negative liquidityDelta implies withdrawing liquidity.
// Positive liquidityDelta implies adding liquidity.
// Updates ticks and pool liquidity. Returns how much of each token is either added or removed.
// Negative returned amounts imply that tokens are removed from the pool.
// Positive returned amounts imply that tokens are added to the pool.
// TODO: tests.
func (k Keeper) updatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec) (sdk.Int, sdk.Int, error) {
	// update tickInfo state
	// TODO: come back to sdk.Int vs sdk.Dec state & truncation
	err := k.initOrUpdateTick(ctx, poolId, lowerTick, liquidityDelta, false)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// TODO: come back to sdk.Int vs sdk.Dec state & truncation
	err = k.initOrUpdateTick(ctx, poolId, upperTick, liquidityDelta, true)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// update position state
	// TODO: come back to sdk.Int vs sdk.Dec state & truncation
	err = k.initOrUpdatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// now calculate amount for token0 and token1
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	sqrtPriceLowerTick, sqrtPriceUpperTick, err := math.TicksToSqrtPrice(lowerTick, upperTick)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	actualAmount0, actualAmount1 := pool.CalcActualAmounts(ctx, lowerTick, upperTick, sqrtPriceLowerTick, sqrtPriceUpperTick, liquidityDelta)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	pool.UpdateLiquidityIfActivePosition(ctx, lowerTick, upperTick, liquidityDelta)

	if err := k.setPool(ctx, pool); err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// The returned amounts are rounded down to avoid returning more to clients than they actually deposited.
	return actualAmount0.TruncateInt(), actualAmount1.TruncateInt(), nil
}

// sendCoinsBetweenPoolAndUser takes the amounts calculated from a join/exit position and executes the send between pool and user
func (k Keeper) sendCoinsBetweenPoolAndUser(ctx sdk.Context, denom0, denom1 string, amount0, amount1 sdk.Int, sender, receiver sdk.AccAddress) error {
	var finalCoinsToSend sdk.Coins
	if amount0.IsPositive() {
		finalCoinsToSend = append(finalCoinsToSend, sdk.NewCoin(denom0, amount0))
	}
	if amount1.IsPositive() {
		finalCoinsToSend = append(finalCoinsToSend, sdk.NewCoin(denom1, amount1))
	}
	err := k.bankKeeper.SendCoins(ctx, sender, receiver, finalCoinsToSend)
	if err != nil {
		return err
	}
	return nil
}

// isInitialPosition checks if the initial sqrtPrice and initial tick are equal to zero.
// If so, this is the first position to be created for this pool, and we return true.
// If not, we return false.
func (k Keeper) isInitialPosition(initialSqrtPrice sdk.Dec, initialTick sdk.Int) bool {
	if initialSqrtPrice.Equal(sdk.ZeroDec()) && initialTick.Equal(sdk.ZeroInt()) {
		return true
	}
	return false
}

// createInitialPosition ensures that the first position created on this pool includes both asset0 and asset1
// This is required so we can set the pool's sqrtPrice and calculate it's initial tick from this
func (k Keeper) initializeInitialPosition(ctx sdk.Context, pool types.ConcentratedPoolExtension, amount0Desired, amount1Desired sdk.Int) error {
	// Check that the position includes some amount of both asset0 and asset1
	if !amount0Desired.GT(sdk.ZeroInt()) || !amount1Desired.GT(sdk.ZeroInt()) {
		return types.InitialLiquidityZeroError{Amount0: amount0Desired, Amount1: amount1Desired}
	}

	// Calculate the spot price and sqrt price from the amount provided
	initialSpotPrice := amount1Desired.ToDec().Quo(amount0Desired.ToDec())
	initialSqrtPrice, err := initialSpotPrice.ApproxSqrt()
	if err != nil {
		return err
	}

	// Calculate the initial tick from the initial spot price
	fmt.Printf("initialSpotPrice %v \n", initialSpotPrice)
	initialTick := math.PriceToTick(initialSpotPrice)
	fmt.Printf("initialTick %v \n", initialTick)

	// Set the pool's current sqrt price and current tick to the above calculated values
	pool.SetCurrentSqrtPrice(initialSqrtPrice)
	pool.SetCurrentTick(initialTick)
	err = k.setPool(ctx, pool)
	if err != nil {
		return err
	}
	return nil
}

// GetSecondsPerLiquidityInside returns the seconds per liquidity between two ticks in a given pool
func (k Keeper) GetSecondsPerLiquidityInside(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64) (sdk.Dec, error) {
	// Get the Seconds Per Liquidity Outside for the lower tick
	lowerTickInfo, err := k.getTickInfo(ctx, poolId, lowerTick)
	if err != nil {
		return sdk.Dec{}, err
	}

	// Get the Seconds Per Liquidity Outside for the upper tick
	upperTickInfo, err := k.getTickInfo(ctx, poolId, upperTick)
	if err != nil {
		return sdk.Dec{}, err
	}

	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	globalSecondsPerLiquidity := pool.GetGlobalSecondsPerLiquidity()

	// If tick is active, we need to calculate a new seconds per liquidity inside since we only update these when crossing ticks
	if pool.GetCurrentTick().GTE(sdk.NewInt(lowerTick)) || pool.GetCurrentTick().LT(sdk.NewInt(upperTick)) {
		// Get the current tick
		currentTick := pool.GetCurrentTick().Int64()

		// Utilize a zeroForOne swap strategy since we want to look left or equal to the current tick
		swapStrategy := swapstrategy.New(true, sdk.NewDec(99999999), k.storeKey)
		lastInitializedTick, _ := swapStrategy.NextInitializedTick(ctx, pool.GetId(), currentTick)
		lastInitializedTickInfo, err := k.getTickInfo(ctx, poolId, lastInitializedTick.Int64())
		if err != nil {
			return sdk.Dec{}, err
		}

		// Determine the seconds that have passed since the last time the last initialized tick was crossed
		// Update the global time with this difference and calculate the new global seconds per liquidity outside
		newSecondsInactive := time.Duration(ctx.BlockTime().Sub(pool.GetTimeOfCreation())) + lastInitializedTickInfo.SecondsInactive
		newSecondsPerLiquidityOutside := sdk.MustNewDecFromStr(fmt.Sprintf("%f", newSecondsInactive.Seconds())).Quo(lastInitializedTickInfo.LiquidityGross)
		globalSecondsPerLiquidity = globalSecondsPerLiquidity.Add(newSecondsPerLiquidityOutside)
	}

	secondsPerLiquidityInside := globalSecondsPerLiquidity.Sub(lowerTickInfo.SecondsPerLiquidityOutside).Sub(upperTickInfo.SecondsPerLiquidityOutside)
	return secondsPerLiquidityInside, nil
}
