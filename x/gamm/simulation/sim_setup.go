package gammsimulation

import (
	"github.com/osmosis-labs/osmosis/v16/simulation/simtypes"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/keeper"
)

func DefaultActions(keeper keeper.Keeper) []simtypes.Action {
	return []simtypes.Action{
		simtypes.NewMsgBasedAction("MsgJoinPool", keeper, RandomJoinPoolMsg).WithFrequency(simtypes.Frequent),
		simtypes.NewMsgBasedAction("MsgExitPool", keeper, RandomExitPoolMsg),
		simtypes.NewMsgBasedAction("CreateUniV2Msg", keeper, RandomCreateUniV2Msg).WithFrequency(simtypes.Frequent),
		simtypes.NewMsgBasedAction("SwapExactAmountIn", keeper, RandomSwapExactAmountIn),
		simtypes.NewMsgBasedAction("SwapExactAmountOut", keeper, RandomSwapExactAmountOut),
		simtypes.NewMsgBasedAction("JoinSwapExternAmountIn", keeper, RandomJoinSwapExternAmountIn),
		simtypes.NewMsgBasedAction("JoinSwapShareAmountOut", keeper, RandomJoinSwapShareAmountOut),
		simtypes.NewMsgBasedAction("ExitSwapExternAmountOut", keeper, RandomExitSwapExternAmountOut),
		simtypes.NewMsgBasedAction("ExitSwapShareAmountIn", keeper, RandomExitSwapShareAmountIn),
	}
}
