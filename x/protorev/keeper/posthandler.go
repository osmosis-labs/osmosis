package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ProtoRevDecorator struct {
	ProtoRevKeeper Keeper
}

func NewProtoRevDecorator(protoRevDecorator Keeper) ProtoRevDecorator {
	return ProtoRevDecorator{
		ProtoRevKeeper: protoRevDecorator,
	}
}

// This posthandler will first check if there were any swaps in the tx. If so, collect all of the pools, build three
// pool routes for cyclic arbitrage, and then execute the optimal route if it exists.
func (protoRevDec ProtoRevDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	return next(ctx, tx, simulate)
}
