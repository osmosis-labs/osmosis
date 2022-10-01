package stableswap

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v12/osmomath"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/internal/cfmm_common"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

var _ types.PoolI = &Pool{}

// NewStableswapPool returns a stableswap pool
// Invariants that are assumed to be satisfied and not checked:
// * len(initialLiquidity) = 2
// * FutureGovernor is valid
// * poolID doesn't already exist
func NewStableswapPool(poolId uint64,
	stableswapPoolParams PoolParams, initialLiquidity sdk.Coins,
	scalingFactors []uint64, scalingFactorController string,
	futureGovernor string,
) (Pool, error) {
	if len(scalingFactors) == 0 {
		scalingFactors = make([]uint64, len(initialLiquidity))
		for i := range scalingFactors {
			scalingFactors[i] = 1
		}
	}

	if err := validateScalingFactors(scalingFactors, len(initialLiquidity)); err != nil {
		return Pool{}, err
	}

	pool := Pool{
		Address:                 types.NewPoolAddress(poolId).String(),
		Id:                      poolId,
		PoolParams:              stableswapPoolParams,
		TotalShares:             sdk.NewCoin(types.GetPoolShareDenom(poolId), types.InitPoolSharesSupply),
		PoolLiquidity:           initialLiquidity,
		ScalingFactor:           scalingFactors,
		ScalingFactorController: scalingFactorController,
		FuturePoolGovernor:      futureGovernor,
	}

	return pool, nil
}

func (p Pool) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(p.Address)
	if err != nil {
		panic(fmt.Sprintf("could not bech32 decode address of pool with id: %d", p.GetId()))
	}
	return addr
}

