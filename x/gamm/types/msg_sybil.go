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
	_ SybilResistantFee = MsgJoinSwapExternAmountIn{}
	_ SybilResistantFee = MsgJoinSwapShareAmountOut{}
	_ SybilResistantFee = MsgExitSwapShareAmountIn{}
	_ SybilResistantFee = MsgExitSwapExternAmountOut{}
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
//
// MsgJoinSwapExternAmountIn implements SybilResistantFee
//
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

// MsgJoinSwapSharesAmountOut implements SybilResistantFee

func (msg MsgJoinSwapShareAmountOut) GetTokenToFee() sdk.Coin {
	return sdk.NewCoin(msg.TokenInDenom, msg.TokenInMaxAmount)
}

func (msg MsgJoinSwapShareAmountOut) PoolIdOnPath() (path []uint64) {
	path[0] = msg.PoolId
	return path
}

func (msg MsgJoinSwapShareAmountOut) TokenDenomsOnPath() (denom []string) {
	denom[0] = msg.TokenInDenom
	return denom
}

// MsgExitSwapShareAmountIn implements SybilResistantFee

func (msg MsgExitSwapShareAmountIn) GetTokenToFee() sdk.Coin {
	return sdk.NewCoin(msg.TokenOutDenom, msg.TokenOutMinAmount)
}

func (msg MsgExitSwapShareAmountIn) PoolIdOnPath() (path []uint64) {
	path[0] = msg.PoolId
	return path
}

func (msg MsgExitSwapShareAmountIn) TokenDenomsOnPath() (denom []string) {
	denom[0] = msg.TokenOutDenom
	return denom
}

// MsgExitSwapExternAmountOut implements SybilResistantFee
//
func (msg MsgExitSwapExternAmountOut) GetTokenToFee() sdk.Coin {
	return sdk.NewCoin(msg.TokenOut.Denom, msg.TokenOut.Amount)
}

func (msg MsgExitSwapExternAmountOut) PoolIdOnPath() (path []uint64) {
	path[0] = msg.PoolId
	return path
}

func (msg MsgExitSwapExternAmountOut) TokenDenomsOnPath() (denom []string) {
	denom[0] = msg.TokenOut.Denom
	return denom
}
