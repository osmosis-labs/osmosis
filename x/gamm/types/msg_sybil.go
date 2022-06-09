package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Sybil Resistant Fee Swap Msg defines a simple interface for getting
// - poolIds on a swap msg route
// - token to apply the sybil resistant fees to
type SybilResistantSwapFees interface {
	TokenDenomsOnPath() []string
	PoolIdOnPath() []uint64
	GetSwapFeeToken() sdk.Coin
}

var (
	_ SybilResistantSwapFees = MsgSwapExactAmountOut{}
	_ SybilResistantSwapFees = MsgSwapExactAmountIn{}
	// _ SybilResistantFee = MsgJoinSwapExternAmountIn{}
	// _ SybilResistantFee = MsgJoinSwapShareAmountOut{}
	// _ SybilResistantFee = MsgExitSwapShareAmountIn{}
	// _ SybilResistantFee = MsgExitSwapExternAmountOut{}
)

func (msg MsgSwapExactAmountOut) PoolIdOnPath() []uint64 {
	ids := make([]uint64, 0, len(msg.Routes))
	for i := 0; i < len(msg.Routes); i++ {
		ids = append(ids, msg.Routes[i].PoolId)
	}
	return ids
}

func (msg MsgSwapExactAmountOut) GetSwapFeeToken() sdk.Coin {
	return msg.GetTokenOut()
}

func (msg MsgSwapExactAmountIn) PoolIdOnPath() []uint64 {
	ids := make([]uint64, 0, len(msg.Routes))
	for i := 0; i < len(msg.Routes); i++ {
		ids = append(ids, msg.Routes[i].PoolId)
	}
	return ids
}

func (msg MsgSwapExactAmountIn) GetSwapFeeToken() sdk.Coin {
	return msg.GetTokenIn()
}

func (msg MsgJoinSwapExternAmountIn) PoolIdOnPath() (path []uint64) {
	path[0] = msg.PoolId
	return path
}

func (msg MsgJoinSwapExternAmountIn) GetTokenToFee() sdk.Coin {
	return msg.TokenIn
}

func (msg MsgJoinSwapExternAmountIn) TokenDenomsOnPath() (denom []string) {
	denom[0] = msg.TokenIn.Denom
	return denom
}

// TODO: Add sybil resistant swap fees to single asset join/exit msgs
// 		- implement each Join/Exit swap msgs
//		- add get weights function to the GammKeeper for applying correct swap fees
//      - example logic:
//			- determine if balancer pool or stableswap
//				- stableswap -> 50-50 pool
//				- balancer -> get weights
//			- apply swap fee to the correct ratio
//
//	- Implement in this order:
// 		1. MsgJoinSwapExternAmountIn
//			- Fee applied to token in
//		2. MsgExitSwapExternAmountOut
//			- Fee applied to token out
//		3. MsgJoinSwapShareAmountOut
//			- Fee applied to token in
//		4. MsgExitSwapShareAmountIn
//			- Fee applied to token out
//		5. Apply to superfluid staking bonding fee somehow ?
// // *** Assumes 50/50 pool - change by considering weights in balancer pools
//
// func (msg MsgJoinSwapShareAmountOut) GetSwapFeeToken() sdk.Coin {
// 	return sdk.NewCoin(msg.TokenInDenom, msg.TokenInMaxAmount.QuoRaw(2))
// }

// func (msg MsgJoinSwapShareAmountOut) PoolIdOnPath() (path []uint64) {
// 	path[0] = msg.PoolId
// 	return path
// }

// func (msg MsgJoinSwapShareAmountOut) TokenDenomsOnPath() (denom []string) {
// 	denom[0] = msg.TokenInDenom
// 	return denom
// }

// // *** Assumes 50/50 pool
// func (msg MsgExitSwapShareAmountIn) GetSwapFeeToken() sdk.Coin {
// 	return sdk.NewCoin(msg.TokenOutDenom, msg.TokenOutMinAmount.QuoRaw(2))
// }

// func (msg MsgExitSwapShareAmountIn) PoolIdOnPath() (path []uint64) {
// 	path[0] = msg.PoolId
// 	return path
// }

// func (msg MsgExitSwapShareAmountIn) TokenDenomsOnPath() (denom []string) {
// 	denom[0] = msg.TokenOutDenom
// 	return denom
// }

// // *** Assumes 50/50 pool
// func (msg MsgExitSwapExternAmountOut) GetSwapFeeToken() sdk.Coin {
// 	return sdk.NewCoin(msg.TokenOut.Denom, msg.TokenOut.Amount.QuoRaw(2))
// }

// func (msg MsgExitSwapExternAmountOut) PoolIdOnPath() (path []uint64) {
// 	path[0] = msg.PoolId
// 	return path
// }

// func (msg MsgExitSwapExternAmountOut) TokenDenomsOnPath() (denom []string) {
// 	denom[0] = msg.TokenOut.Denom
// 	return denom
// }
