package incentives_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	osmoapp "github.com/osmosis-labs/osmosis/app"
	"github.com/osmosis-labs/osmosis/x/incentives"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestIncentivesExportGenesis(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	genesis := incentives.ExportGenesis(ctx, app.IncentivesKeeper)
	require.Equal(t, genesis.Params.DistrEpochIdentifier, "week")
	require.Len(t, genesis.Gauges, 0)

	addr := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	startTime := time.Now()
	err := simapp.FundAccount(app.BankKeeper, ctx, addr, coins)
	require.NoError(t, err)
	gaugeID, err := app.IncentivesKeeper.CreateGauge(ctx, true, addr, coins, distrTo, startTime, 1)
	require.NoError(t, err)

	genesis = incentives.ExportGenesis(ctx, app.IncentivesKeeper)
	require.Equal(t, genesis.Params.DistrEpochIdentifier, "week")
	require.Len(t, genesis.Gauges, 1)

	require.Equal(t, genesis.Gauges[0], types.Gauge{
		Id:                gaugeID,
		IsPerpetual:       true,
		DistributeTo:      distrTo,
		Coins:             coins,
		NumEpochsPaidOver: 1,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins(nil),
		StartTime:         startTime.UTC(),
	})

	denom := "stake"
	duration := time.Hour
	currentReward := types.CurrentReward{
		LastProcessedEpoch: 5,
		TotalShares:        sdk.NewCoin("stake", sdk.NewInt(10)),
		Denom:              denom,
		LockDuration:       duration,
	}
	app.IncentivesKeeper.SetCurrentReward(ctx, currentReward, denom, duration)

	cumulativeRewardRatio := sdk.NewDecCoins(sdk.NewInt64DecCoin(denom, 1000))

	app.IncentivesKeeper.SetHistoricalReward(ctx, cumulativeRewardRatio, denom, duration, 1)
	app.IncentivesKeeper.SetHistoricalReward(ctx, cumulativeRewardRatio, denom, duration, 3)
	app.IncentivesKeeper.SetHistoricalReward(ctx, cumulativeRewardRatio, denom, duration, 5)

	// periodLockReward := types.PeriodLockReward{
	// 	LockId: 1,
	// 	LastEligibleEpochs: []*types.LastEligibleEpochByDurationAndDenom{
	// 		{
	// 			Denom:        "gamm/pool/1",
	// 			LockDuration: time.Hour,
	// 			Epoch:        1,
	// 		},
	// 	},
	// }
	// app.IncentivesKeeper.SetPeriodLockReward(ctx, periodLockReward)
	// genesis = incentives.ExportGenesis(ctx, app.IncentivesKeeper)

	// historicalReward1 := types.HistoricalReward{
	// 	CumulativeRewardRatio: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom, 1000)),
	// 	LastEligibleEpoch:     1,
	// }
	// historicalReward2 := types.HistoricalReward{
	// 	CumulativeRewardRatio: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom, 1000)),
	// 	LastEligibleEpoch:     3,
	// }
	// historicalReward3 := types.HistoricalReward{
	// 	CumulativeRewardRatio: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom, 1000)),
	// 	LastEligibleEpoch:     5,
	// }

	// require.Equal(t, genesis.GenesisReward[0].CurrentReward, currentReward)
	// require.Equal(t, genesis.GenesisReward[0].HistoricalReward[0], historicalReward1)
	// require.Equal(t, genesis.GenesisReward[0].HistoricalReward[1], historicalReward2)
	// require.Equal(t, genesis.GenesisReward[0].HistoricalReward[2], historicalReward3)
	// require.Equal(t, genesis.PeriodLockReward[0], periodLockReward)
}

func TestIncentivesInitGenesis(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	validateGenesis := types.DefaultGenesis().Params.Validate()
	require.NoError(t, validateGenesis)

	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	startTime := time.Now()
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	gauge := types.Gauge{
		Id:                1,
		IsPerpetual:       false,
		DistributeTo:      distrTo,
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins(nil),
		StartTime:         startTime.UTC(),
	}
	incentives.InitGenesis(ctx, app.IncentivesKeeper, types.GenesisState{
		Params: types.Params{
			DistrEpochIdentifier: "week",
		},
		Gauges: []types.Gauge{gauge},
		LockableDurations: []time.Duration{
			time.Second,
			time.Hour,
			time.Hour * 3,
			time.Hour * 7,
		},
	})

	gauges := app.IncentivesKeeper.GetGauges(ctx)
	require.Len(t, gauges, 1)
	require.Equal(t, gauges[0], gauge)

}
