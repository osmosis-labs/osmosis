package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	appParams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/market/types"

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

func (k Keeper) Swap(
	ctx sdk.Context,
	trader sdk.AccAddress,
	receiver sdk.AccAddress,
	offerCoin sdk.Coin,
	askDenom string,
) (*types.MsgSwapResponse, error) {
	// Compute exchange rates between the ask and offer
	swapDecCoin, spread, err := k.ComputeSwap(ctx, offerCoin, askDenom)
	if err != nil {
		return nil, err
	}

	// Charge a spread if applicable; the spread is burned
	var feeDecCoin sdk.DecCoin
	if spread.IsPositive() {
		feeDecCoin = sdk.NewDecCoinFromDec(swapDecCoin.Denom, spread.Mul(swapDecCoin.Amount))
	} else {
		feeDecCoin = sdk.NewDecCoin(swapDecCoin.Denom, osmomath.ZeroInt())
	}

	// Send offer coins to module account
	offerCoins := sdk.NewCoins(offerCoin)
	err = k.BankKeeper.SendCoinsFromAccountToModule(ctx, trader, types.ModuleName, offerCoins)
	if err != nil {
		return nil, err
	}

	if offerCoin.Denom != appParams.BaseCoinUnit { // stable -> melody or stable -> stable
		// Burn offered coins and subtract from the trader's account
		err = k.BankKeeper.BurnCoins(ctx, types.ModuleName, offerCoins)
		if err != nil {
			return nil, err
		}
	}

	// Subtract fee from the swap coin
	swapDecCoin.Amount = swapDecCoin.Amount.Sub(feeDecCoin.Amount)

	// Mint asked coins and credit Trader's account
	swapCoin, decimalCoin := swapDecCoin.TruncateDecimal()

	// Ensure to fail the swap tx when zero swap coin
	if !swapCoin.IsPositive() {
		return nil, types.ErrZeroSwapCoin
	}

	feeDecCoin = feeDecCoin.Add(decimalCoin) // add truncated decimalCoin to swapFee
	feeCoin, _ := feeDecCoin.TruncateDecimal()

	mintCoins := sdk.NewCoins(swapCoin, feeCoin)

	taxReceiverAddr, err := sdk.AccAddressFromBech32(k.GetParams(ctx).TaxReceiver)
	if err != nil {
		return nil, err
	}

	// mint only stable coin
	if askDenom != appParams.BaseCoinUnit { // melody -> stable or stable -> stable
		err = k.BankKeeper.MintCoins(ctx, types.ModuleName, mintCoins)
		if err != nil {
			return nil, err
		}

		// Send swap coin to the trader
		swapCoins := sdk.NewCoins(swapCoin)
		err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, swapCoins)
		if err != nil {
			return nil, err
		}

		// Send fees to TxReceiver
		err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, taxReceiverAddr, sdk.NewCoins(feeCoin))
		if err != nil {
			return nil, fmt.Errorf("could not send from exchange vault to recipient: %w", err)
		}
	} else { // stable -> melody
		// native coin transfer using exchange vault
		marketVaultBalance := k.GetExchangePoolBalance(ctx)
		if marketVaultBalance.Amount.LT(swapCoin.Amount) {
			return nil, errorsmod.Wrapf(types.ErrNotEnoughBalanceOnMarketVaults, "Market vaults do not have enough coins to swap. Available amount: (main: %v), needed amount: %v",
				marketVaultBalance.Amount, swapCoin.Amount)
		}

		err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, sdk.NewCoins(swapCoin))
		if err != nil {
			return nil, fmt.Errorf("could not send from exchange vault to recipient: %w", err)
		}

		// Send fees to TxReceiver
		err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, taxReceiverAddr, sdk.NewCoins(feeCoin))
		if err != nil {
			return nil, fmt.Errorf("could not send from exchange vault to recipient: %w", err)
		}
	}

	return &types.MsgSwapResponse{
		SwapCoin: swapCoin,
		SwapFee:  feeCoin,
	}, nil
}
