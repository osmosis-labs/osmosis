package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	authenticators "github.com/osmosis-labs/osmosis/v21/x/authenticator/keeper"
	authpost "github.com/osmosis-labs/osmosis/v21/x/authenticator/post"

	protorevkeeper "github.com/osmosis-labs/osmosis/v21/x/protorev/keeper"
)

func NewPostHandler(
	protoRevKeeper *protorevkeeper.Keeper,
	authenticatorKeeper *authenticators.Keeper,
) sdk.PostHandler {
	return sdk.ChainPostDecorators(
		protorevkeeper.NewProtoRevDecorator(*protoRevKeeper),
		authpost.NewAuthenticatorDecorator(authenticatorKeeper),
	)
}
