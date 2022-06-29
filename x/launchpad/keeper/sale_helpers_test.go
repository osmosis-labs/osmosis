package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/launchpad/types"
)

func TestCurrentRound(t *testing.T) {
	start := time.Unix(100, 0)
	before := start.Add(-2 * types.ROUND)
	end1 := start.Add(2 * types.ROUND)
	end2 := start.Add(2*types.ROUND + types.ROUND/2)
	after := end2.Add(2 * types.ROUND)
	tcs := []struct {
		start    time.Time
		end      time.Time
		now      time.Time
		expected int64
	}{
		{start, end1, before, 0},
		{start, end1, start, 0},
		{start, end1, start.Add(types.ROUND / 2), 0},
		{start, end1, start.Add(types.ROUND), 1},
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

func checkUser(require *require.Assertions, u *types.UserPosition, shares, staked, outPerShare, purchased sdk.Int, msg interface{}) {
	require.Equal(shares.String(), u.Shares.String(), msg, "shares")
	require.Equal(staked.String(), u.Staked.String(), msg, "staked")
	require.Equal(outPerShare.String(), u.OutPerShare.String(), msg, "outPerShare")
	require.Equal(purchased.String(), u.Purchased.String(), msg, "purchased")
}

func checkSale(require *require.Assertions, p *types.Sale, round int64, outRemainig, outSold, outPerShare, staked, income, shares sdk.Int) {
	require.Equal(round, p.Round, "round")
	require.Equal(outRemainig.String(), p.OutRemaining.String(), "outRemaining")
	require.Equal(outSold.String(), p.OutSold.String(), "outSold")
	require.Equal(outPerShare.String(), p.OutPerShare.String(), "outPerShare")
	require.Equal(staked.String(), p.Staked.String(), "staked")
	require.Equal(income.String(), p.Income.String(), "income")
	require.Equal(shares.String(), p.Shares.String(), "shares")
}
