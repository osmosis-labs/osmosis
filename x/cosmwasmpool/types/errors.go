package types

import "fmt"

type InvalidPoolTypeErr struct {
	ActualPool interface{}
}

func (e InvalidPoolTypeErr) Error() string {
	return fmt.Sprintf("given pool does not implement cosmwasm pool extension, implements %T", e.ActualPool)
}

type PoolNotFoundError struct {
	PoolId uint64
}

func (e PoolNotFoundError) Error() string {
	return fmt.Sprintf("pool not found. pool id (%d)", e.PoolId)
}
