package stableswap

import (
	"encoding/json"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/internal/cfmm_common"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

var (
	_ poolmanagertypes.PoolI = &Pool{}
	_ types.CFMMPoolI        = &Pool{}
)

// NewStableswapPool returns a stableswap pool
// Invariants that are assumed to be satisfied and not checked:
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

	scalingFactors, err := applyScalingFactorMultiplier(scalingFactors)
	if err != nil {
		return Pool{}, err
	}

	if err = validateScalingFactors(scalingFactors, len(initialLiquidity)); err != nil {
		return Pool{}, err
	}

	if err = validatePoolLiquidity(initialLiquidity, scalingFactors); err != nil {
		return Pool{}, err
	}

	if err = types.ValidateFutureGovernor(futureGovernor); err != nil {
		return Pool{}, err
	}

	pool := Pool{
		Address:                 poolmanagertypes.NewPoolAddress(poolId).String(),
		Id:                      poolId,
		PoolParams:              stableswapPoolParams,
		TotalShares:             sdk.NewCoin(types.GetPoolShareDenom(poolId), types.InitPoolSharesSupply),
		PoolLiquidity:           initialLiquidity,
		ScalingFactors:          scalingFactors,
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

func (p Pool) GetSpreadFactor(ctx sdk.Context) sdk.Dec {
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
	return p.ScalingFactors
}

func (p Pool) GetType() poolmanagertypes.PoolType {
	return poolmanagertypes.Stableswap
}

// CONTRACT: scaling factors follow the same index with pool liquidity denoms
func (p Pool) GetScalingFactorByDenom(denom string) uint64 {
	for i, coin := range p.PoolLiquidity {
		if denom == coin.Denom {
			return p.ScalingFactors[i]
		}
	}

	return 0
}

func (p Pool) NumAssets() int {
	return len(p.PoolLiquidity)
}

// scaleCoin returns the BigDec amount of the
// input token after scaling it by the token's scaling factor
func (p Pool) scaleCoin(input sdk.Coin, roundingDirection osmomath.RoundingDirection) (osmomath.BigDec, error) {
	scalingFactor := p.GetScalingFactorByDenom(input.Denom)
	scaledAmount, err := osmomath.DivIntByU64ToBigDec(input.Amount, scalingFactor, roundingDirection)
	if err != nil {
		return osmomath.BigDec{}, err
	}
	return scaledAmount, nil
}

// getDescaledPoolAmt descales the passed in amount
// by the scaling factor of the passed in denom
func (p Pool) getDescaledPoolAmt(denom string, amount osmomath.BigDec) sdk.Dec {
	scalingFactor := p.GetScalingFactorByDenom(denom)

	return amount.MulInt64(int64(scalingFactor)).SDKDec()
}

// scaledSortedPoolReserves sorts and scales passed in pool reserves such that the denom
// `first` and the denom `second` are ordered first and second,
// respectively. The rest of the ordering is not specified but
// deterministic.
//
// Returns reserve amounts as an array of type BigDec.
func (p Pool) scaledSortedPoolReserves(first string, second string, round osmomath.RoundingDirection) ([]osmomath.BigDec, error) {
	reorderedLiquidity, reorderedScalingFactors, err := p.reorderReservesAndScalingFactors(first, second)
	if err != nil {
		return nil, err
	}

	if err := validateScalingFactors(reorderedScalingFactors, len(reorderedLiquidity)); err != nil {
		return nil, err
	}

	return osmomath.DivCoinAmtsByU64ToBigDec(reorderedLiquidity, reorderedScalingFactors, round)
}

// reorderReservesAndScalingFactors takes the pool liquidity and scaling factors, and reorders them s.t.
// reorderedReserves[0] = p.GetLiquidity().AmountOf(first)
// reorderedScalingFactors[0] = p.ScalingFactors[p.getLiquidityIndexMap()[first]]
// Similarly, reordering happens for second and index 1.
//
// The remainder of the lists includes every remaining (reserve asset, scaling factor) pair,
// in a deterministic but unspecified order.
//
// Returns an error if the pool does not contain either of first or second.
func (p Pool) reorderReservesAndScalingFactors(first string, second string) ([]sdk.Coin, []uint64, error) {
	coins := p.PoolLiquidity
	scalingFactors := p.ScalingFactors
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

// updatePoolLiquidityForExit updates the pool liquidity and total shares after an exit.
// The function sanity checks that not all tokens of a given denom are removed,
// and panics if thats the case.
func (p *Pool) updatePoolLiquidityForExit(tokensOut sdk.Coins, exitingShares sdk.Int) {
	p.updatePoolLiquidityForSwap(sdk.Coins{}, tokensOut)
	p.TotalShares.Amount = p.TotalShares.Amount.Sub(exitingShares)
}

// updatePoolForJoin updates the pool liquidity and total shares after a join.
// The function sanity checks that no new denoms were added to the pool
// and panics if this is the case.
func (p *Pool) updatePoolForJoin(tokensIn sdk.Coins, newShares sdk.Int) {
	numTokens := p.NumAssets()
	p.PoolLiquidity = p.PoolLiquidity.Add(tokensIn...)
	if len(p.PoolLiquidity) != numTokens {
		panic(fmt.Sprintf("updatePoolForJoin changed number of tokens in pool from %d to %d", numTokens, len(p.PoolLiquidity)))
	}
	p.TotalShares.Amount = p.TotalShares.Amount.Add(newShares)
}

// TODO: These should all get moved to amm.go
// CalcOutAmtGivenIn calculates expected output amount given input token
func (p Pool) CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, spreadFactor sdk.Dec) (tokenOut sdk.Coin, err error) {
	if tokenIn.Len() != 1 {
		return sdk.Coin{}, errors.New("stableswap CalcOutAmtGivenIn: tokenIn is of wrong length")
	}
	outAmtDec, err := p.calcOutAmtGivenIn(tokenIn[0], tokenOutDenom, spreadFactor)
	if err != nil {
		return sdk.Coin{}, err
	}

	// we ignore the decimal component, as token out amount must round down
	tokenOutAmt := outAmtDec.TruncateInt()
	if !tokenOutAmt.IsPositive() {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrInvalidMathApprox,
			fmt.Sprintf("token amount must be positive, got %v", tokenOutAmt))
	}
	return sdk.NewCoin(tokenOutDenom, tokenOutAmt), nil
}

// SwapOutAmtGivenIn executes a swap given a desired input amount
func (p *Pool) SwapOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, spreadFactor sdk.Dec) (tokenOut sdk.Coin, err error) {
	if err = validatePoolLiquidity(p.PoolLiquidity.Add(tokenIn...), p.ScalingFactors); err != nil {
		return sdk.Coin{}, err
	}

	tokenOut, err = p.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, spreadFactor)
	if err != nil {
		return sdk.Coin{}, err
	}

	p.updatePoolLiquidityForSwap(tokenIn, sdk.NewCoins(tokenOut))

	return tokenOut, nil
}