func (p Pool) String() string {
	out, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func (p Pool) GetId() uint64 {
	return p.Id
}

func (p Pool) GetSwapFee(ctx sdk.Context) sdk.Dec {
	return p.PoolParams.SwapFee
}

func (p Pool) GetExitFee(ctx sdk.Context) sdk.Dec {
	return p.PoolParams.ExitFee
}

func (p Pool) IsActive(ctx sdk.Context) bool {
	return true
}

// Returns the coins in the pool owned by all LP shareholders
func (p Pool) GetTotalPoolLiquidity(ctx sdk.Context) sdk.Coins {
	return p.PoolLiquidity
}

func (p Pool) GetTotalShares() sdk.Int {
	return p.TotalShares.Amount
}

func (p Pool) GetScalingFactors() []uint64 {
	return p.ScalingFactor
}

// CONTRACT: scaling factors follow the same index with pool liquidity denoms
func (p Pool) GetScalingFactorByLiquidityIndex(liquidityIndex int) uint64 {
	return p.ScalingFactor[liquidityIndex]
}

func (p Pool) NumAssets() int {
	return len(p.PoolLiquidity)
}

// scaledInput returns scaled input tokens for usage in AMM equations
func (p Pool) scaleCoin(input sdk.Coin, roundingDirection osmomath.RoundingDirection) (osmomath.BigDec, error) {
	liquidityIndexes := p.getLiquidityIndexMap()
	scalingFactor := p.GetScalingFactorByLiquidityIndex(liquidityIndexes[input.Denom])
	scaledAmount, err := osmomath.DivIntByU64ToBigDec(input.Amount, scalingFactor, roundingDirection)
	if err != nil {
		return osmomath.BigDec{}, err
	}
	return scaledAmount, nil
}

// getDescaledPoolAmts gets descaled amount of given denom and amount
// TODO: Review rounding of this in all contexts
func (p Pool) getDescaledPoolAmt(denom string, amount osmomath.BigDec) sdk.Dec {
	liquidityIndexes := p.getLiquidityIndexMap()
	liquidityIndex := liquidityIndexes[denom]

	scalingFactor := p.GetScalingFactorByLiquidityIndex(liquidityIndex)

	return amount.MulInt64(int64(scalingFactor)).SDKDec()
}

// getLiquidityIndexMap creates a map of denoms to its index in pool liquidity
// TODO: Review all uses of this
func (p Pool) getLiquidityIndexMap() map[string]int {
	poolLiquidity := p.PoolLiquidity
	liquidityIndexMap := make(map[string]int, poolLiquidity.Len())
	for i, coin := range poolLiquidity {
		liquidityIndexMap[coin.Denom] = i
	}
	return liquidityIndexMap
}

func (p Pool) scaledSortedPoolReserves(first string, second string, round osmomath.RoundingDirection) ([]osmomath.BigDec, error) {
	reorderedLiquidity, reorderedScalingFactors, err := p.reorderReservesAndScalingFactors(first, second)
	if err != nil {
		return nil, err
	}
	return osmomath.DivCoinAmtsByU64ToBigDec(reorderedLiquidity, reorderedScalingFactors, round)
}

// reorderReservesAndScalingFactors takes the pool liquidity and scaling factors, and reorders them s.t.
// reorderedReserves[0] = p.GetLiquidity().AmountOf(first)
// reorderedScalingFactors[0] = p.ScalingFactors[p.getLiquidityIndexMap()[first]]
// and the same for index 1, and second.
//
// The remainder of the lists includes every remaining (reserve asset, scaling factor) pair,
// in a deterministic order.
//
// Returns an error if the pool does not contain either of first or second.
func (p Pool) reorderReservesAndScalingFactors(first string, second string) ([]sdk.Coin, []uint64, error) {
	coins := p.PoolLiquidity
	scalingFactors := p.ScalingFactor
	reorderedReserves := make([]sdk.Coin, len(coins))
	reorderedScalingFactors := make([]uint64, len(coins))
	curIndex := 2
	for i, coin := range coins {
		if coin.Denom == first {
			reorderedReserves[0] = coin
			reorderedScalingFactors[0] = scalingFactors[i]
		} else if coin.Denom == second {
			reorderedReserves[1] = coin
			reorderedScalingFactors[1] = scalingFactors[i]
		} else {
			// if we hit this case, then oneof first or second is not in pool liquidity
			if curIndex == len(coins) {
				return nil, nil, fmt.Errorf("one of denom (%s, %s) not found in pool liquidity", first, second)
			}
			reorderedReserves[curIndex] = coin
			reorderedScalingFactors[curIndex] = scalingFactors[i]
			curIndex += 1
		}
	}
	return reorderedReserves, reorderedScalingFactors, nil
}

// updatePoolLiquidityForSwap updates the pool liquidity.
// It requires caller to validate that tokensIn and tokensOut only consist of
// denominations in the pool.
// The function sanity checks this, and panics if not the case.
func (p *Pool) updatePoolLiquidityForSwap(tokensIn sdk.Coins, tokensOut sdk.Coins) {
	numTokens := p.PoolLiquidity.Len()
	// update liquidity
	p.PoolLiquidity = p.PoolLiquidity.Add(tokensIn...).Sub(tokensOut)
	// sanity check that no new denoms were added
	if len(p.PoolLiquidity) != numTokens {
		panic("updatePoolLiquidityForSwap changed number of tokens in pool")
	}
}

// updatePoolLiquidityForExit updates the pool liquidity after an exit.
// The function sanity checks that not all tokens of a given denom are removed,
// and panics if thats the case.
func (p *Pool) updatePoolLiquidityForExit(tokensOut sdk.Coins) {
	p.updatePoolLiquidityForSwap(sdk.Coins{}, tokensOut)
}

func (p *Pool) updatePoolForJoin(tokensIn sdk.Coins, newShares sdk.Int) {
	numTokens := p.NumAssets()
	p.PoolLiquidity = p.PoolLiquidity.Add(tokensIn...)
	if len(p.PoolLiquidity) != numTokens {
		panic(fmt.Sprintf("updatePoolForJoin changed number of tokens in pool from %d to %d", numTokens, len(p.PoolLiquidity)))
	}
	p.TotalShares.Amount = p.TotalShares.Amount.Add(newShares)
}

// TODO: These should all get moved to amm.go
func (p Pool) CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	if tokenIn.Len() != 1 {
		return sdk.Coin{}, errors.New("stableswap CalcOutAmtGivenIn: tokenIn is of wrong length")
	}
	outAmtDec, err := p.calcOutAmtGivenIn(tokenIn[0], tokenOutDenom, swapFee)
	if err != nil {
		return sdk.Coin{}, err
	}

	// we ignore the decimal component, as token out amount must round down
	tokenOutAmt := outAmtDec.TruncateInt()
	if !tokenOutAmt.IsPositive() {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount must be positive")
	}
	return sdk.NewCoin(tokenOutDenom, tokenOutAmt), nil
}

