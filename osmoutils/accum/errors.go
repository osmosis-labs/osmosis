package accum

import (
	"fmt"
)

type NoPositionError struct {
	Address string
}

func (e NoPositionError) Error() string {
	return fmt.Sprintf("no position found for address (%s)", e.Address)
}
