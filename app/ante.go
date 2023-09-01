package app

import (
	wasm "github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	ibcante "github.com/cosmos/ibc-go/v4/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v4/modules/core/keeper"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"

	osmoante "github.com/osmosis-labs/osmosis/v19/ante"
	v9 "github.com/osmosis-labs/osmosis/v19/app/upgrades/v9"

	txfeeskeeper "github.com/osmosis-labs/osmosis/v19/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/v19/x/txfees/types"

	authante "github.com/osmosis-labs/osmosis/v19/x/authenticator/ante"
	authenticators "github.com/osmosis-labs/osmosis/v19/x/authenticator/keeper"
)

// Link to default ante handler used by cosmos sdk:
// https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/x/auth/ante/ante.go#L41
func NewAnteHandler(
	appOpts servertypes.AppOptions,
	wasmConfig wasm.Config,
	txCounterStoreKey sdk.StoreKey,
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
		ante.NewRejectExtensionOptionsDecorator(),
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
		ante.NewSetPubKeyDecorator(accountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(accountKeeper),
		ante.NewSigGasConsumeDecorator(accountKeeper, sigGasConsumer),
		authante.NewAuthenticatorDecorator(accountKeeper, authenticatorKeeper, signModeHandler),
		ante.NewIncrementSequenceDecorator(accountKeeper),
		ibcante.NewAnteDecorator(channelKeeper),
	)
}
