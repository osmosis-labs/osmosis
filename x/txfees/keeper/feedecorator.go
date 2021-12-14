package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/x/txfees/types"
)

// MempoolFeeDecorator will check if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config).
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MempoolFeeDecorator
type MempoolFeeDecorator struct {
	TxFeesKeeper Keeper
}

func NewMempoolFeeDecorator(txFeesKeeper Keeper) MempoolFeeDecorator {
	return MempoolFeeDecorator{
		TxFeesKeeper: txFeesKeeper,
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

	// If we are in CheckTx, this function is ran locally to determine if these fees are sufficient
	// to enter our mempool.
	// So we ensure that the provided fees meet a minimum threshold for the validator,
	// converting every non-osmo specified asset into an osmo-equivalent amount, to determine sufficiency.
	if (ctx.IsCheckTx() || ctx.IsReCheckTx()) && !simulate {
		minGasPrices := ctx.MinGasPrices()
		minBaseGasPrice := minGasPrices.AmountOf(baseDenom)
		if !(minBaseGasPrice.IsZero()) {
			if len(feeCoins) != 1 {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "no fee attached")
			}
			err = mfd.TxFeesKeeper.IsSufficientFee(ctx, minBaseGasPrice, feeTx.GetGas(), feeCoins[0])
			if err != nil {
				return ctx, err
			}
		}
	}

	return next(ctx, tx, simulate)
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
