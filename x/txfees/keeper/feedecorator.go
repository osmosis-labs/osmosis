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

func NewMempoolFeeDecorator(txFeesKeeper Keeper, opts types.MempoolFeeOptions) MempoolFeeDecorator {
	return MempoolFeeDecorator{
		TxFeesKeeper: txFeesKeeper,
		Opts:         opts,
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
	// to enter our mempool.
	// So we ensure that the provided fees meet a minimum threshold for the validator,
	// converting every non-osmo specified asset into an osmo-equivalent amount, to determine sufficiency.
	// If the msg is applicable for sybil resistant fees, add the swap fees paid to the tx fee when considering
	if (ctx.IsCheckTx() || ctx.IsReCheckTx()) && !simulate {
		// create sybil resistant fee structure
		sybil, err := mfd.GetMinBaseGasPriceForTx(ctx, baseDenom, feeTx)
		if err != nil {
			return ctx, err
		}

		// no gas price, go on
		if sybil.GasPrice.IsZero() {
			return next(ctx, tx, simulate)
		}

		// *** Sybil swap fees cannot pay for a tx entirely w/o a tx fee.
		//     	The entire tx.Fee() will be deducted in the DeductFeeDecorator
		//     	but the amount is not compared to the gas. Therefore, a tx fee
		// 	 	can be short on the gas cost and the swap fees can make up the differencee.
		// no fee attached, and non-zero gas price -> reject tx
		if len(feeCoins) != 1 {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "no fee attached with non-zero gas")
		}

		// check if sybil resistant fees are sufficient for gas price
		if err := mfd.TxFeesKeeper.IsSufficientFee(ctx, sybil, feeTx.GetGas(), feeCoins[0]); err != nil {
			return ctx, err
		}
	}
	// gas price can be paid for by tx using resistant swap fees
	return next(ctx, tx, simulate)
}

// IsSufficientFee checks if the feeCoin provided (in any asset), is worth enough osmo at current spot prices
// to pay the gas cost of this tx.
func (k Keeper) IsSufficientFee(ctx sdk.Context, sybil Sybil, gasRequested uint64, feeCoin sdk.Coin) error {
	baseDenom, err := k.GetBaseDenom(ctx)
	if err != nil {
		return err
	}

	// Determine the required fees by multiplying the required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(int64(gasRequested))
	requiredBaseFee := sdk.NewCoin(baseDenom, sybil.GasPrice.Mul(glDec).Ceil().RoundInt())

	// Convert tx fee to base denom
	convertedFee, err := k.ConvertToBaseToken(ctx, feeCoin)
	if err != nil {
		return err
	}

	// Add converted fee from tx to sybil fees paid
	totalFee := sybil.AddToFeesPaid(convertedFee)

	// now including the swap fees paid for in the msg
	if !(totalFee.FeesPaid.IsGTE(requiredBaseFee)) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s which converts to %s. required: %s", feeCoin, convertedFee, requiredBaseFee)
	}

	return nil
}

func (mfd MempoolFeeDecorator) GetMinBaseGasPriceForTx(ctx sdk.Context, baseDenom string, tx sdk.FeeTx) (Sybil, error) {
	// Get min gas price for node
	cfgMinGasPrice := ctx.MinGasPrices().AmountOf(baseDenom)

	// Check if high gas tx
	if tx.GetGas() >= mfd.Opts.HighGasTxThreshold {
		cfgMinGasPrice = sdk.MaxDec(cfgMinGasPrice, mfd.Opts.MinGasPriceForHighGasTx)
	}

	// Check if arbitration tx
	if txfee_filters.IsArbTxLoose(tx) {
		cfgMinGasPrice = sdk.MaxDec(cfgMinGasPrice, mfd.Opts.MinGasPriceForArbitrageTx)
	}

	// Create sybil fee structure
	sybil := NewSybil(cfgMinGasPrice, sdk.NewCoin(baseDenom, sdk.ZeroInt()))

	// Check if message qualifies for sybil resistant fees
	msg, isSybilSwap := tx.GetMsgs()[0].(gammtypes.SybilResistantFeeSwap)
	if !isSybilSwap {
		return sybil, nil
	}

	// Get token for swap fee amounts
	token := msg.GetTokenToFee()

	// Check if token to apply swap fees to is a feetoken
	if !mfd.TxFeesKeeper.feeTokenExists(ctx, token.Denom) {
		return sybil, nil
	}

	// Get fees paid in swap fees
	feesPaid, err := mfd.TxFeesKeeper.getFeesPaid(ctx, msg.PoolIdOnPath(), msg.TokenDenomsOnPath(), token)
	if err != nil {
		return Sybil{}, err
	}

	// Add fees paid to sybil
	return sybil.AddToFeesPaid(feesPaid), nil
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
		return ctx, fmt.Errorf("fee collector module account (%s) has not been set", types.FeeCollectorName)
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
