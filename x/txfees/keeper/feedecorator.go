package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v7/x/txfees/keeper/txfee_filters"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

// MempoolFeeDecorator will check if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config).
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MempoolFeeDecorator.
type MempoolFeeDecorator struct {
	TxFeesKeeper Keeper
	Opts         types.MempoolFeeOptions
}

type Sybil struct {
	GasPrice     sdk.Dec
	SwapFeesPaid sdk.Int
}

func NewMempoolFeeDecorator(txFeesKeeper Keeper, opts types.MempoolFeeOptions) MempoolFeeDecorator {
	return MempoolFeeDecorator{
		TxFeesKeeper: txFeesKeeper,
		Opts:         opts,
	}
}

func NewSybil(gasPrice sdk.Dec, feesPaid sdk.Int) Sybil {
	return Sybil{
		GasPrice:     gasPrice,
		SwapFeesPaid: feesPaid,
	}
}

func (mfd MempoolFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// The SDK currently requires all txs to be FeeTx's in CheckTx, within its mempool fee decorator.
	// See: https://github.com/cosmos/cosmos-sdk/blob/f726a2398a26bdaf71d78dbf56a82621e84fd098/x/auth/middleware/fee.go#L34-L37
	// So this is not a real restriction at the moment.
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// Ensure that the provided gas is less than the maximum gas per tx,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() && !simulate {
		if feeTx.GetGas() > mfd.Opts.MaxGasWantedPerTx {
			msg := "Too much gas wanted: %d, maximum is %d"
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrOutOfGas, msg, feeTx.GetGas(), mfd.Opts.MaxGasWantedPerTx)
		}
	}

	feeCoins := feeTx.GetFee()

	if len(feeCoins) > 1 {
		return ctx, types.ErrTooManyFeeCoins
	}

	baseDenom, err := mfd.TxFeesKeeper.GetBaseDenom(ctx)
	if err != nil {
		return ctx, err
	}

	// If there is a fee attached to the tx, make sure the fee denom is a denom accepted by the chain
	if len(feeCoins) == 1 {
		feeDenom := feeCoins.GetDenomByIndex(0)
		if feeDenom != baseDenom {
			_, err := mfd.TxFeesKeeper.GetFeeToken(ctx, feeDenom)
			if err != nil {
				return ctx, err
			}
		}
	}
	// TODO: Break this up
	// If we are in CheckTx, this function is ran locally to determine if these fees are sufficient
	// to enter our mempool
	// Ensure that the provided fees meet a minimum threshold for the validator.
	// If the tx msg is a swap, the minimum threshold is reduced by the amount of swap fees already paid.
	// IsSufficientFee gets the path used by for a single or multi-hop swap. Then, gets the pool's on that path
	// their sums associated swap fees. These fees * amounIn or amountOut, depending on the type of msg,
	// represent the sybil resistant fees that have already been paid.
	if (ctx.IsCheckTx() || ctx.IsReCheckTx()) && !simulate {
		sybil := mfd.GetMinBaseGasPriceForTx(ctx, baseDenom, feeTx)
		// continue if there is no gas price to pay
		if !(sybil.GasPrice.IsZero()) {
			// get the message from the tx
			swapMsg, isSwapMsg := tx.GetMsgs()[0].(gammtypes.SwapMsgRoute)
			if !isSwapMsg {
				if len(feeCoins) != 1 {
					return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "no fee attached")
				}
				if err := mfd.TxFeesKeeper.IsSufficientFee(ctx, sybil.GasPrice, feeTx.GetGas(), feeCoins[0]); err != nil {
					return ctx, err
				}
				// The Cosmos SDK docs say to reset the gas meter - should be?
				// https://github.com/cosmos/cosmos-sdk/blob/main/docs/basics/gas-fees.md#antehandler
				return next(ctx, tx, simulate)
			}
			// The message is a swap message
			swapOut, isSwapOut := swapMsg.(gammtypes.SwapMsgAmountOut)
			swapIn, isSwapIn := swapMsg.(gammtypes.SwapMsgAmountIn)
			if isSwapOut {
				// Message is swap exact amount out msg
				// Get pool path
				poolIds := swapOut.PoolIdOnPath()
				swapFees := sdk.ZeroDec()
				// Get swap fees from pools
				for i := range poolIds {
					// Get swap fee
					swapFee, err := mfd.TxFeesKeeper.gammKeeper.GetSwapFee(ctx, poolIds[i])
					if err != nil {
						return ctx, err
					}
					// add to existing swap fees
					swapFees = swapFees.Add(swapFee)
				}
				// Get token to pay fee
				token := swapOut.GetExactTokenOut()
				// Convert token to baseDenom if not already in base denom
				if token.Denom != baseDenom {
					token, err = mfd.TxFeesKeeper.ConvertToBaseToken(ctx, token)
					if err != nil {
						return ctx, err
					}
				}
				// Multiply token by swap fee to get sybil fees paid
				sybil.SwapFeesPaid = token.Amount.ToDec().Mul(swapFees).Ceil().RoundInt()
				// Check if the sybil fees are sufficient for the tx
				if err = mfd.TxFeesKeeper.IsSufficientFeeWithSwap(ctx, sybil, feeTx.GetGas(), feeCoins[0]); err != nil {
					return ctx, err
				}
				// Tx gas price can be paid for with swap fees + tx fee
				return next(ctx, tx, simulate)
			} else if isSwapIn {
				// Message is swap exact amount out msg
				poolIds := swapIn.PoolIdOnPath()
				swapFees := sdk.ZeroDec()
				// Get swap fees from pools
				for i := range poolIds {
					// Get swap fee
					swapFee, err := mfd.TxFeesKeeper.gammKeeper.GetSwapFee(ctx, poolIds[i])
					if err != nil {
						return ctx, err
					}
					swapFees = swapFees.Add(swapFee)
				}
				// Get token to pay fee
				token := swapIn.GetExactTokenIn()
				// Convert token to baseDenom if not already in base denom
				if token.Denom != baseDenom {
					token, err = mfd.TxFeesKeeper.ConvertToBaseToken(ctx, token)
					if err != nil {
						return ctx, err
					}
				}
				// Multiply token by swap fee to get sybil fees paid
				sybil.SwapFeesPaid = token.Amount.ToDec().Mul(swapFees).Ceil().RoundInt()
				// Check if the sybil fees are sufficient for the tx
				if err = mfd.TxFeesKeeper.IsSufficientFeeWithSwap(ctx, sybil, feeTx.GetGas(), feeCoins[0]); err != nil {
					return ctx, err
				}
				// Tx gas price can be paid for with swap fees + tx fee
				return next(ctx, tx, simulate)
			} else {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrTxDecode, "swap msg neither in or out")
			}
		}
	}
	return next(ctx, tx, simulate)
}