// CalcInAmtGivenOut calculates input amount needed to receive given output
func (p Pool) CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, spreadFactor sdk.Dec) (tokenIn sdk.Coin, err error) {
	if tokenOut.Len() != 1 {
		return sdk.Coin{}, errors.New("stableswap CalcInAmtGivenOut: tokenOut is of wrong length")
	}

	amt, err := p.calcInAmtGivenOut(tokenOut[0], tokenInDenom, spreadFactor)
	if err != nil {
		return sdk.Coin{}, err
	}

	// We round up tokenInAmt, as this is whats charged for the swap, for the precise amount out.
	// Otherwise, the pool would under-charge by this rounding error.
	tokenInAmt := amt.Ceil().TruncateInt()

	if !tokenInAmt.IsPositive() {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrInvalidMathApprox, "token amount must be positive")
	}
	return sdk.NewCoin(tokenInDenom, tokenInAmt), nil
}

// SwapInAmtGivenOut executes a swap given a desired output amount
func (p *Pool) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, spreadFactor sdk.Dec) (tokenIn sdk.Coin, err error) {
	tokenIn, err = p.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, spreadFactor)
	if err != nil {
		return sdk.Coin{}, err
	}

	if err = validatePoolLiquidity(p.PoolLiquidity.Add(tokenIn), p.ScalingFactors); err != nil {
		return sdk.Coin{}, err
	}

	p.updatePoolLiquidityForSwap(sdk.NewCoins(tokenIn), tokenOut)

	return tokenIn, nil
}

// SpotPrice calculates the approximate amount of `baseDenom` one would receive for
// an input dx of `quoteDenom` (to simplify calculations, we approximate dx = 1)
func (p Pool) SpotPrice(ctx sdk.Context, quoteAssetDenom string, baseAssetDenom string) (sdk.Dec, error) {
	return p.spotPrice(quoteAssetDenom, baseAssetDenom)
}

func (p Pool) Copy() Pool {
	p2 := p
	p2.PoolLiquidity = sdk.NewCoins(p.PoolLiquidity...)
	return p2
}

func (p *Pool) CalcJoinPoolShares(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	pCopy := p.Copy()
	return pCopy.joinPoolSharesInternal(ctx, tokensIn, spreadFactor)
}

// CalcJoinPoolNoSwapShares calculates the number of shares created to execute an all-asset pool join with the provided amount of `tokensIn`.
// The input tokens must contain the same tokens as in the pool.
//
// Returns the number of shares created, the amount of coins actually joined into the pool as not all may tokens may be joinable.
// If an all-asset join is not possible, returns an error.
func (p Pool) CalcJoinPoolNoSwapShares(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor sdk.Dec) (numShares sdk.Int, tokensJoined sdk.Coins, err error) {
	// ensure that there aren't too many or too few assets in `tokensIn`
	if tokensIn.Len() != p.NumAssets() || !tokensIn.DenomsSubsetOf(p.GetTotalPoolLiquidity(ctx)) {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New("no-swap joins require LP'ing with all assets in pool")
	}

	// execute a no-swap join with as many tokens as possible given a perfect ratio:
	// * numShares is how many shares are perfectly matched.
	// * remainingTokensIn is how many coins we have left to join that have not already been used.
	numShares, remainingTokensIn, err := cfmm_common.MaximalExactRatioJoin(&p, ctx, tokensIn)
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}

	// ensure that no more tokens have been joined than is possible with the given `tokensIn`
	tokensJoined = tokensIn.Sub(remainingTokensIn)
	if tokensJoined.IsAnyGT(tokensIn) {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New("an error has occurred, more coins joined than token In")
	}

	return numShares, tokensJoined, nil
}

