package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	appParams "github.com/osmosis-labs/osmosis/v26/app/params"
	"github.com/osmosis-labs/osmosis/v26/x/market/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ComputeSwap returns the amount of asked coins should be returned for a given offerCoin at the effective
// exchange rate registered with the oracle.
// Returns an Error if the swap is recursive, or the coins to be traded are unknown by the oracle, or the amount
// to trade is too small.
func (k Keeper) ComputeSwap(ctx sdk.Context, offerCoin sdk.Coin, askDenom string) (sdk.DecCoin, osmomath.Dec, error) {
	// Return invalid recursive swap err
	if offerCoin.Denom == askDenom {
		return sdk.DecCoin{}, osmomath.Dec{}, errorsmod.Wrap(types.ErrRecursiveSwap, askDenom)
	}
	// Get swap amount based on the oracle price
	retDecCoin, err := k.ComputeInternalSwap(ctx, sdk.NewDecCoinFromCoin(offerCoin), askDenom)
	if err != nil {
		return sdk.DecCoin{}, osmomath.Dec{}, err
	}

	var spread osmomath.Dec
	if offerCoin.Denom != appParams.BaseCoinUnit && askDenom != appParams.BaseCoinUnit {
		offerTobinTax, err := k.OracleKeeper.GetTobinTax(ctx, offerCoin.Denom)
		if err != nil {
			return sdk.DecCoin{}, osmomath.Dec{}, err
		}

		askTobinTax, err := k.OracleKeeper.GetTobinTax(ctx, askDenom)
		if err != nil {
			return sdk.DecCoin{}, osmomath.Dec{}, err
		}

		// Apply highest tobin tax for the denoms in the swap operation
		if askTobinTax.GT(offerTobinTax) {
			spread = askTobinTax
		} else {
			spread = offerTobinTax
		}

	} else {
		spread = k.MinStabilitySpread(ctx)
	}

	return retDecCoin, spread, nil
}

// ComputeInternalSwap returns the amount of asked DecCoin should be returned for a given offerCoin at the effective
// exchange rate registered with the oracle.
// Different from ComputeSwap, ComputeInternalSwap does not charge a spread as its use is system internal.
func (k Keeper) ComputeInternalSwap(ctx sdk.Context, offerCoin sdk.DecCoin, askDenom string) (sdk.DecCoin, error) {
	if offerCoin.Denom == askDenom {
		return offerCoin, nil
	}

	askRate := osmomath.NewDec(1)
	offerRate := osmomath.NewDec(1)
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
	if retAmount.LTE(osmomath.ZeroDec()) {
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
