package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewConcentratedPool(poolId uint64, firstDenom string, secondDenom string, currentSqrtP sdk.Int, currentTick sdk.Int) Pool {
	denom0 := firstDenom
	denom1 := secondDenom

	// we store token in lexiographical order
	if denom0 < denom1 {
		denom0, denom1 = secondDenom, firstDenom
	}

	pool := Pool{
		Id:               poolId,
		CurrentSqrtPrice: currentSqrtP,
		Token0:           denom0,
		Token1:           denom1,
		CurrentTick:      currentTick,
	}
	return pool
}
