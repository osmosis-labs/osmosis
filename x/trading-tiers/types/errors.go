package types

import (
	"errors"
	"fmt"

	"github.com/osmosis-labs/osmosis/osmomath"
)

var (
	ErrAccountAlreadyOptedIn = errors.New("account already opted in")
)

type InsufficientStakeError struct {
	MinStake  osmomath.Int
	BondedAmt osmomath.Int
}

func (e InsufficientStakeError) Error() string {
	return fmt.Sprintf("insufficient stake: minimum required is %s, but bonded amount is %s", e.MinStake, e.BondedAmt)
}
