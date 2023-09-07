package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authenticators "github.com/osmosis-labs/osmosis/v19/x/authenticator/keeper"
	authpost "github.com/osmosis-labs/osmosis/v19/x/authenticator/post"

	protorevkeeper "github.com/osmosis-labs/osmosis/v19/x/protorev/keeper"
)

func NewPostHandler(
	protoRevKeeper *protorevkeeper.Keeper,
	authenticatorKeeper *authenticators.Keeper,
) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		authpost.NewAuthenticatorDecorator(authenticatorKeeper),
		protorevkeeper.NewProtoRevDecorator(*protoRevKeeper),
	)
}
