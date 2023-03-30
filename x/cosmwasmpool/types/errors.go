package types

import "fmt"

type InvalidPoolTypeError struct {
	ActualPool interface{}
}

func (e InvalidPoolTypeError) Error() string {
	return fmt.Sprintf("given pool does not implement cosmwasm pool extension, implements %T", e.ActualPool)
}

type PoolNotFoundError struct {
	PoolId uint64
}

func (e PoolNotFoundError) Error() string {
	return fmt.Sprintf("pool not found. pool id (%d)", e.PoolId)
}

type InvalidLiquiditySetError struct {
	PoolId     uint64
	TokenCount int
}

func (e InvalidLiquiditySetError) Error() string {
	return fmt.Sprintf("pool %d does not have enough liquidity tokens, need at least 2, had (%d)", e.PoolId, e.TokenCount)
}
