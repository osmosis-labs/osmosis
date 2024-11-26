package keeper_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	osmoapp "github.com/osmosis-labs/osmosis/v27/app"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"

	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

var (
	distrToByDuration = lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}

	distrToNoLock = lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.NoLock,
		Duration:      defaultNoLockDuration,
	}

	distrToNoLockPool1 = lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.NoLock,
		Denom:         "no-lock/i/1",
		Duration:      time.Hour * 24 * 7,
		Timestamp:     time.Time{}.UTC(),
	}

	distrToNoLockPool2 = lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.NoLock,
		Denom:         "no-lock/i/2",
		Duration:      time.Hour * 24 * 7,
		Timestamp:     time.Time{}.UTC(),
	}

	distrToByGroup = lockuptypes.QueryCondition{LockQueryType: lockuptypes.ByGroup}

	gaugeCoins = sdk.Coins{sdk.NewInt64Coin("stake", 10000)}

	gaugeOneRecord = types.InternalGaugeRecord{
		GaugeId:          1,
		CurrentWeight:    osmomath.NewInt(100),
		CumulativeWeight: osmomath.NewInt(200),
	}

	gaugeTwoRecord = types.InternalGaugeRecord{
		GaugeId:          2,
		CurrentWeight:    osmomath.NewInt(100),
		CumulativeWeight: osmomath.NewInt(200),
	}

	expectedGroups = []types.Group{
		{
			GroupGaugeId: 5,
			InternalGaugeInfo: types.InternalGaugeInfo{
				TotalWeight:  gaugeOneRecord.CurrentWeight.Add(gaugeTwoRecord.CurrentWeight),
				GaugeRecords: []types.InternalGaugeRecord{gaugeOneRecord, gaugeTwoRecord},
			},
			SplittingPolicy: types.ByVolume,
		},
	}

	expectedGroupGauges = []types.Gauge{
		{
			Id:                5,
			IsPerpetual:       true,
			DistributeTo:      distrToByGroup,
			Coins:             sdk.Coins(nil),
			NumEpochsPaidOver: 0,
			FilledEpochs:      0,
			DistributedCoins:  sdk.Coins(nil),
			StartTime:         time.Time{}.UTC(),
		},
	}
)

// TestIncentivesExportGenesis tests export genesis command for the incentives module.
func TestIncentivesExportGenesis(t *testing.T) {
	// export genesis using default configurations
	// ensure resulting genesis params match default params
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
	genesis := app.IncentivesKeeper.ExportGenesis(ctx)
	require.Equal(t, genesis.Params.DistrEpochIdentifier, "week")
	require.Len(t, genesis.Gauges, 0)

	// create an address and fund with coins
	addr := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 20000), sdk.NewInt64Coin(appparams.BaseCoinUnit, 10000000000)}
	err := testutil.FundAccount(ctx, app.BankKeeper, addr, coins)
	require.NoError(t, err)

	// allow pool creation
	clParams := app.ConcentratedLiquidityKeeper.GetParams(ctx)
	clParams.IsPermissionlessPoolCreationEnabled = true
	app.ConcentratedLiquidityKeeper.SetParams(ctx, clParams)

	// create two pools to be used for group creation
	msgCreatePool := model.MsgCreateConcentratedPool{
		Sender:       addr.String(),
		Denom0:       "uion",
		Denom1:       appparams.BaseCoinUnit,
		TickSpacing:  100,
		SpreadFactor: osmomath.MustNewDecFromStr("0.0005"),
	}
	_, err = app.PoolManagerKeeper.CreatePool(ctx, msgCreatePool)
	require.NoError(t, err)
	_, err = app.PoolManagerKeeper.CreatePool(ctx, msgCreatePool)
	require.NoError(t, err)

	// mints LP tokens and send to address created earlier
	// this ensures the supply exists on chain
	mintLPtokens := sdk.Coins{sdk.NewInt64Coin(distrToByDuration.Denom, 200)}
	err = testutil.FundAccount(ctx, app.BankKeeper, addr, mintLPtokens)
	require.NoError(t, err)

	// create a gauge of every type (byDuration, noLock, byGroup)
	startTime := time.Now()
	gaugeCoins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	createAllGaugeTypes(t, app, ctx, addr, gaugeCoins, startTime)

	// directly modify the weights of the groups so we can see if non zero values persist
	groups, err := app.IncentivesKeeper.GetAllGroups(ctx)
	require.NoError(t, err)
	groups[0].InternalGaugeInfo = expectedGroups[0].InternalGaugeInfo
	app.IncentivesKeeper.SetGroup(ctx, groups[0])

	// export genesis using default configurations
	// ensure resulting genesis params match default params
	genesis = app.IncentivesKeeper.ExportGenesis(ctx)
	require.Equal(t, genesis.Params.DistrEpochIdentifier, "week")

	// check that the gauges created earlier were exported through exportGenesis and still exists on chain
	expectedGauges := expectedGauges(startTime)
	require.Len(t, genesis.Gauges, len(expectedGauges))

	// pool 1 gauge
	require.Equal(t, expectedGauges[0], genesis.Gauges[0])

	// pool 2 gauge
	require.Equal(t, expectedGauges[1], genesis.Gauges[1])

	// duration gauge
	require.Equal(t, expectedGauges[2], genesis.Gauges[2])

	// no lock gauge
	// distrToNoLock denom gets added post creation
	expectedGauges[3].DistributeTo.Denom = "no-lock/e/1"
	require.Equal(t, expectedGauges[3], genesis.Gauges[3])

	// check that the group gauges created earlier were exported through exportGenesis and still exists on chain
	require.Len(t, genesis.GroupGauges, len(expectedGroupGauges))
	require.Equal(t, expectedGroupGauges, genesis.GroupGauges)

	// check group
	require.Len(t, genesis.Groups, 1)
	require.Equal(t, expectedGroups, genesis.Groups)
}

