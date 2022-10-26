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

type tooFewPoolAssetsError struct {
	ActualNumPoolAssets int
}

func (e tooFewPoolAssetsError) Error() string {
	return fmt.Sprintf("pool should have at least 2 assets, as they must be swapping between at least two assets, was given (%d)", e.ActualNumPoolAssets)
}

type tooManyPoolAssetsError struct {
	ActualNumPoolAssets int
}

func (e tooManyPoolAssetsError) Error() string {
	return fmt.Sprintf("pool has too many assets (currently capped at 8 assets per pool, was given (%d)", e.ActualNumPoolAssets)
}
