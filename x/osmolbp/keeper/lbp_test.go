package keeper

import (
	"testing"
	"time"

	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
	"github.com/stretchr/testify/assert"
)

func TestCurrentRound(t *testing.T) {
	start := time.Unix(100, 0)
	before := start.Add(-2 * api.ROUND)
	end1 := start.Add(2 * api.ROUND)
	end2 := start.Add(2*api.ROUND + api.ROUND/2)
	after := end2.Add(2 * api.ROUND)
	tcs := []struct {
		start    time.Time
		end      time.Time
		now      time.Time
		expected uint64
	}{
		{start, end1, before, 0},
		{start, end1, start, 0},
		{start, end1, start.Add(api.ROUND / 2), 0},
		{start, end1, start.Add(api.ROUND), 1},
		{start, end1, end1, 2},
		{start, end1, after, 2},

		{start, end1, end2, 2},
		{start, end1, after, 2},
	}
	assert := assert.New(t)
	for i, tc := range tcs {
		assert.Equal(tc.expected, currentRound(tc.start, tc.end, tc.now), "tc: %d", i)
	}
}
