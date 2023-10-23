package usecase

import (
	"errors"
	"fmt"
)

var (
	ErrNilCurrentRoute = errors.New("currentRoute cannot be nil")
)

type SortedPoolsAndPoolsUsedLengthMismatchError struct {
	SortedPoolsLen int
	PoolsUsedLen   int
}

func (e SortedPoolsAndPoolsUsedLengthMismatchError) Error() string {
	return fmt.Sprintf("length of sorted pools (%d) and pools used (%d) must be the same", e.SortedPoolsLen, e.PoolsUsedLen)
}

type SortedPoolsAndPoolsInRouteLengthMismatchError struct {
	SortedPoolsLen int
	PoolsInRoute   int
}

func (e SortedPoolsAndPoolsInRouteLengthMismatchError) Error() string {
	return fmt.Sprintf("length of pools in route (%d) should not exceed length of sorted pools (%d)", e.PoolsInRoute, e.SortedPoolsLen)
}

type OnlyBalancerPoolsSupportedError struct {
	ActualType int32
}

func (e OnlyBalancerPoolsSupportedError) Error() string {
	return fmt.Sprintf("pool type (%d) is invalid. Only balancer is currently supported", e.ActualType)
}

type FailedToCastPoolModelError struct {
	ExpectedModel string
	ActualModel   string
}

func (e FailedToCastPoolModelError) Error() string {
	return fmt.Sprintf("failed to cast pool model (%s) to the desired type (%s)", e.ActualModel, e.ExpectedModel)
}
