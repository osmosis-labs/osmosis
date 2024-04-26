//go:build excludeIncrement
// +build excludeIncrement

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

	osmoante "github.com/osmosis-labs/osmosis/v24/ante"
	v9 "github.com/osmosis-labs/osmosis/v24/app/upgrades/v9"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	smartaccountante "github.com/osmosis-labs/osmosis/v24/x/smart-account/ante"
	smartaccountkeeper "github.com/osmosis-labs/osmosis/v24/x/smart-account/keeper"

	txfeeskeeper "github.com/osmosis-labs/osmosis/v24/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/v24/x/txfees/types"
)

// Link to default ante handler used by cosmos sdk:
// https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/x/auth/ante/ante.go#L41
// N.B. There is a sister file called `ante_eq.go` that is used for production.
// Opposite to this, it also chains `IncrementSequenceDecorator` which is not needed for e2e testing but is critical for production.
// If you make a change here, make sure to make the same change in `ante.go`.
func NewAnteHandler(
	appOpts servertypes.AppOptions,
	wasmConfig wasmtypes.WasmConfig,
	txCounterStoreKey storetypes.StoreKey,
	accountKeeper ante.AccountKeeper,
	smartAccountKeeper *smartaccountkeeper.Keeper,
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

	// classicSignatureVerificationDecorator is the old flow to enable a circuit breaker
	classicSignatureVerificationDecorator := sdk.ChainAnteDecorators(
		deductFeeDecorator,
		// We use the old pubkey decorator here to ensure that accounts work as expected,
		// in SetPubkeyDecorator we set a pubkey in the account store, for authenticators
		// we avoid this code path completely.
		ante.NewSetPubKeyDecorator(accountKeeper),
		ante.NewValidateSigCountDecorator(accountKeeper),
		ante.NewSigGasConsumeDecorator(accountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(accountKeeper, signModeHandler),
		// ante.NewIncrementSequenceDecorator(accountKeeper),
		ibcante.NewRedundantRelayDecorator(channelKeeper),
	)

	// authenticatorVerificationDecorator is the new authenticator flow that's embedded into the circuit breaker ante
	authenticatorVerificationDecorator := sdk.ChainAnteDecorators(
		smartaccountante.NewEmitPubKeyDecoratorEvents(accountKeeper),
		ante.NewValidateSigCountDecorator(accountKeeper), // we can probably remove this as multisigs are not supported here
		// Both the signature verification, fee deduction, and gas consumption functionality
		// is embedded in the authenticator decorator
		smartaccountante.NewAuthenticatorDecorator(smartAccountKeeper, accountKeeper, signModeHandler, deductFeeDecorator),
	)

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
		smartaccountante.NewCircuitBreakerDecorator(
			smartAccountKeeper,
			authenticatorVerificationDecorator,
			classicSignatureVerificationDecorator,
		),
	)
}
