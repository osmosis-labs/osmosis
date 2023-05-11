package accum

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ZeroSharesError = errors.New("shares must be non-zero")
)

type NoPositionError struct {
	Name string
}

func (e NoPositionError) Error() string {
	return fmt.Sprintf("no position found for position key (%s)", e.Name)
}

type NegativeIntervalAccumulationPerShareError struct {
	IntervalAccumulationPerShare sdk.DecCoins
}

func (e NegativeIntervalAccumulationPerShareError) Error() string {
	return fmt.Sprintf("interval accumulation per share must be non-negative, was (%s)", e.IntervalAccumulationPerShare)
}

type NegativeAccDifferenceError struct {
	AccumulatorDifference sdk.DecCoins
}

func (e NegativeAccDifferenceError) Error() string {
	return fmt.Sprintf("difference (%s) between the old and the new accumulator value is negative", e.AccumulatorDifference)
}

type AccumDoesNotExistError struct {
	AccumName string
}

func (e AccumDoesNotExistError) Error() string {
	return fmt.Sprintf("Accumulator name %s does not exist in store", e.AccumName)
}

type NegativeRewardsAdditionError struct {
	AccumName    string
	PositionName string
}

func (e NegativeRewardsAdditionError) Error() string {
	return fmt.Sprintf("Attempted to add negative rewards to position %s of the accumulator %s", e.PositionName, e.AccumName)
}
