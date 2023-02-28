package osmoutils

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatFixedLengthU64(t *testing.T) {
	tests := map[string]struct {
		d    uint64
		want string
	}{
		"0":       {0, "00000000000000000000"},
		"1":       {1, "00000000000000000001"},
		"9":       {9, "00000000000000000009"},
		"10":      {10, "00000000000000000010"},
		"123":     {123, "00000000000000000123"},
		"max u64": {math.MaxUint64, "18446744073709551615"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := FormatFixedLengthU64(tt.d)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, len(got), 20)
		})
	}
}

func TestParseTimeString(t *testing.T) {
	tests := map[string]struct {
		timeStr string
		want    time.Time
	}{
		"0": {"2023-03-03 08:59:42.68331893 +0000 UTC", time.Now()},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseTimeString(tt.timeStr)
			assert.NoError(t, err)
			fmt.Println(got)

		})
	}
}
