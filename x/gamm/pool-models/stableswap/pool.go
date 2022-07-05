package stableswap

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/internal/cfmm_common"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

var _ types.PoolI = &Pool{}

func (pa Pool) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(pa.Address)
	if err != nil {
		panic(fmt.Sprintf("could not bech32 decode address of pool with id: %d", pa.GetId()))
	}
	return addr
}

func (pa Pool) String() string {
	out, err := json.Marshal(pa)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func (pa Pool) GetId() uint64 {
	return pa.Id
}

func (pa Pool) GetSwapFee(ctx sdk.Context) sdk.Dec {
	return pa.PoolParams.SwapFee
}

func (pa Pool) GetExitFee(ctx sdk.Context) sdk.Dec {
	return pa.PoolParams.ExitFee
}

func (pa Pool) IsActive(ctx sdk.Context) bool {
	return true
}

// Returns the coins in the pool owned by all LP shareholders
func (pa Pool) GetTotalPoolLiquidity(ctx sdk.Context) sdk.Coins {
	return pa.PoolLiquidity
}

func (pa Pool) GetTotalShares() sdk.Int {
	return pa.TotalShares.Amount
}

func (pa Pool) GetScalingFactors() []uint64 {
	return pa.ScalingFactor
}

// CONTRACT: scaling factors follow the same index with pool liquidity denoms
func (pa Pool) GetScalingFactorByLiquidityIndex(liquidityIndex int) uint64 {
	return pa.ScalingFactor[liquidityIndex]
}

func (pa Pool) NumAssets() int {
	return len(pa.PoolLiquidity)
}

// returns pool liquidity of the provided denoms, in the same order the denoms were provided in
func (pa Pool) getPoolAmts(denoms ...string) ([]sdk.Int, error) {
	result := make([]sdk.Int, len(denoms))
	poolLiquidity := pa.PoolLiquidity
	for i, d := range denoms {
		amt := poolLiquidity.AmountOf(d)
		if amt.IsZero() {
			return []sdk.Int{}, fmt.Errorf("denom %s does not exist in pool", d)
		}
		result[i] = amt
	}
	return result, nil
}

// getScaledPoolAmts returns scaled amount of pool liquidity based on each asset's precisions
func (pa Pool) getScaledPoolAmts(denoms ...string) ([]sdk.Dec, error) {
	result := make([]sdk.Dec, len(denoms))
	poolLiquidity := pa.PoolLiquidity
	liquidityIndexes := pa.getLiquidityIndexMap()

	for i, denom := range denoms {
		liquidityIndex := liquidityIndexes[denom]

		amt := poolLiquidity.AmountOf(denom)
		if amt.IsZero() {
			return []sdk.Dec{}, fmt.Errorf("denom %s does not exist in pool", denom)
		}
		scalingFactor := pa.GetScalingFactorByLiquidityIndex(liquidityIndex)
		result[i] = amt.ToDec().QuoInt64Mut(int64(scalingFactor))
	}
	return result, nil
}

// getDescaledPoolAmts gets descaled amount of given denom and amount
func (pa Pool) getDescaledPoolAmt(denom string, amount sdk.Dec) sdk.Dec {
	liquidityIndexes := pa.getLiquidityIndexMap()
	liquidityIndex := liquidityIndexes[denom]

	scalingFactor := pa.GetScalingFactorByLiquidityIndex(liquidityIndex)
	return amount.MulInt64(int64(scalingFactor))
}

// getLiquidityIndexMap creates a map of denoms to its index in pool liquidity
func (pa Pool) getLiquidityIndexMap() map[string]int {
	poolLiquidity := pa.PoolLiquidity
	liquidityIndexMap := make(map[string]int, poolLiquidity.Len())
	for i, coin := range poolLiquidity {
		liquidityIndexMap[coin.Denom] = i
	}
	return liquidityIndexMap
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
	p.PoolLiquidity = p.PoolLiquidity.Add(tokensIn...)
	p.TotalShares.Amount = p.TotalShares.Amount.Add(newShares)
}

// TODO: These should all get moved to amm.go
func (pa Pool) CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	if tokenIn.Len() != 1 {
		return sdk.Coin{}, errors.New("stableswap CalcOutAmtGivenIn: tokenIn is of wrong length")
	}
	outAmtDec, err := pa.calcOutAmtGivenIn(tokenIn[0], tokenOutDenom, swapFee)
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

func (pa *Pool) SwapOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	tokenOut, err = pa.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee)
	if err != nil {
		return sdk.Coin{}, err
	}

	pa.updatePoolLiquidityForSwap(tokenIn, sdk.NewCoins(tokenOut))

	return tokenOut, nil
}

func (pa Pool) CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	if tokenOut.Len() != 1 {
		return sdk.Coin{}, errors.New("stableswap CalcInAmtGivenOut: tokenOut is of wrong length")
	}
	// TODO: Refactor this later to handle scaling factors
	amt, err := pa.calcInAmtGivenOut(tokenOut[0], tokenInDenom, swapFee)
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

func (pa *Pool) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	tokenIn, err = pa.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee)
	if err != nil {
		return sdk.Coin{}, err
	}

	pa.updatePoolLiquidityForSwap(sdk.NewCoins(tokenIn), tokenOut)

	return tokenIn, nil
}

func (pa Pool) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	reserves, err := pa.getScaledPoolAmts(baseAssetDenom, quoteAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	scaledSpotPrice := spotPrice(reserves[0], reserves[1])
	spotPrice := pa.getDescaledPoolAmt(baseAssetDenom, scaledSpotPrice)

	return spotPrice, nil
}

func (pa Pool) Copy() Pool {
	pa2 := pa
	pa2.PoolLiquidity = sdk.NewCoins(pa.PoolLiquidity...)
	return pa2
}

func (pa *Pool) CalcJoinPoolShares(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	paCopy := pa.Copy()
	return paCopy.joinPoolSharesInternal(ctx, tokensIn, swapFee)
}

func (pa *Pool) JoinPool(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error) {
	numShares, _, err = pa.joinPoolSharesInternal(ctx, tokensIn, swapFee)
	return numShares, err
}

func (pa *Pool) ExitPool(ctx sdk.Context, exitingShares sdk.Int, exitFee sdk.Dec) (exitingCoins sdk.Coins, err error) {
	exitingCoins, err = pa.CalcExitPoolShares(ctx, exitingShares, exitFee)
	if err != nil {
		return sdk.Coins{}, err
	}

	pa.TotalShares.Amount = pa.TotalShares.Amount.Sub(exitingShares)
	pa.updatePoolLiquidityForExit(exitingCoins)

	return exitingCoins, nil
}

func (pa Pool) CalcExitPoolShares(ctx sdk.Context, exitingShares sdk.Int, exitFee sdk.Dec) (exitingCoins sdk.Coins, err error) {
	return cfmm_common.CalcExitPool(ctx, &pa, exitingShares, exitFee)
}

// no-op for stableswap
func (pa *Pool) PokePool(blockTime time.Time) {}
