package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Sybil Resistant Fee Swap Msg defines a simple interface for getting
// - poolIds on a swap msg route
// - token to apply the sybil resistant fees to
type SybilResistantFee interface {
	TokenDenomsOnPath() []string
	PoolIdOnPath() []uint64
	GetTokenToFee() sdk.Coin
}

var (
	_ SybilResistantFee = MsgSwapExactAmountOut{}
	_ SybilResistantFee = MsgSwapExactAmountIn{}
)

// MsgSwapExactAmountOut implements SybilResistantFee

func (msg MsgSwapExactAmountOut) GetTokenToFee() sdk.Coin {
	return msg.GetTokenOut()
}

func (msg MsgSwapExactAmountOut) PoolIdOnPath() []uint64 {
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

func (msg MsgSwapExactAmountIn) PoolIdOnPath() []uint64 {
	ids := make([]uint64, 0, len(msg.Routes))
	for i := 0; i < len(msg.Routes); i++ {
		ids = append(ids, msg.Routes[i].PoolId)
	}
	return ids
}
