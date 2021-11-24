package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	channelkeeper "github.com/cosmos/ibc-go/v2/modules/core/04-channel/keeper"
	ibcante "github.com/cosmos/ibc-go/v2/modules/core/ante"
	txfeeskeeper "github.com/osmosis-labs/osmosis/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/x/txfees/types"
)

// Link to default ante handler used by cosmos sdk:
// https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/x/auth/ante/ante.go#L41
func NewAnteHandler(
	ak ante.AccountKeeper, bankKeeper authtypes.BankKeeper,
	txFeesKeeper txfeeskeeper.Keeper, spotPriceCalculator txfeestypes.SpotPriceCalculator,
	sigGasConsumer ante.SignatureVerificationGasConsumer,
	signModeHandler signing.SignModeHandler,
	channelKeeper channelkeeper.Keeper,
) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewRejectExtensionOptionsDecorator(),
		NewMempoolMaxGasPerTxDecorator(),
		// Use Mempool Fee Decorator from our txfees module instead of default one from auth
		// https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/middleware/fee.go#L34
		txfeeskeeper.NewMempoolFeeDecorator(txFeesKeeper),
		ante.NewValidateBasicDecorator(),
		ante.TxTimeoutHeightDecorator{},
		ante.NewValidateMemoDecorator(ak),
		ante.NewConsumeGasForTxSizeDecorator(ak),
		ante.NewDeductFeeDecorator(ak, bankKeeper, nil),
		ante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(ak),
		ante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		ante.NewSigVerificationDecorator(ak, signModeHandler),
		ante.NewIncrementSequenceDecorator(ak),
		ibcante.NewAnteDecorator(channelKeeper),
	)
}

// NewMempoolMaxGasPerTxDecorator will check if the transaction's gas
// is greater than the local validator's max_gas_wanted_per_tx.
// TODO: max_gas_per_tx is hardcoded here, should move to being defined in app.toml.
// If gas_wanted is too high, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If gas is sufficiently low or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MempoolMaxGasPerTxDecorator
type MempoolMaxGasPerTxDecorator struct{}

func NewMempoolMaxGasPerTxDecorator() MempoolMaxGasPerTxDecorator {
	return MempoolMaxGasPerTxDecorator{}
}

func (mgd MempoolMaxGasPerTxDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	gas := feeTx.GetGas()
	// maximum gas wanted per tx set to 25M
	max_gas_wanted_per_tx := uint64(25000000)

	// Ensure that the provided gas is less than the maximum gas per tx.
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() && !simulate {
		if gas > max_gas_wanted_per_tx {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrOutOfGas, "Too much gas wanted: %d, maximum is 25,000,000", gas)
		}
	}

	return next(ctx, tx, simulate)
}
