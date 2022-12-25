package accum

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type NoPositionError struct {
	Address sdk.AccAddress
}

func (e NoPositionError) Error() string {
	return fmt.Sprintf("no position found for address (%s)", e.Address)
}
