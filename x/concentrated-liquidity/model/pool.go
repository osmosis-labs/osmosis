package model

import (
	"encoding/json"
	fmt "fmt"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

const (
	incentivesAddressPrefix = "incentives"
)

var (
	_   types.ConcentratedPoolExtension = &Pool{}
	one                                 = sdk.OneDec()
)

// NewConcentratedLiquidityPool creates a new ConcentratedLiquidity pool with the specified parameters.
// The two provided denoms are ordered so that denom0 is lexicographically smaller than denom1.
func NewConcentratedLiquidityPool(poolId uint64, denom0, denom1 string, tickSpacing uint64, swapFee sdk.Dec) (Pool, error) {
	// Ensure that the two denoms are different
	if denom0 == denom1 {
		return Pool{}, types.MatchingDenomError{Denom: denom0}
	}

	// Swap fee must be [0,1)
	if swapFee.IsNegative() || swapFee.GTE(one) {
		return Pool{}, types.InvalidSwapFeeError{ActualFee: swapFee}
	}

	// Create a new pool struct with the specified parameters
	pool := Pool{
		Address:              poolmanagertypes.NewPoolAddress(poolId).String(),
		IncentivesAddress:    osmoutils.NewModuleAddressWithPrefix(types.ModuleName, incentivesAddressPrefix, sdk.Uint64ToBigEndian(poolId)).String(),
		Id:                   poolId,
		CurrentSqrtPrice:     sdk.ZeroDec(),
		CurrentTick:          sdk.ZeroInt(),
		CurrentTickLiquidity: sdk.ZeroDec(),
		Token0:               denom0,
		Token1:               denom1,
		TickSpacing:          tickSpacing,
		ExponentAtPriceOne:   types.ExponentAtPriceOne,
		SwapFee:              swapFee,
	}
	return pool, nil
}

// GetAddress returns the address of the concentrated liquidity pool
func (p Pool) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(p.Address)
	if err != nil {
		panic(fmt.Sprintf("could not bech32 decode address of pool with id: %d", p.GetId()))
	}
	return addr
}

// GetIncentivesAddress returns the address storing incentives of the concentrated liquidity pool.
func (p Pool) GetIncentivesAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(p.IncentivesAddress)
	if err != nil {
		panic(fmt.Sprintf("could not bech32 decode address of pool with id: %d", p.GetId()))
	}
	return addr
}

// GetId returns the id of the concentrated liquidity pool
func (p Pool) GetId() uint64 {
	return p.Id
}

// String returns the json marshalled string of the pool
func (p Pool) String() string {
	out, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// GetSwapFee returns the swap fee of the pool
func (p Pool) GetSwapFee(ctx sdk.Context) sdk.Dec {
	return p.SwapFee
}

// IsActive returns true if the pool is active
func (p Pool) IsActive(ctx sdk.Context) bool {
	return true
}

// SpotPrice returns the spot price of the pool.
// If base asset is the Token0 of the pool, we use the current sqrt price of the pool.
// If not, we calculate the inverse of the current sqrt price of the pool.
func (p Pool) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	// validate base asset is in pool
	if baseAssetDenom != p.Token0 && baseAssetDenom != p.Token1 {
		return sdk.Dec{}, fmt.Errorf("base asset denom (%s) is not in pool with (%s, %s) pair", baseAssetDenom, p.Token0, p.Token1)
	}
	// validate quote asset is in pool
	if quoteAssetDenom != p.Token0 && quoteAssetDenom != p.Token1 {
		return sdk.Dec{}, fmt.Errorf("quote asset denom (%s) is not in pool with (%s, %s) pair", quoteAssetDenom, p.Token0, p.Token1)
	}

	if baseAssetDenom == p.Token0 {
		return p.CurrentSqrtPrice.Power(2), nil
	}
	return sdk.NewDec(1).Quo(p.CurrentSqrtPrice.Power(2)), nil
}

// GetToken0 returns the token0 of the pool
func (p Pool) GetToken0() string {
	return p.Token0
}

