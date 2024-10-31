package keeper_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	simapp "github.com/osmosis-labs/osmosis/v27/app"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	pool_incentives "github.com/osmosis-labs/osmosis/v27/x/pool-incentives"
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
)

var (
	now         = time.Now().UTC()
	testGenesis = types.GenesisState{
		Params: types.Params{
			MintedDenom: appparams.BaseCoinUnit,
		},
		LockableDurations: []time.Duration{
			time.Second,
			time.Minute,
			time.Hour,
		},
		DistrInfo: &types.DistrInfo{
			TotalWeight: osmomath.NewInt(1),
			Records: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  osmomath.NewInt(1),
				},
			},
		},
		AnyPoolToInternalGauges: &types.AnyPoolToInternalGauges{
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
			},
		},
		ConcentratedPoolToNoLockGauges: &types.ConcentratedPoolToNoLockGauges{
			PoolToGauge: []types.PoolToGauge{
				{
					PoolId:   3,
					GaugeId:  3,
					Duration: 0,
				},
			},
		},
	}
)

func TestMarshalUnmarshalGenesis(t *testing.T) {
	dirName := fmt.Sprintf("%d", rand.Int())
	app := simapp.SetupWithCustomHome(false, dirName)

	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	encodingConfig := simapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	am := pool_incentives.NewAppModule(*app.PoolIncentivesKeeper)

	genesis := testGenesis
	app.PoolIncentivesKeeper.InitGenesis(ctx, &genesis)
	assert.Equal(t, app.PoolIncentivesKeeper.GetDistrInfo(ctx), *testGenesis.DistrInfo)

	genesisExported := am.ExportGenesis(ctx, appCodec)
	os.RemoveAll(dirName)

	assert.NotPanics(t, func() {
		app := simapp.Setup(false)
		ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := pool_incentives.NewAppModule(*app.PoolIncentivesKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}

func TestInitGenesis(t *testing.T) {
	dirName := fmt.Sprintf("%d", rand.Int())
	app := simapp.SetupWithCustomHome(false, dirName)

	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	app.PoolIncentivesKeeper.InitGenesis(ctx, &genesis)

	params := app.PoolIncentivesKeeper.GetParams(ctx)
	require.Equal(t, params, genesis.Params)

	durations := app.PoolIncentivesKeeper.GetLockableDurations(ctx)
	require.Equal(t, durations, genesis.LockableDurations)

	distrInfo := app.PoolIncentivesKeeper.GetDistrInfo(ctx)
	require.Equal(t, distrInfo, *genesis.DistrInfo)

	os.RemoveAll(dirName)
}

func (s *KeeperTestSuite) TestExportGenesis() {
	ctx := s.App.BaseApp.NewContextLegacy(false, tmproto.Header{})
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
	var expectedPoolToGauges types.AnyPoolToInternalGauges
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
	s.Equal(genesisExported.AnyPoolToInternalGauges, &expectedPoolToGauges)
}

// This test validates that all store indexes are set correctly
// for NoLock gauges after exporting and then reimporting genesis.
func (s *KeeperTestSuite) TestImportExportGenesis_ExternalNoLock() {
	s.SetupTest()

	// Prepare concentrated pool
	clPool := s.PrepareConcentratedPool()

	// Fund account to create gauge
	s.FundAcc(s.TestAccs[0], defaultCoins.Add(defaultCoins...))

	// Since this test creates or adds to a gauge, we need to ensure a route exists in protorev hot routes.
	// The pool doesn't need to actually exist for this test, so we can just ensure the denom pair has some entry.
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, sdk.DefaultBondDenom, 9999)

	// Create external non-perpetual gauge
	externalGaugeID, err := s.App.IncentivesKeeper.CreateGauge(s.Ctx, false, s.TestAccs[0], defaultCoins.Add(defaultCoins...), lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.NoLock,
		Duration:      defaultNoLockDuration,
	}, s.Ctx.BlockTime(), 2, clPool.GetId())
	s.Require().NoError(err)

	// We expect internal gauge to be created first
	internalGaugeID := externalGaugeID - 1

	// Export genesis
	export := s.App.PoolIncentivesKeeper.ExportGenesis(s.Ctx)

	// Validate that only one link for internal gauges is created
	s.Require().Equal(1, len(export.AnyPoolToInternalGauges.PoolToGauge))

	// Validate that 2 links, one for external and one for internal gauge, are created
	s.Require().Equal(2, len(export.ConcentratedPoolToNoLockGauges.PoolToGauge))

	// Reset state
	s.SetupTest()

	// Import genesis
	s.App.PoolIncentivesKeeper.InitGenesis(s.Ctx, export)

	// Get the general link between external gauge ID and pool
	poolID, err := s.App.PoolIncentivesKeeper.GetPoolIdFromGaugeId(s.Ctx, externalGaugeID, 0)
	s.Require().NoError(err)
	s.Require().Equal(clPool.GetId(), poolID)

	// Get the general link between internal gauge ID and pool
	poolID, err = s.App.PoolIncentivesKeeper.GetPoolIdFromGaugeId(s.Ctx, internalGaugeID, 0)
	s.Require().NoError(err)
	s.Require().Equal(clPool.GetId(), poolID)

	// Get the internal gauge
	incentivesEpochDuration := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Duration
	internalGaugeIDAfterImport, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, poolID, incentivesEpochDuration)
	s.Require().NoError(err)
	s.Require().Equal(internalGaugeID, internalGaugeIDAfterImport)
}
