package types

type LiquidityChangeType int

const (
	AddLiquidity LiquidityChangeType = iota
	RemoveLiquidity
)

// LiquidityChangeMsg defines a simple interface for determining if an LP msg
// is removing or adding liquidity.
type LiquidityChangeMsg interface {
	LiquidityChangeType() LiquidityChangeType
}

var (
	_ LiquidityChangeMsg = MsgExitPool{}
	_ LiquidityChangeMsg = MsgExitSwapShareAmountIn{}
	_ LiquidityChangeMsg = MsgExitSwapExternAmountOut{}
)

var (
	_ LiquidityChangeMsg = MsgJoinPool{}
	_ LiquidityChangeMsg = MsgJoinSwapExternAmountIn{}
	_ LiquidityChangeMsg = MsgJoinSwapShareAmountOut{}
)

func (msg MsgExitPool) LiquidityChangeType() LiquidityChangeType {
	return RemoveLiquidity
}

func (msg MsgExitSwapShareAmountIn) LiquidityChangeType() LiquidityChangeType {
	return RemoveLiquidity
}

func (msg MsgExitSwapExternAmountOut) LiquidityChangeType() LiquidityChangeType {
	return RemoveLiquidity
}

func (msg MsgJoinPool) LiquidityChangeType() LiquidityChangeType {
	return AddLiquidity
}

func (msg MsgJoinSwapExternAmountIn) LiquidityChangeType() LiquidityChangeType {
	return AddLiquidity
}

func (msg MsgJoinSwapShareAmountOut) LiquidityChangeType() LiquidityChangeType {
	return AddLiquidity
}
