package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// SwapMsg defines a simple interface for getting the token denoms on a swap message route.
type SwapMsgRoute interface {
	TokenInDenom() string
	TokenOutDenom() string
	TokenDenomsOnPath() []string
	GetTokenToFee() sdk.Coin
	GetPoolIdOnPath() []uint64
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

func (msg MsgSwapExactAmountOut) TokenDenomsOnPath() []string {
	if msg.Routes == nil {
		return []string{}
	}

	if len(msg.Routes) == 0 {
		return []string{}
	}

	denoms := make([]string, 0, len(msg.Routes)+1)
	for i := 0; i < len(msg.Routes); i++ {
		denoms = append(denoms, msg.Routes[i].TokenInDenom)
	}
	denoms = append(denoms, msg.TokenOutDenom())
	return denoms
}

func (msg MsgSwapExactAmountOut) GetTokenToFee() sdk.Coin {
	if msg.TokenOut.Amount.LT(sdk.ZeroInt()) {
		return sdk.Coin{}
	}

	if _, ok := sdk.NewIntFromString(msg.TokenOut.Denom); ok == true {
		return sdk.Coin{}
	}
	return msg.TokenOut
}

func (msg MsgSwapExactAmountOut) GetPoolIdOnPath() []uint64 {
	ids := make([]uint64, 0, len(msg.Routes))
	for i := 0; i < len(msg.Routes); i++ {
		ids = append(ids, msg.Routes[i].PoolId)
	}
	return ids
}

func (msg MsgSwapExactAmountIn) TokenInDenom() string {
	return msg.TokenIn.Denom
}

func (msg MsgSwapExactAmountIn) TokenOutDenom() string {
	lastRouteIndex := len(msg.Routes) - 1
	return msg.Routes[lastRouteIndex].GetTokenOutDenom()
}

func (msg MsgSwapExactAmountIn) TokenDenomsOnPath() []string {
	if msg.Routes == nil {
		return []string{}
	}

	if len(msg.Routes) == 0 {
		return []string{}
	}

	denoms := make([]string, 0, len(msg.Routes)+1)
	denoms = append(denoms, msg.TokenInDenom())
	for i := 0; i < len(msg.Routes); i++ {
		denoms = append(denoms, msg.Routes[i].TokenOutDenom)
	}
	return denoms
}

func (msg MsgSwapExactAmountIn) GetTokenToFee() sdk.Coin {
	if msg.TokenIn.Amount.LT(sdk.ZeroInt()) {
		return sdk.Coin{}
	}

	if _, ok := sdk.NewIntFromString(msg.TokenIn.Denom); ok == true {
		return sdk.Coin{}
	}

	return msg.TokenIn
}

func (msg MsgSwapExactAmountIn) GetPoolIdOnPath() []uint64 {
	ids := make([]uint64, 0, len(msg.Routes))
	for i := 0; i < len(msg.Routes); i++ {
		ids = append(ids, msg.Routes[i].PoolId)
	}
	return ids
}