func (p *Pool) SwapOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	tokenOut, err = p.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee)
	if err != nil {
		return sdk.Coin{}, err
	}

	p.updatePoolLiquidityForSwap(tokenIn, sdk.NewCoins(tokenOut))

	return tokenOut, nil
}

func (p Pool) CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	if tokenOut.Len() != 1 {
		return sdk.Coin{}, errors.New("stableswap CalcInAmtGivenOut: tokenOut is of wrong length")
	}
	// TODO: Refactor this later to handle scaling factors
	amt, err := p.calcInAmtGivenOut(tokenOut[0], tokenInDenom, swapFee)
	if err != nil {
		return sdk.Coin{}, err
	}

	// We round up tokenInAmt, as this is whats charged for the swap, for the precise amount out.
	// Otherwise, the pool would under-charge by this rounding error.
	tokenInAmt := amt.Ceil().TruncateInt()

	if !tokenInAmt.IsPositive() {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount must be positive")
	}
	return sdk.NewCoin(tokenInDenom, tokenInAmt), nil
}

func (p *Pool) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	tokenIn, err = p.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee)
	if err != nil {
		return sdk.Coin{}, err
	}

	p.updatePoolLiquidityForSwap(sdk.NewCoins(tokenIn), tokenOut)

	return tokenIn, nil
}

func (p Pool) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	return p.spotPrice(baseAssetDenom, quoteAssetDenom)
}

func (p Pool) Copy() Pool {
	p2 := p
	p2.PoolLiquidity = sdk.NewCoins(p.PoolLiquidity...)
	return p2
}

func (p *Pool) CalcJoinPoolShares(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	pCopy := p.Copy()
	return pCopy.joinPoolSharesInternal(ctx, tokensIn, swapFee)
}

// TODO: implement this
func (p *Pool) CalcJoinPoolNoSwapShares(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	return sdk.ZeroInt(), nil, err
}

func (p *Pool) JoinPool(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error) {
	numShares, _, err = p.joinPoolSharesInternal(ctx, tokensIn, swapFee)
	return numShares, err
}

// TODO: implement this
func (p *Pool) JoinPoolNoSwap(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error) {
	return sdk.ZeroInt(), err
}

func (p *Pool) ExitPool(ctx sdk.Context, exitingShares sdk.Int, exitFee sdk.Dec) (exitingCoins sdk.Coins, err error) {
	exitingCoins, err = p.CalcExitPoolCoinsFromShares(ctx, exitingShares, exitFee)
	if err != nil {
		return sdk.Coins{}, err
	}

	p.TotalShares.Amount = p.TotalShares.Amount.Sub(exitingShares)
	p.updatePoolLiquidityForExit(exitingCoins)

	return exitingCoins, nil
}

func (p Pool) CalcExitPoolCoinsFromShares(ctx sdk.Context, exitingShares sdk.Int, exitFee sdk.Dec) (exitingCoins sdk.Coins, err error) {
	return cfmm_common.CalcExitPool(ctx, &p, exitingShares, exitFee)
}

// no-op for stableswap
func (p *Pool) PokePool(blockTime time.Time) {}

// SetStableSwapScalingFactors sets scaling factors for pool to the given amount
// It should only be able to be successfully called by the pool's ScalingFactorGovernor
// TODO: move commented test for this function from x/gamm/keeper/pool_service_test.go once a pool_test.go file has been created for stableswap
func (p *Pool) SetStableSwapScalingFactors(ctx sdk.Context, scalingFactors []uint64, sender string) error {
	if sender != p.ScalingFactorController {
		return types.ErrNotScalingFactorGovernor
	}

	if err := validateScalingFactors(scalingFactors, p.PoolLiquidity.Len()); err != nil {
		return err
	}

	p.ScalingFactor = scalingFactors
	return nil
}

func validateScalingFactorController(scalingFactorController string) error {
	if len(scalingFactorController) == 0 {
		return nil
	}
	_, err := sdk.AccAddressFromBech32(scalingFactorController)
	return err
}

func validateScalingFactors(scalingFactors []uint64, numAssets int) error {
	if len(scalingFactors) != numAssets {
		return types.ErrInvalidStableswapScalingFactors
	}

	for _, scalingFactor := range scalingFactors {
		if scalingFactor == 0 || int64(scalingFactor) <= 0 {
			return types.ErrInvalidStableswapScalingFactors
		}
	}

	return nil
}
