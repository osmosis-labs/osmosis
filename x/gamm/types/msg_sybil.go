package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SybilResistantFee Swap Msg defines an interface for getting
// - poolIds on a swap msg route
// - token to apply the sybil resistant fees to
type SybilResistantFee interface {
	GetTokenDenomsOnPath() []string
	GetPoolIdOnPath() []uint64
	GetTokenToFee() sdk.Coin
}

var (
	_ SybilResistantFee = MsgSwapExactAmountOut{}
	_ SybilResistantFee = MsgSwapExactAmountIn{}
)

// MsgSwapExactAmountOut implements SybilResistantFee
func (msg MsgSwapExactAmountOut) GetTokenDenomsOnPath() []string {
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

// MsgSwapExactAmountIn implements SybilResistantFee

func (msg MsgSwapExactAmountIn) GetTokenDenomsOnPath() []string {
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