// GetToken1 returns the token1 of the pool
func (p Pool) GetToken1() string {
	return p.Token1
}

// GetCurrentSqrtPrice returns the current sqrt price of the pool
func (p Pool) GetCurrentSqrtPrice() sdk.Dec {
	return p.CurrentSqrtPrice
}

// GetCurrentTick returns the current tick of the pool
func (p Pool) GetCurrentTick() sdk.Int {
	return p.CurrentTick
}

// GetTickSpacing returns the current tick spacing parameter of the pool
func (p Pool) GetTickSpacing() uint64 {
	return p.TickSpacing
}

// GetExponentAtPriceOne returns the precision factor at price one of the pool
func (p Pool) GetExponentAtPriceOne() sdk.Int {
	return p.ExponentAtPriceOne
}

// GetLiquidity returns the liquidity of the pool
func (p Pool) GetLiquidity() sdk.Dec {
	return p.CurrentTickLiquidity
}

// GetLastLiquidityUpdate returns the last time there was a change in pool liquidity or active tick.
func (p Pool) GetLastLiquidityUpdate() time.Time {
	return p.LastLiquidityUpdate
}

func (p Pool) GetType() poolmanagertypes.PoolType {
	return poolmanagertypes.Concentrated
}

// UpdateLiquidity updates the liquidity of the pool. Note that this method is mutative.
func (p *Pool) UpdateLiquidity(newLiquidity sdk.Dec) {
	p.CurrentTickLiquidity = p.CurrentTickLiquidity.Add(newLiquidity)
}

// SetCurrentSqrtPrice updates the current sqrt price of the pool when the first position is created.
func (p *Pool) SetCurrentSqrtPrice(newSqrtPrice sdk.Dec) {
	p.CurrentSqrtPrice = newSqrtPrice
}

// SetCurrentTick updates the current tick of the pool when the first position is created.
func (p *Pool) SetCurrentTick(newTick sdk.Int) {
	p.CurrentTick = newTick
}

// SetTickSpacing updates the tick spacing parameter of the pool.
func (p *Pool) SetTickSpacing(tickSpacing uint64) {
	p.TickSpacing = tickSpacing
}

// SetLastLiquidityUpdate updates the pool's LastLiquidityUpdate to newTime.
func (p *Pool) SetLastLiquidityUpdate(newTime time.Time) {
	p.LastLiquidityUpdate = newTime
}

// updateLiquidityIfActivePosition updates the pool's liquidity if the position is active.
// Returns true if updated, false otherwise.
func (p *Pool) UpdateLiquidityIfActivePosition(ctx sdk.Context, lowerTick, upperTick int64, liquidityDelta sdk.Dec) bool {
	if p.isCurrentTickInRange(lowerTick, upperTick) {
		p.CurrentTickLiquidity = p.CurrentTickLiquidity.Add(liquidityDelta)
		return true
	}
	return false
}

