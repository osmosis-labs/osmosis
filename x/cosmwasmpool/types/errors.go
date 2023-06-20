package types

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyPoolIds                         = errors.New("pool id list cannot be empty")
	ErrNoneOfCodeIdAndContractCodeSpecified = errors.New("both code id and byte code are unset. Only one must be specified.")
	ErrBothOfCodeIdAndContractCodeSpecified = errors.New("both code id and byte code are set. Only one must be specified.")
)

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

type CodeIdNotWhitelistedError struct {
	CodeId uint64
}

func (e CodeIdNotWhitelistedError) Error() string {
	return fmt.Sprintf("cannot create coswasm pool with the given code id (%d). Please whitelist it via governance", e.CodeId)
}