// IsSufficientFee checks if the feeCoin provided (in any asset), is worth enough osmo at current spot prices
// to pay the gas cost of this tx.
func (k Keeper) IsSufficientFeeWithSwap(ctx sdk.Context, sybil Sybil, gasRequested uint64, feeCoin sdk.Coin) error {
	baseDenom, err := k.GetBaseDenom(ctx)
	if err != nil {
		return err
	}

	// Determine the required fees by multiplying the required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(int64(gasRequested))
	requiredBaseFee := sdk.NewCoin(baseDenom, sybil.GasPrice.Mul(glDec).Ceil().RoundInt())

	convertedFee, err := k.ConvertToBaseToken(ctx, feeCoin)
	if err != nil {
		return err
	}
	// Add sybil fees paid to the converted fee
	convertedFee.Amount = convertedFee.Amount.Add(sybil.SwapFeesPaid)
	// Converted fee now is including the swap fee paid for in the msg
	if !(convertedFee.IsGTE(requiredBaseFee)) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s which converts to %s. required: %s", feeCoin, convertedFee, requiredBaseFee)
	}

	return nil
}

func (k Keeper) IsSufficientFee(ctx sdk.Context, minBaseGasPrice sdk.Dec, gasRequested uint64, feeCoin sdk.Coin) error {
	baseDenom, err := k.GetBaseDenom(ctx)
	if err != nil {
		return err
	}

	// Determine the required fees by multiplying the required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(int64(gasRequested))
	requiredBaseFee := sdk.NewCoin(baseDenom, minBaseGasPrice.Mul(glDec).Ceil().RoundInt())

	convertedFee, err := k.ConvertToBaseToken(ctx, feeCoin)
	if err != nil {
		return err
	}
	if !(convertedFee.IsGTE(requiredBaseFee)) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s which converts to %s. required: %s", feeCoin, convertedFee, requiredBaseFee)
	}

	return nil
}

