package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	defaultTakerFeeDenom = "udym"
)

// chargeTakerFee extracts the taker fee from the given tokenIn and sends it to the appropriate
// module account. It returns the tokenIn after the taker fee has been extracted.
func (k Keeper) chargeTakerFee(ctx sdk.Context, takerFeeCoin sdk.Coin, sender sdk.AccAddress) error {
	// We determine the distributution of the taker fee based on its denom
	// If the denom is the base denom:
	if takerFeeCoin.Denom == defaultTakerFeeDenom {
		//FIXME: BURN!
		return nil
	} else {
		return k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(takerFeeCoin), sender)
	}
}

// Returns remaining amount in to swap, and takerFeeCoins.
// returns (1 - takerFee) * tokenIn, takerFee * tokenIn
func (k Keeper) calcTakerFeeExactIn(tokenIn sdk.Coin, takerFee sdk.Dec) (sdk.Coin, sdk.Coin) {
	amountInAfterSubTakerFee := sdk.NewDecFromInt(tokenIn.Amount).MulTruncate(sdk.OneDec().Sub(takerFee))
	tokenInAfterSubTakerFee := sdk.NewCoin(tokenIn.Denom, amountInAfterSubTakerFee.TruncateInt())
	takerFeeCoin := sdk.NewCoin(tokenIn.Denom, tokenIn.Amount.Sub(tokenInAfterSubTakerFee.Amount))

	return tokenInAfterSubTakerFee, takerFeeCoin
}

// here we need the output to be (tokenIn / (1 - takerFee), takerFee * tokenIn)
func (k Keeper) calcTakerFeeExactOut(tokenIn sdk.Coin, takerFee sdk.Dec) (sdk.Coin, sdk.Coin) {
	amountInAfterAddTakerFee := sdk.NewDecFromInt(tokenIn.Amount).Quo(sdk.OneDec().Sub(takerFee))
	tokenInAfterAddTakerFee := sdk.NewCoin(tokenIn.Denom, amountInAfterAddTakerFee.Ceil().TruncateInt())
	takerFeeCoin := sdk.NewCoin(tokenIn.Denom, tokenInAfterAddTakerFee.Amount.Sub(tokenIn.Amount))
	return tokenInAfterAddTakerFee, takerFeeCoin
}
