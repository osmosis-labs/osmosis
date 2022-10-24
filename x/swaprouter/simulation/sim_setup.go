package swaproutersimulation

import (
	"github.com/osmosis-labs/osmosis/v12/simulation/simtypes"
)

func DefaultActions(keeper SimulationKeeper) []simtypes.Action {
	return []simtypes.Action{
		simtypes.NewMsgBasedAction("SwapExactAmountIn", keeper, RandomSwapExactAmountIn),
		simtypes.NewMsgBasedAction("SwapExactAmountOut", keeper, RandomSwapExactAmountOut),
	}
}
