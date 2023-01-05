package types

import (
	"testing"
	time "time"

	"github.com/stretchr/testify/require"
)

func TestDowntimeToDurationAscending(t *testing.T) {
	numEntries := 0
	lastDur := time.Duration(0)
	DowntimeToDuration.Ascend(Downtime(0), func(_ Downtime, v time.Duration) bool {
		numEntries++
		require.Greater(t, v, lastDur)
		return true
	})
	require.Equal(t, numEntries, 25)
}
