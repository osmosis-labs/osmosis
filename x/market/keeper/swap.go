package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	appParams "github.com/osmosis-labs/osmosis/v23/app/params"
	"github.com/osmosis-labs/osmosis/v23/x/market/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ComputeSwap returns the amount of asked coins should be returned for a given offerCoin at the effective
// exchange rate registered with the oracle.
// Returns an Error if the swap is recursive, or the coins to be traded are unknown by the oracle, or the amount
// to trade is too small.
func (k Keeper) ComputeSwap(ctx sdk.Context, offerCoin sdk.Coin, askDenom string) (sdk.DecCoin, sdk.Dec, error) {
	// Return invalid recursive swap err
	if offerCoin.Denom == askDenom {
		return sdk.DecCoin{}, sdk.Dec{}, errorsmod.Wrap(types.ErrRecursiveSwap, askDenom)
	}
	// Get swap amount based on the oracle price
	retDecCoin, err := k.ComputeInternalSwap(ctx, sdk.NewDecCoinFromCoin(offerCoin), askDenom)
	if err != nil {
		return sdk.DecCoin{}, sdk.Dec{}, err
	}

	// Symphony => Symphony swap
	// Apply only tobin tax without constant product spread
	// TODO: yurii: revisit stable => stable swaps, apply tobix tax if needed.
	//if offerCoin.Denom != appParams.BaseCoinUnit && askDenom != appParams.BaseCoinUnit {
	//	var tobinTax sdk.Dec
	//	offerTobinTax, err2 := k.OracleKeeper.GetTobinTax(ctx, offerCoin.Denom)
	//	if err2 != nil {
	//		return sdk.DecCoin{}, sdk.Dec{}, err2
	//	}
	//
	//	askTobinTax, err2 := k.OracleKeeper.GetTobinTax(ctx, askDenom)
	//	if err2 != nil {
	//		return sdk.DecCoin{}, sdk.Dec{}, err2
	//	}
	//
	//	// Apply highest tobin tax for the denoms in the swap operation
	//	if askTobinTax.GT(offerTobinTax) {
	//		tobinTax = askTobinTax
	//	} else {
	//		tobinTax = offerTobinTax
	//	}
	//
	//	spread := tobinTax
	//	return retDecCoin, spread, nil
	//}
	return retDecCoin, sdk.ZeroDec(), nil
}

// ComputeInternalSwap returns the amount of asked DecCoin should be returned for a given offerCoin at the effective
// exchange rate registered with the oracle.
// Different from ComputeSwap, ComputeInternalSwap does not charge a spread as its use is system internal.
func (k Keeper) ComputeInternalSwap(ctx sdk.Context, offerCoin sdk.DecCoin, askDenom string) (sdk.DecCoin, error) {
	if offerCoin.Denom == askDenom {
		return offerCoin, nil
	}

	askRate := sdk.NewDec(1)
	offerRate := sdk.NewDec(1)
	if offerCoin.Denom == appParams.BaseCoinUnit { // melody -> stable
		exchangeRatio, err := k.OracleKeeper.GetMelodyExchangeRate(ctx, askDenom)
		if err != nil {
			return sdk.DecCoin{}, errorsmod.Wrap(types.ErrNoEffectivePrice, askDenom)
		}
		offerRate = exchangeRatio
	} else if askDenom == appParams.BaseCoinUnit { // stable -> melody
		exchangeRatio, err := k.OracleKeeper.GetMelodyExchangeRate(ctx, offerCoin.Denom)
		if err != nil {
			return sdk.DecCoin{}, errorsmod.Wrap(types.ErrNoEffectivePrice, offerCoin.Denom)
		}
		askRate = exchangeRatio
	} else { // stable -> stable
		var err error
		askRate, err = k.OracleKeeper.GetMelodyExchangeRate(ctx, offerCoin.Denom)
		if err != nil {
			return sdk.DecCoin{}, errorsmod.Wrap(types.ErrNoEffectivePrice, offerCoin.Denom)
		}

		offerRate, err = k.OracleKeeper.GetMelodyExchangeRate(ctx, askDenom)
		if err != nil {
			return sdk.DecCoin{}, errorsmod.Wrap(types.ErrNoEffectivePrice, askDenom)
		}
	}

	retAmount := offerCoin.Amount.Mul(askRate).Quo(offerRate)
	if retAmount.LTE(sdk.ZeroDec()) {
		return sdk.DecCoin{}, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, offerCoin.String())
	}

	return sdk.NewDecCoinFromDec(askDenom, retAmount), nil
}

// simulateSwap interface for simulate swap
func (k Keeper) simulateSwap(ctx sdk.Context, offerCoin sdk.Coin, askDenom string) (sdk.Coin, error) {
	if askDenom == offerCoin.Denom {
		return sdk.Coin{}, errorsmod.Wrap(types.ErrRecursiveSwap, askDenom)
	}

	if offerCoin.Amount.BigInt().BitLen() > 100 {
		return sdk.Coin{}, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, offerCoin.String())
	}

	swapCoin, spread, err := k.ComputeSwap(ctx, offerCoin, askDenom)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(sdkerrors.ErrPanic, err.Error())
	}

	if spread.IsPositive() {
		swapFeeAmt := spread.Mul(swapCoin.Amount)
		if swapFeeAmt.IsPositive() {
			swapFee := sdk.NewDecCoinFromDec(swapCoin.Denom, swapFeeAmt)
			swapCoin = swapCoin.Sub(swapFee)
		}
	}

	retCoin, _ := swapCoin.TruncateDecimal()
	return retCoin, nil
}
