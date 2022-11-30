package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	protorevkeeper "github.com/osmosis-labs/osmosis/v13/x/protorev/keeper"
)

// Link to the default ante handler used by cosmos sdk:
// https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/x/auth/posthandler/post.go#L11
func NewPostHandler(protoRevKeeper *protorevkeeper.Keeper) sdk.AnteHandler {
	protoRevDecorator := protorevkeeper.NewProtoRevDecorator(*protoRevKeeper)
	return sdk.ChainAnteDecorators(protoRevDecorator)
}
