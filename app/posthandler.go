package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	protorevkeeper "github.com/osmosis-labs/osmosis/v21/x/protorev/keeper"
)

func NewPostHandler(protoRevKeeper *protorevkeeper.Keeper) sdk.PostHandler {
	protoRevDecorator := protorevkeeper.NewProtoRevDecorator(*protoRevKeeper)
	return sdk.ChainPostDecorators(protoRevDecorator)
}
