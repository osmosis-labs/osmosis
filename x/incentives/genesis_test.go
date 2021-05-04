package incentives_test

import (
	"testing"
	"time"

	simapp "github.com/c-osmosis/osmosis/app"
	"github.com/c-osmosis/osmosis/x/incentives"
	"github.com/c-osmosis/osmosis/x/incentives/types"
	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestIncentivesExportGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	genesis := incentives.ExportGenesis(ctx, app.IncentivesKeeper)
	require.Equal(t, genesis.Params.DistrEpochIdentifier, "weekly")
	require.Len(t, genesis.Pots, 0)

	addr := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	startTime := time.Now()
	app.BankKeeper.SetBalances(ctx, addr, coins)
	potID, err := app.IncentivesKeeper.CreatePot(ctx, true, addr, coins, distrTo, startTime, 1)
	require.NoError(t, err)

	genesis = incentives.ExportGenesis(ctx, app.IncentivesKeeper)
	require.Equal(t, genesis.Params.DistrEpochIdentifier, "weekly")
	require.Len(t, genesis.Pots, 1)

	require.Equal(t, genesis.Pots[0], types.Pot{
		Id:                potID,
		IsPerpetual:       true,
		DistributeTo:      distrTo,
		Coins:             coins,
		NumEpochsPaidOver: 1,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime.UTC(),
	})
}

func TestIncentivesInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	startTime := time.Now()
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	incentives.InitGenesis(ctx, app.IncentivesKeeper, types.GenesisState{
		Params: types.Params{
			DistrEpochIdentifier: "weekly",
		},
		Pots: []types.Pot{
			{
				Id:                1,
				IsPerpetual:       false,
				DistributeTo:      distrTo,
				Coins:             coins,
				NumEpochsPaidOver: 2,
				FilledEpochs:      0,
				DistributedCoins:  sdk.Coins{},
				StartTime:         startTime,
			},
		},
	})

	pots := app.IncentivesKeeper.GetPots(ctx)
	require.Len(t, pots, 1)
	require.Equal(t, pots[0], types.Pot{
		Id:                1,
		IsPerpetual:       false,
		DistributeTo:      distrTo,
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime.UTC(),
	})
}
