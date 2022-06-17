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
	return msg.TokenDenomsOnPath()
}

func (msg MsgSwapExactAmountOut) GetTokenToFee() sdk.Coin {
	return msg.GetTokenOut()
}

func (msg MsgSwapExactAmountOut) GetPoolIdOnPath() []uint64 {
	ids := make([]uint64, 0, len(msg.Routes))
	for i := 0; i < len(msg.Routes); i++ {
		ids = append(ids, msg.Routes[i].PoolId)
	}
	return ids
}

// MsgSwapExactAmountIn implements SybilResistantFee

func (msg MsgSwapExactAmountIn) GetTokenToFee() sdk.Coin {
	return msg.GetTokenIn()
}

func (msg MsgSwapExactAmountIn) GetPoolIdOnPath() []uint64 {
	ids := make([]uint64, 0, len(msg.Routes))
	for i := 0; i < len(msg.Routes); i++ {
		ids = append(ids, msg.Routes[i].PoolId)
	}
	return ids
}

func (msg MsgSwapExactAmountIn) GetTokenDenomsOnPath() []string {
	return msg.TokenDenomsOnPath()
}