func (p *Pool) JoinPool(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor sdk.Dec) (sdk.Int, error) {
	numShares, _, err := p.joinPoolSharesInternal(ctx, tokensIn, spreadFactor)
	return numShares, err
}

func (p *Pool) JoinPoolNoSwap(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor sdk.Dec) (sdk.Int, error) {
	newShares, tokensJoined, err := p.CalcJoinPoolNoSwapShares(ctx, tokensIn, spreadFactor)
	if err != nil {
		return sdk.Int{}, err
	}

	// update pool with the calculated share and liquidity needed to join pool
	p.updatePoolForJoin(tokensJoined, newShares)
	return newShares, nil
}

func (p *Pool) ExitPool(ctx sdk.Context, exitingShares sdk.Int, exitFee sdk.Dec) (exitingCoins sdk.Coins, err error) {
	exitingCoins, err = p.CalcExitPoolCoinsFromShares(ctx, exitingShares, exitFee)
	if err != nil {
		return sdk.Coins{}, err
	}

	postExitLiquidity := p.PoolLiquidity.Sub(exitingCoins)
	if err := validatePoolLiquidity(postExitLiquidity, p.ScalingFactors); err != nil {
		return sdk.Coins{}, err
	}

	p.updatePoolLiquidityForExit(exitingCoins, exitingShares)

	return exitingCoins, nil
}

func (p Pool) CalcExitPoolCoinsFromShares(ctx sdk.Context, exitingShares sdk.Int, exitFee sdk.Dec) (exitingCoins sdk.Coins, err error) {
	return cfmm_common.CalcExitPool(ctx, &p, exitingShares, exitFee)
}

// SetScalingFactors sets scaling factors for pool to the given amount
// It should only be able to be successfully called by the pool's ScalingFactorGovernor
// TODO: move commented test for this function from x/gamm/keeper/pool_service_test.go once a pool_test.go file has been created for stableswap
func (p *Pool) SetScalingFactors(ctx sdk.Context, scalingFactors []uint64, sender string) error {
	if sender != p.ScalingFactorController {
		return types.ErrNotScalingFactorGovernor
	}

	scalingFactors, err := applyScalingFactorMultiplier(scalingFactors)
	if err != nil {
		return err
	}

	if err = validateScalingFactors(scalingFactors, p.PoolLiquidity.Len()); err != nil {
		return err
	}

	if err = validatePoolLiquidity(p.PoolLiquidity, scalingFactors); err != nil {
		return err
	}

	p.ScalingFactors = scalingFactors
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
		return types.ErrInvalidScalingFactorLength
	}

	for _, scalingFactor := range scalingFactors {
		if int64(scalingFactor) <= 0 {
			return types.ErrInvalidScalingFactors
		}
	}

	return nil
}

// assumes liquidity is all pool liquidity, in correct sorted order
func validatePoolLiquidity(liquidity sdk.Coins, scalingFactors []uint64) error {
	liquidityCount := len(liquidity)
	scalingFactorCount := len(scalingFactors)
	if liquidityCount != scalingFactorCount {
		return types.LiquidityAndScalingFactorCountMismatchError{LiquidityCount: liquidityCount, ScalingFactorCount: scalingFactorCount}
	}

	if liquidityCount < types.MinNumOfAssetsInPool {
		return types.ErrTooFewPoolAssets
	} else if liquidityCount > types.MaxNumOfAssetsInPool {
		return types.ErrTooManyPoolAssets
	}

	liquidityCopy := make(sdk.Coins, liquidityCount)
	copy(liquidityCopy, liquidity)
	liquidityCopy.Sort()

	for i, asset := range liquidity {
		if asset != liquidityCopy[i] {
			return types.UnsortedPoolLiqError{ActualLiquidity: liquidity}
		}

		scaledAmount := asset.Amount.Quo(sdk.NewInt(int64(scalingFactors[i])))
		if scaledAmount.GT(types.StableswapMaxScaledAmtPerAsset) {
			return types.ErrHitMaxScaledAssets
		} else if scaledAmount.LT(sdk.NewInt(types.StableswapMinScaledAmtPerAsset)) {
			return types.ErrHitMinScaledAssets
		}
	}

	return nil
}

func applyScalingFactorMultiplier(scalingFactors []uint64) ([]uint64, error) {
	newScalingFactors := make([]uint64, len(scalingFactors))
	for i, scalingFactor := range scalingFactors {
		newScalingFactors[i] = scalingFactor * types.ScalingFactorMultiplier

		if newScalingFactors[i] < scalingFactor {
			return nil, types.ErrInvalidScalingFactors
		}
	}

	return newScalingFactors, nil
}

func (p *Pool) AsSerializablePool() poolmanagertypes.PoolI {
	return p
}
