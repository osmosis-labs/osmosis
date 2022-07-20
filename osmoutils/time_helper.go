package osmoutils

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func FormatTimeString(t time.Time) string {
	return t.UTC().Round(0).Format(sdk.SortableTimeFormat)
}

// Parses a string encoded using FormatTimeString back into a time.Time
func ParseTimeString(s string) (time.Time, error) {
	t, err := time.Parse(sdk.SortableTimeFormat, s)
	if err != nil {
		return t, err
	}
	return t.UTC().Round(0), nil
}
