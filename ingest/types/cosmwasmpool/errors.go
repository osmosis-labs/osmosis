package cosmwasmpool

import "fmt"

type OrderbookUnsupportedDenomError struct {
	Denom      string
	QuoteDenom string
	BaseDenom  string
}

func (e OrderbookUnsupportedDenomError) Error() string {
	return fmt.Sprintf("Denom (%s) is not supported by orderbook (%s/%s)", e.Denom, e.BaseDenom, e.QuoteDenom)
}

type DuplicatedDenomError struct {
	Denom string
}

func (e DuplicatedDenomError) Error() string {
	return fmt.Sprintf("Denom (%s) is duplicated", e.Denom)
}

type OrderbookOrderNotAvailableError struct {
	PoolId    uint64
	Direction OrderbookDirection
}

func (e OrderbookOrderNotAvailableError) Error() string {
	return fmt.Sprintf("There is no %s order in pool (%d)", e.Direction.String(), e.PoolId)
}
