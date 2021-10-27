package incentives_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/osmosis-labs/osmosis/app"
	"github.com/osmosis-labs/osmosis/x/incentives"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestIncentivesExportGenesis(t *testing.T) {
	app := simapp.Setup(false)
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
	err := app.BankKeeper.SetBalances(ctx, addr, coins)
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
		Period:             4,
		LastProcessedEpoch: 7,
		Coin:               sdk.NewCoin("stake", sdk.NewInt(10)),
		Denom:              denom,
		Duration:           duration,
	}
	app.IncentivesKeeper.SetCurrentReward(ctx, currentReward, denom, duration)
	historicalReward1 := types.HistoricalReward{
		Period:            1,
		LastEligibleEpoch: 1,
	}
	historicalReward2 := types.HistoricalReward{
		CumulativeRewardRatio: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom, 1000)),
		Period:                2,
		LastEligibleEpoch:     3,
	}
	historicalReward3 := types.HistoricalReward{
		CumulativeRewardRatio: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom, 1000)),
		Period:                3,
		LastEligibleEpoch:     5,
	}
	app.IncentivesKeeper.AddHistoricalReward(ctx, historicalReward1, denom, duration, 1, 1)
	app.IncentivesKeeper.AddHistoricalReward(ctx, historicalReward2, denom, duration, 2, 3)
	app.IncentivesKeeper.AddHistoricalReward(ctx, historicalReward3, denom, duration, 3, 5)
	periodLockReward := types.PeriodLockReward{
		ID:     1,
		Period: map[string]uint64{"gamm/pool/1/1h0s": 1},
	}
	app.IncentivesKeeper.SetPeriodLockReward(ctx, periodLockReward)
	genesis = incentives.ExportGenesis(ctx, app.IncentivesKeeper)
	require.Equal(t, genesis.GenesisReward[0].CurrentReward, currentReward)
	require.Equal(t, genesis.GenesisReward[0].HistoricalReward[0], historicalReward1)
	require.Equal(t, genesis.GenesisReward[0].HistoricalReward[1], historicalReward2)
	require.Equal(t, genesis.GenesisReward[0].HistoricalReward[2], historicalReward3)
	require.Equal(t, genesis.PeriodLockReward[0], periodLockReward)
}

func TestIncentivesInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
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
	denom := "stake"
	duration := time.Hour
	currentReward := types.CurrentReward{
		Period:             2,
		LastProcessedEpoch: 9,
		Coin:               sdk.NewCoin("stake", sdk.NewInt(10)),
		Denom:              denom,
		Duration:           duration,
	}
	historicalReward := types.HistoricalReward{
		CumulativeRewardRatio: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom, 1000)),
		Period:                1,
		LastEligibleEpoch:     1,
	}
	genesisReward := types.GenesisReward{
		CurrentReward:    currentReward,
		HistoricalReward: []types.HistoricalReward{historicalReward},
	}
	periodLockReward := types.PeriodLockReward{
		ID:     1,
		Period: map[string]uint64{"gamm/pool/1/1h0s": 1},
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
		GenesisReward:    []types.GenesisReward{genesisReward},
		PeriodLockReward: []types.PeriodLockReward{periodLockReward},
	})

	gauges := app.IncentivesKeeper.GetGauges(ctx)
	require.Len(t, gauges, 1)
	require.Equal(t, gauges[0], gauge)
	currentRewardRestored, _ := app.IncentivesKeeper.GetCurrentReward(ctx, denom, duration)
	require.Equal(t, currentRewardRestored, currentReward)
	periodLockRewardRestored, _ := app.IncentivesKeeper.GetPeriodLockReward(ctx, 1)
	require.Equal(t, periodLockRewardRestored, periodLockReward)
}
