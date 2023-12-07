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

type PoolNotFoundError struct {
	PoolID uint64
}

func (e PoolNotFoundError) Error() string {
	return fmt.Sprintf("pool with ID (%d) is not found", e.PoolID)
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

type FailedToCastPoolModelError struct {
	ExpectedModel string
	ActualModel   string
}

func (e FailedToCastPoolModelError) Error() string {
	return fmt.Sprintf("failed to cast pool model (%s) to the desired type (%s)", e.ActualModel, e.ExpectedModel)
}

type ConcentratedNoLiquidityError struct {
	PoolId uint64
}

func (e ConcentratedNoLiquidityError) Error() string {
	return fmt.Sprintf("pool (%d) has no liquidity", e.PoolId)
}

type ConcentratedCurrentTickNotWithinBucketError struct {
	PoolId             uint64
	CurrentBucketIndex int64
	TotalBuckets       int64
}

func (e ConcentratedCurrentTickNotWithinBucketError) Error() string {
	return fmt.Sprintf("current bucket index (%d) is out of range (%d) for pool (%d)", e.CurrentBucketIndex, e.TotalBuckets, e.PoolId)
}

type ConcentratedCurrentTickAndBucketMismatchError struct {
	CurrentTick int64
	LowerTick   int64
	UpperTick   int64
}

func (e ConcentratedCurrentTickAndBucketMismatchError) Error() string {
	return fmt.Sprintf("current tick (%d) is not within bucket (%d, %d)", e.CurrentTick, e.LowerTick, e.UpperTick)
}

type ConcentratedZeroCurrentSqrtPriceError struct {
	PoolId uint64
}

func (e ConcentratedZeroCurrentSqrtPriceError) Error() string {
	return fmt.Sprintf("pool (%d) has zero current sqrt price", e.PoolId)
}

type ConcentratedNotEnoughLiquidityToCompleteSwapError struct {
	PoolId   uint64
	AmountIn string
}

func (e ConcentratedNotEnoughLiquidityToCompleteSwapError) Error() string {
	return fmt.Sprintf("not enough liquidity to complete swap in pool (%d) with amount in (%s)", e.PoolId, e.AmountIn)
}

type ConcentratedTickModelNotSetError struct {
	PoolId uint64
}

func (e ConcentratedTickModelNotSetError) Error() string {
	return fmt.Sprintf("tick model is not set on pool (%d)", e.PoolId)
}

type TransmuterInsufficientBalanceError struct {
	Denom         string
	BalanceAmount string
	Amount        string
}

func (e TransmuterInsufficientBalanceError) Error() string {
	return fmt.Sprintf("insufficient balance of token (%s), balance (%s), amount (%s)", e.Denom, e.BalanceAmount, e.Amount)
}

type StaleHeightError struct {
	StoredHeight            uint64
	TimeSinceLastUpdate     int
	MaxAllowedTimeDeltaSecs int
}

func (e StaleHeightError) Error() string {
	return fmt.Sprintf("stored height (%d) is stale, time since last update (%d), max allowed seconds (%d)", e.StoredHeight, e.TimeSinceLastUpdate, e.MaxAllowedTimeDeltaSecs)
}
