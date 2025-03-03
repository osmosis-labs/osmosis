package osmoutils

import (
	"strconv"
)

func Uint64ToString(i uint64) string {
	return strconv.FormatUint(i, 10)
}
