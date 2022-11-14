package types

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyRoutes       = errors.New("provided empty routes")
	ErrInvalidPool       = errors.New("attempting to create an invalid pool")
	ErrTooFewPoolAssets  = errors.New("pool should have at least 2 assets, as they must be swapping between at least two assets")
	ErrTooManyPoolAssets = errors.New("pool has too many assets (currently capped at 8 assets per pool)")
)

type nonPositiveAmountError struct {
	Amount string
}

func (e nonPositiveAmountError) Error() string {
	return fmt.Sprintf("min out amount or max in amount should be positive, was (%s)", e.Amount)
}

type FailedToFindRouteError struct {
	PoolId uint64
}

func (e FailedToFindRouteError) Error() string {
	return fmt.Sprintf("failed to find route for pool id (%d)", e.PoolId)
}

type UndefinedRouteError struct {
	PoolType PoolType
	PoolId   uint64
}

func (e UndefinedRouteError) Error() string {
	return fmt.Sprintf("route is not defined for the given pool type (%s) and pool id (%d)", e.PoolType, e.PoolId)
}
