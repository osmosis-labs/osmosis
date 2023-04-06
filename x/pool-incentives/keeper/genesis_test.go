package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	pool_incentives "github.com/osmosis-labs/osmosis/v15/x/pool-incentives"

	simapp "github.com/osmosis-labs/osmosis/v15/app"

	"github.com/osmosis-labs/osmosis/v15/x/pool-incentives/types"
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

func (suite *KeeperTestSuite) TestExportGenesis() {
	ctx := suite.App.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	suite.App.PoolIncentivesKeeper.InitGenesis(ctx, &genesis)

	lockableDurations := suite.App.PoolIncentivesKeeper.GetLockableDurations(ctx)
	suite.App.IncentivesKeeper.SetLockableDurations(ctx, lockableDurations)
	poolId := suite.PrepareBalancerPool()

	durations := []time.Duration{
		time.Second,
		time.Minute,
		time.Hour,
	}
	suite.App.PoolIncentivesKeeper.SetLockableDurations(ctx, durations)
	savedDurations := suite.App.PoolIncentivesKeeper.GetLockableDurations(ctx)
	suite.Equal(savedDurations, durations)
	var expectedPoolToGauges types.PoolToGauges
	var gauge uint64
	for _, duration := range durations {
		gauge++
		var poolToGauge types.PoolToGauge
		poolToGauge.Duration = duration
		poolToGauge.PoolId = poolId
		poolToGauge.GaugeId = gauge
		expectedPoolToGauges.PoolToGauge = append(expectedPoolToGauges.PoolToGauge, poolToGauge)
	}

	genesisExported := suite.App.PoolIncentivesKeeper.ExportGenesis(ctx)
	suite.Equal(genesisExported.Params, genesis.Params)
	suite.Equal(genesisExported.LockableDurations, durations)
	suite.Equal(genesisExported.DistrInfo, genesis.DistrInfo)
	suite.Equal(genesisExported.PoolToGauges, &expectedPoolToGauges)
}
