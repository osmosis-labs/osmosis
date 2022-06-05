package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// SwapMsg defines a simple interface for getting the token denoms on a swap message route and the poolId.
type SwapMsgRoute interface {
	TokenInDenom() string
	TokenOutDenom() string
	TokenDenomsOnPath() ([]string, []uint64)
}

type SwapExactIn interface {
	SwapMsgRoute

	GetTokenAmountOut() sdk.Int
}

type SwapExactOut interface {
	SwapMsgRoute

	GetTokenAmountIn() sdk.Int
}

var (
	_ SwapMsgRoute = MsgSwapExactAmountOut{}
	_ SwapMsgRoute = MsgSwapExactAmountIn{}
)

func (msg MsgSwapExactAmountOut) TokenInDenom() string {
	return msg.Routes[0].GetTokenInDenom()
}

func (msg MsgSwapExactAmountOut) TokenOutDenom() string {
	return msg.TokenOut.Denom
}

func (msg MsgSwapExactAmountOut) TokenDenomsOnPath() ([]string, []uint64) {
	denoms := make([]string, 0, len(msg.Routes)+1)
	poolId := make([]uint64, 0, len(msg.Routes)+1)
	for i := 0; i < len(msg.Routes); i++ {
		denoms = append(denoms, msg.Routes[i].TokenInDenom)
		poolId = append(poolId, msg.Routes[i].PoolId)
	}
	denoms = append(denoms, msg.TokenOutDenom())
	return denoms, poolId
}

func (msg MsgSwapExactAmountOut) GetTokenAmountIn() sdk.Int {
	return msg.TokenInMaxAmount
}

func (msg MsgSwapExactAmountIn) TokenInDenom() string {
	return msg.TokenIn.Denom
}

func (msg MsgSwapExactAmountIn) TokenOutDenom() string {
	lastRouteIndex := len(msg.Routes) - 1
	return msg.Routes[lastRouteIndex].GetTokenOutDenom()
}

func (msg MsgSwapExactAmountIn) TokenDenomsOnPath() ([]string, []uint64) {
	denoms := make([]string, 0, len(msg.Routes)+1)
	denoms = append(denoms, msg.TokenInDenom())

	poolIds := make([]uint64, 0, len(msg.Routes)+1)

	for i := 0; i < len(msg.Routes); i++ {
		denoms = append(denoms, msg.Routes[i].TokenOutDenom)
		poolIds = append(poolIds, msg.Routes[i].PoolId)
	}
	return denoms, poolIds
}

func (msg MsgSwapExactAmountIn) GetTokenAmountOut() sdk.Int {
	return msg.TokenOutMinAmount
}
