package osmoutils

import (
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func FormatFixedLengthU64(d uint64) string {
	return fmt.Sprintf("%0.20d", d)
}

func FormatTimeString(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10) // unix time
}

// Parses a string encoded using FormatTimeString back into a time.Time
func ParseTimeString(s string) (time.Time, error) {
	t, err := time.Parse(sdk.SortableTimeFormat, s)
	if err != nil {
		return t, err
	}
	return t.UTC().Round(0), nil // TODO: no references on this, maybe change to unix too
}
