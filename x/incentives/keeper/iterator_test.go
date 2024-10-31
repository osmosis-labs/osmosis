package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v27/x/incentives/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

// This test validates that the FilterLocksByMinDuration function
// copies the correct locks when filtering.
// It helps to ensure that we do not use a wrong pointer by referencing
// a loop variable.
func TestFilterLocksByMinDuration(t *testing.T) {
	const (
		numLocks    = 3
		minDuration = 2
	)

	locks := make([]lockuptypes.PeriodLock, numLocks)
	for i := 0; i < numLocks; i++ {
		locks[i] = lockuptypes.PeriodLock{
			ID:       uint64(i + 1),
			Duration: minDuration,
		}
	}

	scratchSlice := []*lockuptypes.PeriodLock{}
	filteredLocks := keeper.FilterLocksByMinDuration(locks, minDuration, &scratchSlice)

	require.Equal(t, len(locks), len(filteredLocks))

	for i := 0; i < len(locks); i++ {
		require.Equal(t, locks[i].ID, filteredLocks[i].ID)
	}
}
