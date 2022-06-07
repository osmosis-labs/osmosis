package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Sybil Resistant Fee Swap Msg defines a simple interface for getting
// - poolIds on a swap msg route
// - token to apply the sybil resistanct fees to
type SybilResistantFeeSwap interface {
	TokenDenomsOnPath() []string
	PoolIdOnPath() []uint64
	GetTokenToFee() sdk.Coin
}

var (
	_ SybilResistantFeeSwap = MsgSwapExactAmountOut{}
	_ SybilResistantFeeSwap = MsgSwapExactAmountIn{}
	_ SybilResistantFeeSwap = MsgJoinSwapExternAmountIn{}
)

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

func (msg MsgJoinSwapExternAmountIn) GetTokenToFee() sdk.Coin {
	return msg.TokenIn
}

func (msg MsgJoinSwapExternAmountIn) PoolIdOnPath() (path []uint64) {
	path[0] = msg.PoolId
	return path
}

func (msg MsgJoinSwapExternAmountIn) TokenDenomsOnPath() (denom []string) {
	denom[0] = msg.TokenIn.Denom
	return denom
}
