package domain

import (
	"errors"
	"fmt"
)

var (
	// ErrInternalServerError will throw if any the Internal Server Error happen
	ErrInternalServerError = errors.New("internal Server Error")
	// ErrNotFound will throw if the requested item is not exists
	ErrNotFound = errors.New("your requested Item is not found")
	// ErrConflict will throw if the current action already exists
	ErrConflict = errors.New("your Item already exist")
	// ErrBadParamInput will throw if the given request-body or params is not valid
	ErrBadParamInput = errors.New("given Param is not valid")
)

// InvalidPoolTypeError is an error type for invalid pool type.
type InvalidPoolTypeError struct {
	PoolType int32
}

func (e InvalidPoolTypeError) Error() string {
	return "invalid pool type: " + string(e.PoolType)
}

type ConcentratedPoolNoTickModelError struct {
	PoolId uint64
}

func (e ConcentratedPoolNoTickModelError) Error() string {
	return fmt.Sprintf("concentrated pool (%d) has no tick model", e.PoolId)
}

type TakerFeeNotFoundForDenomPairError struct {
	Denom0 string
	Denom1 string
}

func (e TakerFeeNotFoundForDenomPairError) Error() string {
	return fmt.Sprintf("taker fee not found for denom pair (%s, %s)", e.Denom0, e.Denom1)
}
