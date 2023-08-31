package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	pool_incentives "github.com/osmosis-labs/osmosis/v19/x/pool-incentives"

	simapp "github.com/osmosis-labs/osmosis/v19/app"

	"github.com/osmosis-labs/osmosis/v19/x/pool-incentives/types"
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
		PoolToGauges: &types.PoolToGauges{
			PoolToGauge: []types.PoolToGauge{
				{
					PoolId:   1,
					GaugeId:  1,
					Duration: time.Second,
				},
				{
					PoolId:   2,
					GaugeId:  2,
					Duration: time.Second,
				},
				// This duplication with zero duration
				// can happen with "NoLock" gauges
				// where the link containing the duration
				// is used to signify that the gauge is internal
				// while the link without the duration is used
				// for general purpose. This redundancy is
				// made for convinience of plugging in the
				// later added "NoLock" gauge into the existing
				// logic without having to change majority of the queries.
				{
					PoolId:  2,
					GaugeId: 2,
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

func (s *KeeperTestSuite) TestExportGenesis() {
	ctx := s.App.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	s.App.PoolIncentivesKeeper.InitGenesis(ctx, &genesis)

	lockableDurations := s.App.PoolIncentivesKeeper.GetLockableDurations(ctx)
	s.App.IncentivesKeeper.SetLockableDurations(ctx, lockableDurations)
	poolId := s.PrepareBalancerPool()

	durations := []time.Duration{
		time.Second,
		time.Minute,
		time.Hour,
	}
	s.App.PoolIncentivesKeeper.SetLockableDurations(ctx, durations)
	savedDurations := s.App.PoolIncentivesKeeper.GetLockableDurations(ctx)
	s.Equal(savedDurations, durations)
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

	genesisExported := s.App.PoolIncentivesKeeper.ExportGenesis(ctx)
	s.Equal(genesisExported.Params, genesis.Params)
	s.Equal(genesisExported.LockableDurations, durations)
	s.Equal(genesisExported.DistrInfo, genesis.DistrInfo)
	s.Equal(genesisExported.PoolToGauges, &expectedPoolToGauges)
}