// calcActualAmounts calculates and returns actual amounts based on where the current tick is located relative to position's
// lower and upper ticks.
// There are 3 possible cases:
// -The position is active ( lowerTick <= p.CurrentTick < upperTick).
//   - The provided liquidity is distributed in both tokens.
//   - Actual amounts might differ from desired because we recalculate them from liquidity delta and sqrt price.
//
// - Current tick is below the position ( p.CurrentTick < lowerTick).
//   - The provided liquidity is distributed in token0 only.
//
// - Current tick is above the position ( p.CurrentTick >= p.upperTick ).
//   - The provided liquidity is distributed in token1 only.
//
// Note, that liquidityDelta can be positive or negative but cannot be zero.
// If zero, an error is returned.
// If positive, we assume, liquidity being added. As a result, we round up so that
// we request a user to add more liquidity in favor of the pool.
// If negative, we assume, liquidity being removed. As a result, we round down so that
// we request a user to remove less liquidity in favor of the pool.
func (p Pool) CalcActualAmounts(ctx sdk.Context, lowerTick, upperTick int64, liquidityDelta sdk.Dec) (actualAmountDenom0 sdk.Dec, actualAmountDenom1 sdk.Dec, err error) {
	if liquidityDelta.IsZero() {
		return sdk.Dec{}, sdk.Dec{}, types.ErrZeroLiquidity
	}

	if lowerTick >= upperTick {
		return sdk.Dec{}, sdk.Dec{}, types.InvalidLowerUpperTickError{LowerTick: lowerTick, UpperTick: upperTick}
	}

	// Transform the provided ticks into their corresponding sqrtPrices.
	sqrtPriceLowerTick, sqrtPriceUpperTick, err := math.TicksToSqrtPrice(lowerTick, upperTick)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}

	// When liquidity delta is positive, that means that we are adding liquidity.
	// Therefore, we should round up to require user provide a higher amount
	// in favor of the pool.
	// When liquidity delta is negative, that means that we are removing liquidity.
	// Therefore, we should round down to require user provide a lower amount
	// in favor of the pool.
	roundUp := liquidityDelta.IsPositive()

	if p.isCurrentTickInRange(lowerTick, upperTick) {
		// outcome one: the current price falls within the position
		// if this is the case, we attempt to provide liquidity evenly between asset0 and asset1
		// we also update the pool liquidity since the virtual liquidity is modified by this position's creation
		currentSqrtPrice := p.CurrentSqrtPrice
		actualAmountDenom0 = math.CalcAmount0Delta(liquidityDelta, currentSqrtPrice, sqrtPriceUpperTick, roundUp)
		actualAmountDenom1 = math.CalcAmount1Delta(liquidityDelta, currentSqrtPrice, sqrtPriceLowerTick, roundUp)
	} else if p.CurrentTick.LT(sdk.NewInt(lowerTick)) {
		// outcome two: position is below current price
		// this means position is solely made up of asset0
		actualAmountDenom1 = sdk.ZeroDec()
		actualAmountDenom0 = math.CalcAmount0Delta(liquidityDelta, sqrtPriceLowerTick, sqrtPriceUpperTick, roundUp)
	} else {
		// outcome three: position is above current price
		// this means position is solely made up of asset1
		actualAmountDenom0 = sdk.ZeroDec()
		actualAmountDenom1 = math.CalcAmount1Delta(liquidityDelta, sqrtPriceLowerTick, sqrtPriceUpperTick, roundUp)
	}

	return actualAmountDenom0, actualAmountDenom1, nil
}

// isCurrentTickInRange returns true if pool's current tick is within
// the range of the lower and upper ticks. False otherwise.
// TODO: add tests.
func (p Pool) isCurrentTickInRange(lowerTick, upperTick int64) bool {
	return p.CurrentTick.GTE(sdk.NewInt(lowerTick)) && p.CurrentTick.LT(sdk.NewInt(upperTick))
}

// ApplySwap state of pool after swap.
// It specifically overwrites the pool's liquidity, curr tick and the curr sqrt price
func (p *Pool) ApplySwap(newLiquidity sdk.Dec, newCurrentTick sdk.Int, newCurrentSqrtPrice sdk.Dec) error {
	// Check if the new liquidity provided is not negative.
	if newLiquidity.IsNegative() {
		return types.NegativeLiquidityError{Liquidity: newLiquidity}
	}

	// Check if the new sqrt price provided is not negative.
	if newCurrentSqrtPrice.IsNegative() {
		return types.SqrtPriceNegativeError{ProvidedSqrtPrice: newCurrentSqrtPrice}
	}

	// Check if the new tick provided is within boundaries of the pool's precision factor.
	if newCurrentTick.LT(sdk.NewInt(types.MinTick)) || newCurrentTick.GT(sdk.NewInt(types.MaxTick)) {
		return types.TickIndexNotWithinBoundariesError{
			MaxTick:    types.MaxTick,
			MinTick:    types.MinTick,
			ActualTick: newCurrentTick.Int64(),
		}
	}

	p.CurrentTickLiquidity = newLiquidity
	p.CurrentTick = newCurrentTick
	p.CurrentSqrtPrice = newCurrentSqrtPrice

	return nil
}
