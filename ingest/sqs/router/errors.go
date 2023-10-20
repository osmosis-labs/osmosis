package router

import (
	"errors"
	"fmt"
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

var (
	ErrNilCurrentRoute = errors.New("currentRoute cannot be nil")
)