// TestIncentivesInitGenesis takes a genesis state and tests initializing that genesis for the incentives module.
func TestIncentivesInitGenesis(t *testing.T) {
	dirName := fmt.Sprintf("%d", rand.Int())
	app := osmoapp.SetupWithCustomHome(false, dirName)

	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})

	// checks that the default genesis parameters pass validation
	validateGenesis := types.DefaultGenesis().Params.Validate()
	require.NoError(t, validateGenesis)

	startTime := time.Now()
	expectedGauges := expectedGauges(startTime)
	// distrToNoLock denom gets added post creation when being called via createGauge, but
	// we are manually creating the gauges here so we need to add it manually
	expectedGauges[3].DistributeTo.Denom = "no-lock/e/1"

	// initialize genesis with specified parameter, the gauge created earlier, and lockable durations
	app.IncentivesKeeper.InitGenesis(ctx, types.GenesisState{
		Params: types.Params{
			DistrEpochIdentifier: "week",
			InternalUptime:       types.DefaultConcentratedUptime,
		},
		Gauges: expectedGauges,
		LockableDurations: []time.Duration{
			time.Second,
			time.Hour,
			time.Hour * 3,
			time.Hour * 7,
		},
		GroupGauges: expectedGroupGauges,
		Groups:      expectedGroups,
	})

	// check that the gauge created earlier was initialized through initGenesis and still exists on chain
	gauges := app.IncentivesKeeper.GetGauges(ctx)
	require.Len(t, gauges, len(expectedGauges))

	// duration gauge
	require.Equal(t, expectedGauges[2], gauges[0])

	// no lock gauge
	require.Equal(t, expectedGauges[3], gauges[1])

	// pool 1 gauge
	require.Equal(t, expectedGauges[0], gauges[2])

	// pool 2 gauge
	require.Equal(t, expectedGauges[1], gauges[3])

	// group gauge
	groupGauges, err := app.IncentivesKeeper.GetAllGroupsGauges(ctx)
	require.NoError(t, err)
	require.Len(t, groupGauges, len(expectedGroupGauges))
	require.Equal(t, expectedGroupGauges, groupGauges)

	// check group
	groups, err := app.IncentivesKeeper.GetAllGroups(ctx)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	require.Equal(t, expectedGroups, groups)

	os.RemoveAll(dirName)
}

func createAllGaugeTypes(t *testing.T, app *osmoapp.OsmosisApp, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins, startTime time.Time) {
	for _, coin := range coins {
		app.ProtoRevKeeper.SetPoolForDenomPair(ctx, appparams.BaseCoinUnit, coin.Denom, 9999)
	}
	// create a byDuration gauge
	_, err := app.IncentivesKeeper.CreateGauge(ctx, true, addr, coins, distrToByDuration, startTime, 1, 0)
	require.NoError(t, err)

	// create a noLock gauge
	_, err = app.IncentivesKeeper.CreateGauge(ctx, false, addr, coins, distrToNoLock, startTime, 1, 1)
	require.NoError(t, err)

	// create a group which in turn creates a byGroup gauge
	// we must set volume for each of the pools in the group
	// so that the group gauge can be created.
	stakingParams, err := app.StakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	bondDenom := stakingParams.BondDenom
	volumeCoins := sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 100000000000))
	groupPoolIDs := []uint64{1, 2}
	for _, poolID := range groupPoolIDs {
		app.PoolManagerKeeper.SetVolume(ctx, poolID, volumeCoins)
	}
	_, err = app.IncentivesKeeper.CreateGroup(ctx, sdk.Coins{}, 0, addr, groupPoolIDs)
	require.NoError(t, err)
}

func expectedGauges(startTime time.Time) []types.Gauge {
	return []types.Gauge{
		{
			Id:                1,
			IsPerpetual:       true,
			DistributeTo:      distrToNoLockPool1,
			Coins:             sdk.Coins(nil),
			NumEpochsPaidOver: 1,
			FilledEpochs:      0,
			DistributedCoins:  sdk.Coins(nil),
			StartTime:         time.Time{}.UTC(),
		},
		{
			Id:                2,
			IsPerpetual:       true,
			DistributeTo:      distrToNoLockPool2,
			Coins:             sdk.Coins(nil),
			NumEpochsPaidOver: 1,
			FilledEpochs:      0,
			DistributedCoins:  sdk.Coins(nil),
			StartTime:         time.Time{}.UTC(),
		},
		{
			Id:                3,
			IsPerpetual:       true,
			DistributeTo:      distrToByDuration,
			Coins:             gaugeCoins,
			NumEpochsPaidOver: 1,
			FilledEpochs:      0,
			DistributedCoins:  sdk.Coins(nil),
			StartTime:         startTime.UTC(),
		},
		{
			Id:                4,
			IsPerpetual:       false,
			DistributeTo:      distrToNoLock,
			Coins:             gaugeCoins,
			NumEpochsPaidOver: 1,
			FilledEpochs:      0,
			DistributedCoins:  sdk.Coins(nil),
			StartTime:         startTime.UTC(),
		},
	}
}
