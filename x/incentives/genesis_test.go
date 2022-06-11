package incentives_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	osmoapp "github.com/osmosis-labs/osmosis/v10/app"
	"github.com/osmosis-labs/osmosis/v10/x/incentives"
	"github.com/osmosis-labs/osmosis/v10/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v10/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestIncentivesExportGenesis(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	genesis := incentives.ExportGenesis(ctx, *app.IncentivesKeeper)
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

	// mints coins so supply exists on chain
	mintLPtokens := sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	err = simapp.FundAccount(app.BankKeeper, ctx, addr, mintLPtokens)
	require.NoError(t, err)

	gaugeID, err := app.IncentivesKeeper.CreateGauge(ctx, true, addr, coins, distrTo, startTime, 1)
	require.NoError(t, err)

	genesis = incentives.ExportGenesis(ctx, *app.IncentivesKeeper)
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
	incentives.InitGenesis(ctx, *app.IncentivesKeeper, types.GenesisState{
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
