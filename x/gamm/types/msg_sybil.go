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
	// _ SybilResistantFee = MsgJoinSwapExternAmountIn{}
	// _ SybilResistantFee = MsgJoinSwapShareAmountOut{}
	// _ SybilResistantFee = MsgExitSwapShareAmountIn{}
	// _ SybilResistantFee = MsgExitSwapExternAmountOut{}
)

// MsgSwapExactAmountOut implements SybilResistantFees

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

// MsgSwapExactAmountIn implements SybilResistantFees

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

// TODO: correctly implement each of these msg
// In order, add:
// 		- MsgJoinSwapExternAmountIn
//		- MsgExitSwapExternAmountOut
//		- MsgJoinSwapShareAmountOut
//		- MsgExitSwapShareAmountOut
// MsgJoinSwapExternAmountIn implements SybilResistantFees
//
// func (msg MsgJoinSwapExternAmountIn) PoolIdOnPath() (path []uint64) {
// 	path[0] = msg.PoolId
// 	return path
// }

// func (msg MsgJoinSwapExternAmountIn) GetTokenToFee() sdk.Coin {
// 	return msg.TokenIn
// }

// func (msg MsgJoinSwapExternAmountIn) TokenDenomsOnPath() (denom []string) {
// 	denom[0] = msg.TokenIn.Denom
// 	return denom
// }

// MsgJoinSwapSharesAmountOut implements SybilResistantFees
//
// func (msg MsgJoinSwapShareAmountOut) GetTokenToFee() sdk.Coin {
// 	return sdk.NewCoin(msg.TokenInDenom, msg.TokenInMaxAmount)
// }

// func (msg MsgJoinSwapShareAmountOut) PoolIdOnPath() (path []uint64) {
// 	path[0] = msg.PoolId
// 	return path
// }

// func (msg MsgJoinSwapShareAmountOut) TokenDenomsOnPath() (denom []string) {
// 	denom[0] = msg.TokenInDenom
// 	return denom
// }

// MsgExitSwapShareAmountIn implements SybilResistantFees
//
// func (msg MsgExitSwapShareAmountIn) GetTokenToFee() sdk.Coin {
// 	return sdk.NewCoin(msg.TokenOutDenom, msg.TokenOutMinAmount)
// }

// func (msg MsgExitSwapShareAmountIn) PoolIdOnPath() (path []uint64) {
// 	path[0] = msg.PoolId
// 	return path
// }

// func (msg MsgExitSwapShareAmountIn) TokenDenomsOnPath() (denom []string) {
// 	denom[0] = msg.TokenOutDenom
// 	return denom
// }

// MsgExitSwapExternAmountOut implements SybilResistantFees
//
// func (msg MsgExitSwapExternAmountOut) GetTokenToFee() sdk.Coin {
// 	return sdk.NewCoin(msg.TokenOut.Denom, msg.TokenOut.Amount)
// }

// func (msg MsgExitSwapExternAmountOut) PoolIdOnPath() (path []uint64) {
// 	path[0] = msg.PoolId
// 	return path
// }

// func (msg MsgExitSwapExternAmountOut) TokenDenomsOnPath() (denom []string) {
// 	denom[0] = msg.TokenOut.Denom
// 	return denom
// }
