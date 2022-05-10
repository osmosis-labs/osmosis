package stableswap

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

var _ types.PoolI = &Pool{}

const (
	errMsgFmtTooLittlePoolAssetsGiven = "too little pool assets given: %d, need at least 2"
	errMsgFmtNonExistentDenomGiven    = "can't find the PoolAsset (%s)"
	errMsgFmtDuplicateDenomFound      = "duplicate denom (%s) found"
	errMsfFmtDrainedPool              = "pool balance must stay above 0, updating liqudity for token %s would drain the pool to %d"
	errMsgEmptyDenomGiven             = "denom name cannot be empty"
)

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
	coins := sdk.Coins{}
	for _, asset := range pa.PoolAssets {
		coins = coins.Add(asset.Token)
	}
	return coins
}

func (pa Pool) GetTotalShares() sdk.Int {
	return pa.TotalShares.Amount
}

// getScaledPoolAmt returns scaled amount of pool liquidity based on the asset's precisions
func (pa Pool) getScaledPoolAmt(denom string) (sdk.Int, error) {
	_, asset, err := pa.getPoolAssetAndIndex(denom)
	if err != nil {
		return sdk.Int{}, err
	}
	return asset.Token.Amount.Quo(asset.ScalingFactor), nil
}

// getDescaledPoolAmts gets descaled amount of given denom and amount
func (pa Pool) getDescaledPoolAmt(denom string, amtToDeScale sdk.Dec) (sdk.Dec, error) {
	_, asset, err := pa.getPoolAssetAndIndex(denom)
	if err != nil {
		return sdk.Dec{}, err
	}
	return amtToDeScale.MulInt(asset.ScalingFactor), nil
}

// updatePoolLiquidityForSwap updates the pool liquidity.
// It requires caller to validate that tokensIn and tokensOut only consist of
// denominations in the pool.
// The function sanity checks this, and returns an error if that is not the case.
// Additionally, tokensIn and tokensOut must be sorted by denom. If not, function panics.
func (p *Pool) updatePoolLiquidityForSwap(tokensIn sdk.Coins, tokensOut sdk.Coins) error {
	for _, tokenIn := range tokensIn {
		index, pa, err := p.getPoolAssetAndIndex(tokenIn.Denom)
		if err != nil {
			return err
		}
		p.PoolAssets[index].Token = pa.Token.Add(tokenIn)
	}

	for _, tokenOut := range tokensOut {
		index, pa, err := p.getPoolAssetAndIndex(tokenOut.Denom)
		if err != nil {
			return err
		}
		subtracted := pa.Token.Sub(tokenOut)
		if !subtracted.IsPositive() {
			return fmt.Errorf(errMsfFmtDrainedPool, pa.Token.Denom, subtracted.Amount.Int64())
		}
		p.PoolAssets[index].Token = subtracted
	}
	return nil
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

	if err := pa.updatePoolLiquidityForSwap(tokenIn, sdk.NewCoins(tokenOut)); err != nil {
		panic(err)
	}

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

	if err := pa.updatePoolLiquidityForSwap(sdk.NewCoins(tokenIn), tokenOut); err != nil {
		panic(err)
	}

	return tokenIn, nil
}

func (pa Pool) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	baseAssetScaledInt, err := pa.getScaledPoolAmt(baseAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	quoteAssetScaledInt, err := pa.getScaledPoolAmt(quoteAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	scaledSpotPrice := spotPrice(baseAssetScaledInt.ToDec(), quoteAssetScaledInt.ToDec())
	spotPrice, err := pa.getDescaledPoolAmt(baseAssetDenom, scaledSpotPrice)
	if err != nil {
		return sdk.Dec{}, err
	}

	return spotPrice, nil
}

func (pa Pool) CalcJoinPoolShares(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	return sdk.Int{}, sdk.Coins{}, types.ErrNotImplemented
}

func (pa *Pool) JoinPool(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error) {
	return sdk.Int{}, types.ErrNotImplemented
}

func (pa *Pool) ExitPool(ctx sdk.Context, numShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error) {
	return sdk.Coins{}, types.ErrNotImplemented
}

func (pa Pool) CalcExitPoolShares(ctx sdk.Context, numShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error) {
	return sdk.Coins{}, types.ErrNotImplemented
}

// no-op for stableswap
func (pa *Pool) PokePool(blockTime time.Time) {}

// Returns a pool asset, and its index. If err != nil, then the index will be valid.
// CONTRACT: pool must be created with NewStableSwapPool so that the pool assets are sorted
// by denom. Otherwise, the behavior is undefined.
func (pa Pool) getPoolAssetAndIndex(denom string) (int, PoolAsset, error) {
	if denom == "" {
		return -1, PoolAsset{}, errors.New(errMsgEmptyDenomGiven)
	}

	if len(pa.PoolAssets) == 0 {
		return -1, PoolAsset{}, fmt.Errorf(errMsgFmtNonExistentDenomGiven, denom)
	}

	i := sort.Search(len(pa.PoolAssets), func(i int) bool {
		PoolAssetA := pa.PoolAssets[i]

		compare := strings.Compare(PoolAssetA.Token.Denom, denom)
		return compare >= 0
	})

	if i < 0 || i >= len(pa.PoolAssets) {
		return -1, PoolAsset{}, fmt.Errorf(errMsgFmtNonExistentDenomGiven, denom)
	}

	if pa.PoolAssets[i].Token.Denom != denom {
		return -1, PoolAsset{}, fmt.Errorf(errMsgFmtNonExistentDenomGiven, denom)
	}

	return i, pa.PoolAssets[i], nil
}

// validateAndSortInitialPoolAssets validates and sorts the PoolAssets in the pool.
// It is only designed to be called at the pool's creation.
// If the same denom's PoolAsset exists, it will return error.
// It sorts the list of PoolAssets by denom. This is done to enable fast searching for a PoolAsset by denomination.
func (pa *Pool) validateAndSortInitialPoolAssets() error {
	if len(pa.PoolAssets) < 2 {
		return fmt.Errorf(errMsgFmtTooLittlePoolAssetsGiven, len(pa.PoolAssets))
	}

	if err := validatePoolAssetsAgainstDuplicates(pa.PoolAssets); err != nil {
		return err
	}

	SortPoolAssetsByDenom(pa.PoolAssets)
	return nil
}
