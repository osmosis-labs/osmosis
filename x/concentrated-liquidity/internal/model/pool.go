package model

import (
	"encoding/json"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
)

var (
	_ types.ConcentratedPoolExtension = &Pool{}
)

func NewConcentratedLiquidityPool(poolId uint64, denom0, denom1 string, currSqrtPrice sdk.Dec, currTick sdk.Int) (Pool, error) {
	denom0, denom1, err := types.OrderInitialPoolDenoms(denom0, denom1)
	if err != nil {
		return Pool{}, err
	}
	pool := Pool{
		// TODO: move gammtypes.NewPoolAddress(poolId) to swaproutertypes
		Address:          gammtypes.NewPoolAddress(poolId).String(),
		Id:               poolId,
		CurrentSqrtPrice: currSqrtPrice,
		CurrentTick:      currTick,
		Liquidity:        sdk.ZeroDec(),
		Token0:           denom0,
		Token1:           denom1,
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
	return sdk.Dec{}
}

// GetExitFee returns the exit fee of the pool
func (p Pool) GetExitFee(ctx sdk.Context) sdk.Dec {
	return sdk.Dec{}
}

// IsActive returns true if the pool is active
func (p Pool) IsActive(ctx sdk.Context) bool {
	return true
}

// SpotPrice returns the spot price of the pool.
// If base asset is the Token0 of the pool, we use the current sqrt price of the pool.
// If not, we calculate the inverse of the current sqrt price of the pool.
func (p Pool) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	// if zero for one, we use the pool curr sqrt price directly.
	if p.Token0 == baseAssetDenom && p.Token1 == quoteAssetDenom {
		return p.CurrentSqrtPrice.Power(2), nil
	} else if p.Token1 == baseAssetDenom && p.Token0 == quoteAssetDenom { // if not, we calculate the reverse spot price by 1 / currentSqrtPrice^2
		return sdk.NewDec(1).Quo(p.CurrentSqrtPrice.Power(2)), nil
	} else {
		return sdk.Dec{}, fmt.Errorf("base asset denom %s is not in the pool", baseAssetDenom)
	}
}

// GetTotalShares returns the total shares of the pool
func (p Pool) GetTotalShares() sdk.Int {
	return sdk.Int{}
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

// GetLiquidity returns the liquidity of the pool
func (p Pool) GetLiquidity() sdk.Dec {
	return p.Liquidity
}

// UpdateLiquidity updates the liquidity of the pool. Note that this method is mutative.
func (p *Pool) UpdateLiquidity(newLiquidity sdk.Dec) {
	p.Liquidity = p.Liquidity.Add(newLiquidity)
}

// updateLiquidityIfActivePosition updates the pool's liquidity if the position is active.
// Returns true if updated, false otherwise.
// TODO: add tests.
func (p *Pool) UpdateLiquidityIfActivePosition(ctx sdk.Context, lowerTick, upperTick int64, liquidityDelta sdk.Dec) bool {
	if p.isCurrentTickInRange(lowerTick, upperTick) {
		p.Liquidity = p.Liquidity.Add(liquidityDelta)
		return true
	}
	return false
}

// calcActualAmounts calculates and returns actual amounts based on where the current tick is located relative to position's
// lower and upper ticks.
// There are 3 possible cases:
// -The position is active ( lowerTick <= p.CurrentTick < upperTick).
//    * The provided liqudity is distributed in both tokens.
//    * Actual amounts might differ from desired because we recalculate them from liquidity delta and sqrt price.
//      the calculations lead to amounts being off. // TODO: confirm logic is correct
// - Current tick is below the position ( p.CurrentTick < lowerTick).
//    * The provided liquidity is distributed in token0 only.
// - Current tick is above the position ( p.CurrentTick >= p.upperTick ).
//    * The provided liquidity is distributed in token1 only.
// TODO: add tests.
func (p Pool) CalcActualAmounts(ctx sdk.Context, lowerTick, upperTick int64, sqrtRatioLowerTick, sqrtRatioUpperTick sdk.Dec, liquidityDelta sdk.Dec) (actualAmountDenom0 sdk.Dec, actualAmountDenom1 sdk.Dec) {
	if p.isCurrentTickInRange(lowerTick, upperTick) {
		// outcome one: the current price falls within the position
		// if this is the case, we attempt to provide liquidity evenly between asset0 and asset1
		// we also update the pool liquidity since the virtual liquidity is modified by this position's creation
		currentSqrtPrice := p.CurrentSqrtPrice
		actualAmountDenom0 = math.CalcAmount0Delta(liquidityDelta, currentSqrtPrice, sqrtRatioUpperTick, false)
		actualAmountDenom1 = math.CalcAmount1Delta(liquidityDelta, currentSqrtPrice, sqrtRatioLowerTick, false)
	} else if p.CurrentTick.LT(sdk.NewInt(lowerTick)) {
		// outcome two: position is below current price
		// this means position is solely made up of asset0
		actualAmountDenom1 = sdk.ZeroDec()
		actualAmountDenom0 = math.CalcAmount0Delta(liquidityDelta, sqrtRatioLowerTick, sqrtRatioUpperTick, false)
	} else {
		// outcome three: position is above current price
		// this means position is solely made up of asset1
		actualAmountDenom0 = sdk.ZeroDec()
		actualAmountDenom1 = math.CalcAmount1Delta(liquidityDelta, sqrtRatioLowerTick, sqrtRatioUpperTick, false)
	}

	return actualAmountDenom0, actualAmountDenom1
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
	p.Liquidity = newLiquidity
	p.CurrentTick = newCurrentTick
	p.CurrentSqrtPrice = newCurrentSqrtPrice
	return nil
}

// TODO: Figure out what we are going to do with these stubs since we call these in the keeper but are needed to use PoolI
func (p Pool) CalcInAmtGivenOut(
	ctx sdk.Context, tokensOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (p Pool) CalcOutAmtGivenIn(ctx sdk.Context, tokensIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (sdk.Coin, error) {
	return sdk.Coin{}, nil
}

func (p Pool) GetTotalPoolLiquidity(ctx sdk.Context) sdk.Coins {
	return sdk.Coins{}
}

func (p *Pool) SwapInAmtGivenOut(ctx sdk.Context, tokensOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (p *Pool) SwapOutAmtGivenIn(ctx sdk.Context, tokensIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	return sdk.Coin{}, nil
}
