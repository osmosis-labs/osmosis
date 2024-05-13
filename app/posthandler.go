package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	smartaccountkeeper "github.com/osmosis-labs/osmosis/v25/x/smart-account/keeper"
	smartaccountpost "github.com/osmosis-labs/osmosis/v25/x/smart-account/post"

	protorevkeeper "github.com/osmosis-labs/osmosis/v25/x/protorev/keeper"
)

func NewPostHandler(
	protoRevKeeper *protorevkeeper.Keeper,
	smartAccountKeeper *smartaccountkeeper.Keeper,
	accountKeeper *authkeeper.AccountKeeper,
	sigModeHandler authsigning.SignModeHandler,
) sdk.PostHandler {
	return sdk.ChainPostDecorators(
		protorevkeeper.NewProtoRevDecorator(*protoRevKeeper),
		smartaccountpost.NewAuthenticatorPostDecorator(
			smartAccountKeeper,
			accountKeeper,
			sigModeHandler,
			// Add an empty handler here to enable a circuit breaker pattern
			sdk.ChainPostDecorators(sdk.Terminator{}),
		),
	)
}
