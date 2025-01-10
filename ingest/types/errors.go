package types

import "fmt"

type ConcentratedPoolNoTickModelError struct {
	PoolId uint64
}

func (e ConcentratedPoolNoTickModelError) Error() string {
	return fmt.Sprintf("concentrated pool (%d) has no tick model", e.PoolId)
}

type OrderbookPoolInvalidDirectionError struct {
	Direction int64
}

func (e OrderbookPoolInvalidDirectionError) Error() string {
	return fmt.Sprintf("orderbook pool direction (%d) is invalid; must be either -1 or 1", e.Direction)
}