func (mfd MempoolFeeDecorator) GetMinBaseGasPriceForTx(ctx sdk.Context, baseDenom string, tx sdk.FeeTx) Sybil {
	// Get min gas price
	cfgMinGasPrice := ctx.MinGasPrices().AmountOf(baseDenom)
	if tx.GetGas() >= mfd.Opts.HighGasTxThreshold {
		cfgMinGasPrice = sdk.MaxDec(cfgMinGasPrice, mfd.Opts.MinGasPriceForHighGasTx)
	}
	if txfee_filters.IsArbTxLoose(tx) {
		cfgMinGasPrice = sdk.MaxDec(cfgMinGasPrice, mfd.Opts.MinGasPriceForArbitrageTx)
	}

	sybil := NewSybil(cfgMinGasPrice, sdk.ZeroInt())
	return sybil
}

// DeductFeeDecorator deducts fees from the first signer of the tx.
// If the first signer does not have the funds to pay for the fees, we return an InsufficientFunds error.
// We call next AnteHandler if fees successfully deducted.
//
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
type DeductFeeDecorator struct {
	ak             types.AccountKeeper
	bankKeeper     types.BankKeeper
	feegrantKeeper types.FeegrantKeeper
	txFeesKeeper   Keeper
}

func NewDeductFeeDecorator(tk Keeper, ak types.AccountKeeper, bk types.BankKeeper, fk types.FeegrantKeeper) DeductFeeDecorator {
	return DeductFeeDecorator{
		ak:             ak,
		bankKeeper:     bk,
		feegrantKeeper: fk,
		txFeesKeeper:   tk,
	}
}

func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// checks to make sure the module account has been set to collect fees in base token
	if addr := dfd.ak.GetModuleAddress(types.FeeCollectorName); addr == nil {
		return ctx, fmt.Errorf("Fee collector module account (%s) has not been set", types.FeeCollectorName)
	}

	// checks to make sure a separate module account has been set to collect fees not in base token
	if addrNonNativeFee := dfd.ak.GetModuleAddress(types.NonNativeFeeCollectorName); addrNonNativeFee == nil {
		return ctx, fmt.Errorf("non native fee collector module account (%s) has not been set", types.NonNativeFeeCollectorName)
	}

	// fee can be in any denom (checked for validity later)
	fee := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()

	// set the fee payer as the default address to deduct fees from
	deductFeesFrom := feePayer

	// If a fee granter was set, deduct fee from the fee granter's account.
	if feeGranter != nil {
		if dfd.feegrantKeeper == nil {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants is not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, tx.GetMsgs())
			if err != nil {
				return ctx, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
		}

		// if no errors, change the account that is charged for fees to the fee granter
		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := dfd.ak.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", deductFeesFrom)
	}

	// deducts the fees and transfer them to the module account
	if !feeTx.GetFee().IsZero() {
		err = DeductFees(dfd.txFeesKeeper, dfd.bankKeeper, ctx, deductFeesFromAcc, feeTx.GetFee())
		if err != nil {
			return ctx, err
		}
	}

	ctx.EventManager().EmitEvents(sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
	)})

	return next(ctx, tx, simulate)
}

// DeductFees deducts fees from the given account and transfers them to the set module account.
func DeductFees(txFeesKeeper types.TxFeesKeeper, bankKeeper types.BankKeeper, ctx sdk.Context, acc authtypes.AccountI, fees sdk.Coins) error {
	// Checks the validity of the fee tokens (sorted, have positive amount, valid and unique denomination)
	if !fees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	// pulls base denom from TxFeesKeeper (should be uOSMO)
	baseDenom, err := txFeesKeeper.GetBaseDenom(ctx)
	if err != nil {
		return err
	}

	// checks if input fee is uOSMO (assumes only one fee token exists in the fees array (as per the check in mempoolFeeDecorator))
	if fees[0].Denom == baseDenom {
		// sends to FeeCollectorName module account
		err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fees)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
		}
	} else {
		// sends to NonNativeFeeCollectorName module account
		err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.NonNativeFeeCollectorName, fees)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
		}
	}

	return nil
}
