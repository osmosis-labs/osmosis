package types

// SwapMsg defines a simple interface for getting the token denoms on a swap message route.
type SwapMsgRoute interface {
	TokenInDenom() string
	TokenOutDenom() string
	TokenDenomsOnPath() []string
}

type MultiSwapMsgRoute interface {
	GetSwapMsgs() []SwapMsgRoute
}

func (msg MsgSplitRouteSwapExactAmountIn) GetSwapMsgs() []SwapMsgRoute {
	routes := make([]SwapMsgRoute, len(msg.Routes))
	for i := 0; i < len(msg.Routes); i++ {
		routes[i] = SwapAmountInSplitRouteWrapper{msg.Routes[i].Pools, msg.TokenInDenom}
	}
	return routes
}

func (msg MsgSplitRouteSwapExactAmountOut) GetSwapMsgs() []SwapMsgRoute {
	routes := make([]SwapMsgRoute, len(msg.Routes))
	for i := 0; i < len(msg.Routes); i++ {
		routes[i] = SwapAmountOutSplitRouteWrapper{msg.Routes[i].Pools, msg.TokenOutDenom}
	}
	return routes
}

type SwapAmountInSplitRouteWrapper struct {
	Pools   []SwapAmountInRoute `json:"pools"`
	InDenom string              `json:"in_denom"`
}

type SwapAmountOutSplitRouteWrapper struct {
	Pools    []SwapAmountOutRoute `json:"pools"`
	OutDenom string               `json:"in_denom"`
}

var (
	_ SwapMsgRoute = MsgSwapExactAmountIn{}
	_ SwapMsgRoute = MsgSwapExactAmountOut{}
	_ SwapMsgRoute = SwapAmountInSplitRouteWrapper{}
	_ SwapMsgRoute = SwapAmountOutSplitRouteWrapper{}
)

func (msg SwapAmountOutSplitRouteWrapper) TokenInDenom() string {
	return msg.Pools[0].TokenInDenom
}
func (msg SwapAmountOutSplitRouteWrapper) TokenOutDenom() string {
	return msg.OutDenom
}
func (msg SwapAmountOutSplitRouteWrapper) TokenDenomsOnPath() []string {
	denoms := make([]string, 0, len(msg.Pools)+1)
	for i := 0; i < len(msg.Pools); i++ {
		denoms = append(denoms, msg.Pools[i].TokenInDenom)
	}
	denoms = append(denoms, msg.TokenOutDenom())
	return denoms
}

func (msg SwapAmountInSplitRouteWrapper) TokenInDenom() string {
	return msg.InDenom
}
func (msg SwapAmountInSplitRouteWrapper) TokenOutDenom() string {
	return msg.Pools[len(msg.Pools)-1].TokenOutDenom
}
func (msg SwapAmountInSplitRouteWrapper) TokenDenomsOnPath() []string {
	denoms := make([]string, 0, len(msg.Pools)+1)
	denoms = append(denoms, msg.TokenInDenom())
	for i := 0; i < len(msg.Pools); i++ {
		denoms = append(denoms, msg.Pools[i].TokenOutDenom)
	}
	return denoms
}

func (msg MsgSwapExactAmountOut) TokenInDenom() string {
	return msg.Routes[0].GetTokenInDenom()
}

func (msg MsgSwapExactAmountOut) TokenOutDenom() string {
	return msg.TokenOut.Denom
}

func (msg MsgSwapExactAmountOut) TokenDenomsOnPath() []string {
	denoms := make([]string, 0, len(msg.Routes)+1)
	for i := 0; i < len(msg.Routes); i++ {
		denoms = append(denoms, msg.Routes[i].TokenInDenom)
	}
	denoms = append(denoms, msg.TokenOutDenom())
	return denoms
}

func (msg MsgSwapExactAmountIn) TokenInDenom() string {
	return msg.TokenIn.Denom
}

func (msg MsgSwapExactAmountIn) TokenOutDenom() string {
	lastRouteIndex := len(msg.Routes) - 1
	return msg.Routes[lastRouteIndex].GetTokenOutDenom()
}

func (msg MsgSwapExactAmountIn) TokenDenomsOnPath() []string {
	denoms := make([]string, 0, len(msg.Routes)+1)
	denoms = append(denoms, msg.TokenInDenom())
	for i := 0; i < len(msg.Routes); i++ {
		denoms = append(denoms, msg.Routes[i].TokenOutDenom)
	}
	return denoms
}
