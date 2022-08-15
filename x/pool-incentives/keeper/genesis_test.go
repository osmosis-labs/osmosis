package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	pool_incentives "github.com/osmosis-labs/osmosis/v11/x/pool-incentives"

	simapp "github.com/osmosis-labs/osmosis/v11/app"

	"github.com/osmosis-labs/osmosis/v11/x/pool-incentives/types"
)

var (
	now         = time.Now().UTC()
	testGenesis = types.GenesisState{
		Params: types.Params{
			MintedDenom: "uosmo",
		},
		LockableDurations: []time.Duration{
			time.Second,
			time.Minute,
			time.Hour,
		},
		DistrInfo: &types.DistrInfo{
			TotalWeight: sdk.NewInt(1),
			Records: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(1),
				},
			},
		},
	}
)

func TestMarshalUnmarshalGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	encodingConfig := simapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	am := pool_incentives.NewAppModule(*app.PoolIncentivesKeeper)

	genesis := testGenesis
	app.PoolIncentivesKeeper.InitGenesis(ctx, &genesis)
	assert.Equal(t, app.PoolIncentivesKeeper.GetDistrInfo(ctx), *testGenesis.DistrInfo)

	genesisExported := am.ExportGenesis(ctx, appCodec)
	assert.NotPanics(t, func() {
		app := simapp.Setup(false)
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := pool_incentives.NewAppModule(*app.PoolIncentivesKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}

func TestInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	app.PoolIncentivesKeeper.InitGenesis(ctx, &genesis)

	params := app.PoolIncentivesKeeper.GetParams(ctx)
	require.Equal(t, params, genesis.Params)

	durations := app.PoolIncentivesKeeper.GetLockableDurations(ctx)
	require.Equal(t, durations, genesis.LockableDurations)

	distrInfo := app.PoolIncentivesKeeper.GetDistrInfo(ctx)
	require.Equal(t, distrInfo, *genesis.DistrInfo)
}

func TestExportGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	app.PoolIncentivesKeeper.InitGenesis(ctx, &genesis)

	durations := []time.Duration{
		time.Second,
		time.Minute,
		time.Hour,
		time.Hour * 5,
	}
	app.PoolIncentivesKeeper.SetLockableDurations(ctx, durations)
	savedDurations := app.PoolIncentivesKeeper.GetLockableDurations(ctx)
	require.Equal(t, savedDurations, durations)

	genesisExported := app.PoolIncentivesKeeper.ExportGenesis(ctx)
	require.Equal(t, genesisExported.Params, genesis.Params)
	require.Equal(t, genesisExported.LockableDurations, durations)
	require.Equal(t, genesisExported.DistrInfo, genesis.DistrInfo)
}
