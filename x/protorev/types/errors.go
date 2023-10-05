package types

import fmt "fmt"

type NoPoolForDenomPairError struct {
	BaseDenom  string
	MatchDenom string
}

func (e NoPoolForDenomPairError) Error() string {
	return fmt.Sprintf("highest liquidity pool between base %s and match denom %s not found", e.BaseDenom, e.MatchDenom)
}
