package lockup_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/osmosis-labs/osmosis/app"
	"github.com/osmosis-labs/osmosis/x/incentives"
	incentivestypes "github.com/osmosis-labs/osmosis/x/incentives/types"
	"github.com/osmosis-labs/osmosis/x/lockup"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestLockupInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	startTime := time.Now()
	ctx = ctx.WithBlockTime(startTime)

	addr := sdk.AccAddress([]byte("addr1---------------"))

	// init genesis for incentives module
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	distrTo := types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	incentives.InitGenesis(ctx, app.IncentivesKeeper, incentivestypes.GenesisState{
		Params: incentivestypes.DefaultParams(),
		Gauges: []incentivestypes.Gauge{
			{
				Id:                1,
				IsPerpetual:       false,
				DistributeTo:      distrTo,
				Coins:             coins,
				NumEpochsPaidOver: 2,
				FilledEpochs:      0,
				DistributedCoins:  sdk.Coins(nil),
				StartTime:         startTime.UTC(),
			},
		},
		LockableDurations: []time.Duration{time.Second, time.Second * 2},
	})

	// init genesis for lockup module
	lockup.InitGenesis(ctx, app.LockupKeeper, types.GenesisState{
		LastLockId: 2,
		Locks: []types.PeriodLock{
			{
				ID:       1,
				Owner:    addr.String(),
				Duration: time.Second,
				EndTime:  time.Time{},
				Coins:    sdk.Coins{sdk.NewInt64Coin("lptoken", 10)},
			},
			{
				ID:       2,
				Owner:    addr.String(),
				Duration: time.Second * 2,
				EndTime:  time.Time{},
				Coins:    sdk.Coins{sdk.NewInt64Coin("lptoken", 15)},
			},
		},
	})

	estRewards := app.IncentivesKeeper.GetRewardsEst(ctx, addr, []types.PeriodLock{}, 3)
	require.Equal(t, estRewards.String(), "10000stake")
}
