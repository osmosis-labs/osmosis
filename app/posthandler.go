package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	protorevkeeper "github.com/osmosis-labs/osmosis/v19/x/protorev/keeper"
)

func NewPostHandler(protoRevKeeper *protorevkeeper.Keeper) sdk.AnteHandler {
	protoRevDecorator := protorevkeeper.NewProtoRevDecorator(*protoRevKeeper)
	return sdk.ChainAnteDecorators(protoRevDecorator)
}
