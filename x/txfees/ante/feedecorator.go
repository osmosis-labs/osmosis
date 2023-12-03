package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v15/x/txfees/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	keeper "github.com/osmosis-labs/osmosis/v15/x/txfees/keeper"
)

// MempoolFeeDecorator will check if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config).
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MempoolFeeDecorator.
type MempoolFeeDecorator struct {
	TxFeesKeeper keeper.Keeper
}

func NewMempoolFeeDecorator(txFeesKeeper keeper.Keeper) MempoolFeeDecorator {
	return MempoolFeeDecorator{
		TxFeesKeeper: txFeesKeeper,
	}
}

func (mfd MempoolFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidGasLimit, "must provide positive gas")
	}

	//TODO: skip on simulate

	if !ctx.IsCheckTx() && !ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	baseDenom, err := mfd.TxFeesKeeper.GetBaseDenom(ctx)
	if err != nil {
		return ctx, err
	}
	minBaseGasPrice := ctx.MinGasPrices().AmountOf(baseDenom)

	// If minBaseGasPrice is zero, then we don't need to check the fee. Continue
	if minBaseGasPrice.IsZero() {
		return next(ctx, tx, simulate)
	}

	feeCoins := feeTx.GetFee()
	if feeCoins.IsZero() {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "no fees provided")
	}
	// You should only be able to pay with one fee token in a single tx
	if len(feeCoins) > 1 {
		return ctx, sdkerrors.Wrapf(types.ErrTooManyFeeCoins,
			"Expected 1 fee denom attached, got %d", len(feeCoins))
	}
	// If there is a fee attached to the tx, make sure the fee denom is a denom accepted by the chain
	feeDenom := feeCoins.GetDenomByIndex(0)
	if feeDenom != baseDenom {
		_, err := mfd.TxFeesKeeper.GetFeeToken(ctx, feeDenom)
		if err != nil {
			return ctx, err
		}
	}

	// The minimum base gas price is in udym, convert the fee denom's worth to udym terms.
	// Then compare if its sufficient for paying the tx fee.
	err = mfd.IsSufficientFee(ctx, minBaseGasPrice, feeTx.GetGas(), feeCoins[0])
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// IsSufficientFee checks if the feeCoin provided (in any asset), is worth enough udym at current spot prices
// to pay the gas cost of this tx.
func (mfd MempoolFeeDecorator) IsSufficientFee(ctx sdk.Context, minBaseGasPrice sdk.Dec, gasRequested uint64, feeCoin sdk.Coin) error {
	baseDenom, err := mfd.TxFeesKeeper.GetBaseDenom(ctx)
	if err != nil {
		return err
	}

	// Determine the required fees by multiplying the required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(int64(gasRequested))
	requiredBaseFee := sdk.NewCoin(baseDenom, minBaseGasPrice.Mul(glDec).Ceil().RoundInt())

	convertedFee, err := mfd.TxFeesKeeper.ConvertToBaseToken(ctx, feeCoin)
	if err != nil {
		return err
	}
	// check to ensure that the convertedFee should always be greater than or equal to the requireBaseFee
	if !(convertedFee.IsGTE(requiredBaseFee)) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s which converts to %s. required: %s", feeCoin, convertedFee, requiredBaseFee)
	}

	return nil
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
	txFeesKeeper   keeper.Keeper
}

func NewDeductFeeDecorator(tk keeper.Keeper, ak types.AccountKeeper, bk types.BankKeeper, fk types.FeegrantKeeper) DeductFeeDecorator {
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
	if addr := dfd.ak.GetModuleAddress(types.ModuleName); addr == nil {
		return ctx, fmt.Errorf("txfees module account (%s) has not been set", types.ModuleName)
	}

	// checks to make sure the module account has been set to collect fees in base token
	if addr := dfd.ak.GetModuleAddress(types.FeeCollectorName); addr == nil {
		return ctx, fmt.Errorf("fee collector module account (%s) has not been set", types.FeeCollectorName)
	}

	// fee can be in any denom (checked for validity later)
	fee := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	// set the fee payer as the default address to deduct fees from
	deductFeesFrom := feePayer

	if len(fee) > 1 {
		return ctx, sdkerrors.Wrapf(types.ErrTooManyFeeCoins,
			"Expected 1 fee denom attached, got %d", len(fee))
	}

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

	// pulls base denom from TxFeesKeeper (should be udym)
	baseDenom, err := txFeesKeeper.GetBaseDenom(ctx)
	if err != nil {
		return err
	}

	// fees in base denom, sent to the fee collector for distribution
	if fees[0].Denom == baseDenom {
		// sends to FeeCollectorName module account
		err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fees)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
		}
	} else {
		// sends to the txfees module to be swapped and burned
		err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.ModuleName, fees)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
		}
	}

	return nil
}
