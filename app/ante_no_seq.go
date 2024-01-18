//go:build excludeIncrement
// +build excludeIncrement

package app

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	ibcante "github.com/cosmos/ibc-go/v7/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	"github.com/skip-mev/block-sdk/block"

	"github.com/cosmos/cosmos-sdk/client"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"

	osmoante "github.com/osmosis-labs/osmosis/v22/ante"
	v9 "github.com/osmosis-labs/osmosis/v22/app/upgrades/v9"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	auctionkeeper "github.com/skip-mev/block-sdk/x/auction/keeper"

	auctionante "github.com/skip-mev/block-sdk/x/auction/ante"

	txfeeskeeper "github.com/osmosis-labs/osmosis/v22/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/v22/x/txfees/types"
)

// BlockSDKAnteHandlerParams are the parameters necessary to configure the block-sdk antehandlers
type BlockSDKAnteHandlerParams struct {
	freeLane      block.Lane
	mevLane       auctionante.MEVLane
	auctionKeeper auctionkeeper.Keeper
	txConfig      client.TxConfig
}

// Link to default ante handler used by cosmos sdk:
// https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/x/auth/ante/ante.go#L41
// N.B. There is a sister file called `ante_no_seq.go` that is used for e2e testing.
// It leaves out the `IncrementSequenceDecorator` which is not needed for e2e testing.
// If you make a change here, make sure to make the same change in `ante.go`.
func NewAnteHandler(
	appOpts servertypes.AppOptions,
	wasmConfig wasmtypes.WasmConfig,
	txCounterStoreKey storetypes.StoreKey,
	ak ante.AccountKeeper,
	bankKeeper txfeestypes.BankKeeper,
	txFeesKeeper *txfeeskeeper.Keeper,
	spotPriceCalculator txfeestypes.SpotPriceCalculator,
	sigGasConsumer ante.SignatureVerificationGasConsumer,
	signModeHandler signing.SignModeHandler,
	channelKeeper *ibckeeper.Keeper,
	blockSDKParams BlockSDKAnteHandlerParams,
) sdk.AnteHandler {
	mempoolFeeOptions := txfeestypes.NewMempoolFeeOptions(appOpts)
	mempoolFeeDecorator := txfeeskeeper.NewMempoolFeeDecorator(*txFeesKeeper, mempoolFeeOptions)
	sendblockOptions := osmoante.NewSendBlockOptions(appOpts)
	sendblockDecorator := osmoante.NewSendBlockDecorator(sendblockOptions)
	deductFeeDecorator := txfeeskeeper.NewDeductFeeDecorator(*txFeesKeeper, ak, bankKeeper, nil)
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
		ante.NewValidateMemoDecorator(ak),
		ante.NewConsumeGasForTxSizeDecorator(ak),
		block.NewIgnoreDecorator(
			deductFeeDecorator,
			blockSDKParams.freeLane,
		),
		ante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(ak),
		ante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		// ante.NewSigVerificationDecorator(ak, signModeHandler) <-- removed this to prevent failures resulting from invalid tx orders in e2e
		ante.NewIncrementSequenceDecorator(ak),
		ibcante.NewRedundantRelayDecorator(channelKeeper),
		// auction module antehandler
		auctionante.NewAuctionDecorator(
			blockSDKParams.auctionKeeper,
			blockSDKParams.txConfig.TxEncoder(),
			blockSDKParams.mevLane,
		),
	)
}
