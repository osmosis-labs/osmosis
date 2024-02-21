//go:build !excludeIncrement
// +build !excludeIncrement

package app

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	ibcante "github.com/cosmos/ibc-go/v7/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"

	osmoante "github.com/osmosis-labs/osmosis/v23/ante"
	v9 "github.com/osmosis-labs/osmosis/v23/app/upgrades/v9"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	authante "github.com/osmosis-labs/osmosis/v23/x/authenticator/ante"
	authenticators "github.com/osmosis-labs/osmosis/v23/x/authenticator/keeper"

	txfeeskeeper "github.com/osmosis-labs/osmosis/v23/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/v23/x/txfees/types"
)

// Link to default ante handler used by cosmos sdk:
// https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/x/auth/ante/ante.go#L41
// N.B. There is a sister file called `ante_no_seq.go` that is used for e2e testing.
// It leaves out the `IncrementSequenceDecorator` which is not needed for e2e testing.
// If you make a change here, make sure to make the same change in `ante_no_seq.go`.
func NewAnteHandler(
	appOpts servertypes.AppOptions,
	wasmConfig wasmtypes.WasmConfig,
	txCounterStoreKey storetypes.StoreKey,
	accountKeeper ante.AccountKeeper,
	authenticatorKeeper *authenticators.Keeper,
	bankKeeper txfeestypes.BankKeeper,
	txFeesKeeper *txfeeskeeper.Keeper,
	spotPriceCalculator txfeestypes.SpotPriceCalculator,
	sigGasConsumer ante.SignatureVerificationGasConsumer,
	signModeHandler signing.SignModeHandler,
	channelKeeper *ibckeeper.Keeper,
) sdk.AnteHandler {
	mempoolFeeOptions := txfeestypes.NewMempoolFeeOptions(appOpts)
	mempoolFeeDecorator := txfeeskeeper.NewMempoolFeeDecorator(*txFeesKeeper, mempoolFeeOptions)
	sendblockOptions := osmoante.NewSendBlockOptions(appOpts)
	sendblockDecorator := osmoante.NewSendBlockDecorator(sendblockOptions)
	deductFeeDecorator := txfeeskeeper.NewDeductFeeDecorator(*txFeesKeeper, accountKeeper, bankKeeper, nil)
	return sdk.ChainAnteDecorators(
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		wasmkeeper.NewLimitSimulationGasDecorator(wasmConfig.SimulationGasLimit),
		wasmkeeper.NewCountTXDecorator(txCounterStoreKey),
		ante.NewExtensionOptionsDecorator(nil),
		v9.MsgFilterDecorator{},
		// Use Mempool Fee Decorator from our txfees module instead of default one from auth
		// https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/middleware/fee.go#L34
		mempoolFeeDecorator,
		sendblockDecorator,
		ante.NewValidateBasicDecorator(),
		ante.TxTimeoutHeightDecorator{},
		ante.NewValidateMemoDecorator(accountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(accountKeeper),
		deductFeeDecorator,
		authante.NewSetPubKeyDecorator(accountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(accountKeeper),
		// Our authenticator decorator
		authante.NewAuthenticatorDecorator(authenticatorKeeper, accountKeeper, signModeHandler,
			// Both the signature verification and gas consumption handlers functionality
			// is enbedded in the authenticator decorator, we add the both here to enable a
			// circuit breaker pattern.
			sdk.ChainAnteDecorators(
				ante.NewSigGasConsumeDecorator(accountKeeper, sigGasConsumer),
				ante.NewSigVerificationDecorator(accountKeeper, signModeHandler),
				ante.NewIncrementSequenceDecorator(accountKeeper),
				ibcante.NewRedundantRelayDecorator(channelKeeper),
			),
		),
		ante.NewIncrementSequenceDecorator(accountKeeper),
		ibcante.NewRedundantRelayDecorator(channelKeeper),
	)
}
