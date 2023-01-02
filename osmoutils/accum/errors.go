package accum

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type NoPositionError struct {
	Name string
}

func (e NoPositionError) Error() string {
	return fmt.Sprintf("no position found for address (%s)", e.Name)
}

type NegativeCustomAccError struct {
	CustomAccumulatorValue sdk.DecCoins
}

func (e NegativeCustomAccError) Error() string {
	return fmt.Sprintf("customAccumulatorValue must be non-negative, was (%s)", e.CustomAccumulatorValue)
}
