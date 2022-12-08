package swaproutersimulation

import (
	"github.com/osmosis-labs/osmosis/v13/simulation/simtypes"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

func DefaultActions(keeper swaprouter.Keeper, gammKeeper types.GammKeeper) []simtypes.Action {
	// TODO: uncomment this once swap messages are ported from gamm.
	// simKeeper := simulationKeeper{
	// 	keeper:     keeper,
	// 	gammKeeper: gammKeeper,
	// }
	return []simtypes.Action{
		// TODO: uncomment this once swap messages are ported from gamm.
		// simtypes.NewMsgBasedAction("SwapExactAmountIn", simKeeper, RandomSwapExactAmountIn),
		// simtypes.NewMsgBasedAction("SwapExactAmountOut", simKeeper, RandomSwapExactAmountOut),
	}
}
