package app

import (
	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	smartaccountkeeper "github.com/osmosis-labs/osmosis/v28/x/smart-account/keeper"
	smartaccountpost "github.com/osmosis-labs/osmosis/v28/x/smart-account/post"

	protorevkeeper "github.com/osmosis-labs/osmosis/v28/x/protorev/keeper"
)

func NewPostHandler(
	cdc codec.Codec,
	protoRevKeeper *protorevkeeper.Keeper,
	smartAccountKeeper *smartaccountkeeper.Keeper,
	accountKeeper *authkeeper.AccountKeeper,
	sigModeHandler *txsigning.HandlerMap,
) sdk.PostHandler {
	return sdk.ChainPostDecorators(
		protorevkeeper.NewProtoRevDecorator(*protoRevKeeper),
		smartaccountpost.NewAuthenticatorPostDecorator(
			cdc,
			smartAccountKeeper,
			accountKeeper,
			sigModeHandler,
			// Add an empty handler here to enable a circuit breaker pattern
			sdk.ChainPostDecorators(sdk.Terminator{}), //nolint
		),
	)
}
