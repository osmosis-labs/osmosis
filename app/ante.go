package app

import (
	"fmt"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
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
	appOpts servertypes.AppOptions,
	ak ante.AccountKeeper, bankKeeper authtypes.BankKeeper,
	txFeesKeeper *txfeeskeeper.Keeper, spotPriceCalculator txfeestypes.SpotPriceCalculator,
	sigGasConsumer ante.SignatureVerificationGasConsumer,
	signModeHandler signing.SignModeHandler,
	channelKeeper channelkeeper.Keeper,
) sdk.AnteHandler {
	mempoolFeeDecorator := txfeeskeeper.NewMempoolFeeDecorator(*txFeesKeeper)
	// Optional if anyone else is using this repo. Primarily of impact for Osmosis.
	// TODO: Abstract this better
	mempoolFeeDecorator.SetArbMinGasFee(parseArbGasFromConfig(appOpts))
	return sdk.ChainAnteDecorators(
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewRejectExtensionOptionsDecorator(),
		NewMempoolMaxGasPerTxDecorator(),
		// Use Mempool Fee Decorator from our txfees module instead of default one from auth
		// https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/middleware/fee.go#L34
		mempoolFeeDecorator,
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

// TODO: Abstract this function better. We likely need a parse `osmosis-mempool` config section.
func parseArbGasFromConfig(appOpts servertypes.AppOptions) sdk.Dec {
	arbMinFeeInterface := appOpts.Get("osmosis-mempool.arbitrage-min-gas-fee")
	arbMinFee := txfeeskeeper.DefaultArbMinGasFee
	if arbMinFeeInterface != nil {
		arbMinFeeStr, ok := arbMinFeeInterface.(string)
		if !ok {
			panic("invalidly configured osmosis-mempool.arbitrage-min-gas-fee")
		}
		var err error
		// pre-pend 0 to allow the config to start with a decimal, e.g. ".01"
		arbMinFee, err = sdk.NewDecFromStr("0" + arbMinFeeStr)
		if err != nil {
			panic(fmt.Errorf("invalidly configured osmosis-mempool.arbitrage-min-gas-fee, err= %v", err))
		}
	}
	return arbMinFee
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
