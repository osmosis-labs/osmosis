package usecase

import (
	"errors"
	"fmt"
)

var (
	ErrNilCurrentRoute     = errors.New("currentRoute cannot be nil")
	ErrNilRouterRepository = errors.New("router repository is not set")
	ErrNilPoolsRepository  = errors.New("pools repository is not set")
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

type TokenOutDenomMatchesTokenInDenomError struct {
	Denom string
}

func (e TokenOutDenomMatchesTokenInDenomError) Error() string {
	return fmt.Sprintf("token out denom matches token in denom (%s). Must be different", e.Denom)
}

type NoPoolsInRouteError struct {
	RouteIndex int
}

func (e NoPoolsInRouteError) Error() string {
	return fmt.Sprintf("route %d has no pools", e.RouteIndex)
}

type TokenOutMismatchBetweenRoutesError struct {
	TokenOutDenomRouteA string
	TokenOutDenomRouteB string
}

func (e TokenOutMismatchBetweenRoutesError) Error() string {
	return fmt.Sprintf("all routes must have the same final token out denom. Observed (%s) and (%s)", e.TokenOutDenomRouteA, e.TokenOutDenomRouteB)
}

type RoutePoolWithTokenInDenomError struct {
	RouteIndex   int
	TokenInDenom string
}

func (e RoutePoolWithTokenInDenomError) Error() string {
	return fmt.Sprintf("route %d has an intermediary pool with token in denom %s", e.RouteIndex, e.TokenInDenom)
}

type RoutePoolWithTokenOutDenomError struct {
	RouteIndex    int
	TokenOutDenom string
}

func (e RoutePoolWithTokenOutDenomError) Error() string {
	return fmt.Sprintf("route %d has an intermediary pool with token out denom %s", e.RouteIndex, e.TokenOutDenom)
}

type PreviousTokenOutDenomNotInPoolError struct {
	RouteIndex            int
	PoolId                uint64
	PreviousTokenOutDenom string
}

func (e PreviousTokenOutDenomNotInPoolError) Error() string {
	return fmt.Sprintf("previous token out denom (%s) not found in pool (%d), route index (%d)", e.PreviousTokenOutDenom, e.PoolId, e.RouteIndex)
}

type CurrentTokenOutDenomNotInPoolError struct {
	RouteIndex           int
	PoolId               uint64
	CurrentTokenOutDenom string
}

func (e CurrentTokenOutDenomNotInPoolError) Error() string {
	return fmt.Sprintf("current token out denom (%s) not found in pool (%d), route index (%d)", e.CurrentTokenOutDenom, e.PoolId, e.RouteIndex)
}
