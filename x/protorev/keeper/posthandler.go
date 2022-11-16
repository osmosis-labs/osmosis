package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

type ProtoRevDecorator struct {
	ProtoRevKeeper Keeper
}

func NewProtoRevDecorator(protoRevDecorator Keeper) ProtoRevDecorator {
	return ProtoRevDecorator{
		ProtoRevKeeper: protoRevDecorator,
	}
}

func (protoRevDec ProtoRevDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	swappedPools := ExtractSwappedPools(tx)

	// short-circuit if there are no pools that were touched
	if len(swappedPools) == 0 {
		return next(ctx, tx, simulate)
	}

	return next(ctx, tx, simulate)
}

// ExtractSwappedPools checks if there were any swaps made on pools and if so
// returns the pool ids of each pool that was traded on.
func ExtractSwappedPools(tx sdk.Tx) map[uint64]bool {
	swappedPools := make(map[uint64]bool)

	// Extract only swaps types and the swapped pools from the tx
	for _, msg := range tx.GetMsgs() {
		if swap, ok := msg.(*gammtypes.MsgSwapExactAmountIn); ok {
			for _, route := range swap.Routes {
				swappedPools[route.PoolId] = true
			}
		} else if swap, ok := msg.(*gammtypes.MsgSwapExactAmountOut); ok {
			for _, route := range swap.Routes {
				swappedPools[route.PoolId] = true
			}
		}
	}

	return swappedPools
}
