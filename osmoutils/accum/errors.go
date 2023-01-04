package accum

import (
	"fmt"
)

type NoPositionError struct {
	Name string
}

func (e NoPositionError) Error() string {
	return fmt.Sprintf("no position found for address (%s)", e.Name)
}
