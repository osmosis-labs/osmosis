package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	authenticators "github.com/osmosis-labs/osmosis/v21/x/authenticator/keeper"
	authpost "github.com/osmosis-labs/osmosis/v21/x/authenticator/post"

	protorevkeeper "github.com/osmosis-labs/osmosis/v21/x/protorev/keeper"
)

func NewPostHandler(
	protoRevKeeper *protorevkeeper.Keeper,
	authenticatorKeeper *authenticators.Keeper,
	accountKeeper *authkeeper.AccountKeeper,
	sigModeHandler authsigning.SignModeHandler,

) sdk.PostHandler {
	return sdk.ChainPostDecorators(
		protorevkeeper.NewProtoRevDecorator(*protoRevKeeper),
		authpost.NewAuthenticatorDecorator(authenticatorKeeper, accountKeeper, sigModeHandler),
	)
}
