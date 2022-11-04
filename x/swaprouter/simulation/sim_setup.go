package swaproutersimulation

import (
	"github.com/osmosis-labs/osmosis/v12/simulation/simtypes"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

func DefaultActions(keeper swaprouter.Keeper, gammKeeper types.GammKeeper) []simtypes.Action {
	simKeeper := simulationKeeper{
		keeper:     keeper,
		gammKeeper: gammKeeper,
	}
	return []simtypes.Action{
		simtypes.NewMsgBasedAction("SwapExactAmountIn", simKeeper, RandomSwapExactAmountIn),
		simtypes.NewMsgBasedAction("SwapExactAmountOut", simKeeper, RandomSwapExactAmountOut),
		simtypes.NewMsgBasedAction("CreateUniV2Msg", keeper, RandomCreateUniV2Msg).WithFrequency(simtypes.Frequent),
	}
}
