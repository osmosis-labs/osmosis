package swaproutersimulation

import (
	"github.com/osmosis-labs/osmosis/v12/simulation/simtypes"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter"
)

func DefaultActions(keeper swaprouter.Keeper) []simtypes.Action {
	return []simtypes.Action{
		simtypes.NewMsgBasedAction("SwapExactAmountIn", keeper, RandomSwapExactAmountIn),
		simtypes.NewMsgBasedAction("SwapExactAmountOut", keeper, RandomSwapExactAmountOut),
	}
}
