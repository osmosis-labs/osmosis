package osmoutils

import "fmt"

type DecNotFoundError struct {
	Key string
}

func (e DecNotFoundError) Error() string {
	return fmt.Sprintf("no osmomath.Dec at key (%s)", e.Key)
}
