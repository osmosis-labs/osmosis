package epochs_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/osmosis-labs/osmosis/app"
	"github.com/osmosis-labs/osmosis/x/epochs"
	"github.com/osmosis-labs/osmosis/x/epochs/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestEpochInfoChangesBeginEndBlockersAndInitGenesis(t *testing.T) {
	var app *simapp.OsmosisApp
	var ctx sdk.Context
	var epochInfo types.EpochInfo

	now := time.Now()

	tests := []struct {
		expCurrentEpochStartTime time.Time
		expCurrentEpoch          int64
		expCurrentEpochEnded     bool
		fn                       func()
	}{
		{
			expCurrentEpochStartTime: now,
			expCurrentEpoch:          0,
			expCurrentEpochEnded:     true,
			fn: func() {
				epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		{
			expCurrentEpochStartTime: now.Add(time.Second),
			expCurrentEpoch:          1,
			expCurrentEpochEnded:     false,
			fn: func() {
				ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				epochs.BeginBlocker(ctx, app.EpochsKeeper)
				epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		{
			expCurrentEpochStartTime: now.Add(time.Second),
			expCurrentEpoch:          1,
			expCurrentEpochEnded:     false,
			fn: func() {
				ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				epochs.BeginBlocker(ctx, app.EpochsKeeper)
				epochs.EndBlocker(ctx, app.EpochsKeeper)
				epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		{
			expCurrentEpochStartTime: now.Add(time.Second),
			expCurrentEpoch:          1,
			expCurrentEpochEnded:     false,
			fn: func() {
				ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				epochs.BeginBlocker(ctx, app.EpochsKeeper)
				epochs.EndBlocker(ctx, app.EpochsKeeper)
				ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 31))
				epochs.BeginBlocker(ctx, app.EpochsKeeper)
				epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		{
			expCurrentEpochStartTime: now.Add(time.Second),
			expCurrentEpoch:          1,
			expCurrentEpochEnded:     true,
			fn: func() {
				ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				epochs.BeginBlocker(ctx, app.EpochsKeeper)
				epochs.EndBlocker(ctx, app.EpochsKeeper)
				ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 31))
				epochs.BeginBlocker(ctx, app.EpochsKeeper)
				epochs.EndBlocker(ctx, app.EpochsKeeper)
				epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		{
			expCurrentEpochStartTime: now.Add(time.Hour * 24 * 31),
			expCurrentEpoch:          2,
			expCurrentEpochEnded:     false,
			fn: func() {
				ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				epochs.BeginBlocker(ctx, app.EpochsKeeper)
				epochs.EndBlocker(ctx, app.EpochsKeeper)
				ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 31))
				epochs.BeginBlocker(ctx, app.EpochsKeeper)
				epochs.EndBlocker(ctx, app.EpochsKeeper)
				ctx.WithBlockHeight(4).WithBlockTime(now.Add(time.Hour * 24 * 32))
				epochs.BeginBlocker(ctx, app.EpochsKeeper)
				epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		{
			expCurrentEpochStartTime: now.Add(time.Hour * 24 * 31),
			expCurrentEpoch:          2,
			expCurrentEpochEnded:     false,
			fn: func() {
				ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				epochs.BeginBlocker(ctx, app.EpochsKeeper)
				epochs.EndBlocker(ctx, app.EpochsKeeper)
				ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 31))
				epochs.BeginBlocker(ctx, app.EpochsKeeper)
				epochs.EndBlocker(ctx, app.EpochsKeeper)
				ctx.WithBlockHeight(4).WithBlockTime(now.Add(time.Hour * 24 * 32))
				epochs.BeginBlocker(ctx, app.EpochsKeeper)
				epochs.EndBlocker(ctx, app.EpochsKeeper)
				epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
	}

	for _, test := range tests {
		app = simapp.Setup(false)
		ctx = app.BaseApp.NewContext(false, tmproto.Header{})

		// On init genesis, default epochs information is set
		// To check init genesis again, should make it fresh status
		epochInfos := app.EpochsKeeper.AllEpochInfos(ctx)
		for _, epochInfo := range epochInfos {
			app.EpochsKeeper.DeleteEpochInfo(ctx, epochInfo.Identifier)
		}

		ctx = ctx.WithBlockHeight(1).WithBlockTime(now)

		// check init genesis
		epochs.InitGenesis(ctx, app.EpochsKeeper, types.GenesisState{
			Epochs: []types.EpochInfo{
				{
					Identifier:            "monthly",
					StartTime:             time.Time{},
					Duration:              time.Hour * 24,
					CurrentEpoch:          0,
					CurrentEpochStartTime: time.Time{},
					EpochCountingStarted:  true,
					CurrentEpochEnded:     true,
				},
			},
		})

		test.fn()

		require.Equal(t, epochInfo.Identifier, "monthly")
		require.Equal(t, epochInfo.StartTime.UTC().String(), now.UTC().String())
		require.Equal(t, epochInfo.Duration, time.Hour*24)
		require.Equal(t, epochInfo.CurrentEpoch, test.expCurrentEpoch)
		require.Equal(t, epochInfo.CurrentEpochStartTime.UTC().String(), test.expCurrentEpochStartTime.UTC().String())
		require.Equal(t, epochInfo.EpochCountingStarted, true)
		require.Equal(t, epochInfo.CurrentEpochEnded, test.expCurrentEpochEnded)
	}
}

func TestEpochStartingOneMonthAfterInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// On init genesis, default epochs information is set
	// To check init genesis again, should make it fresh status
	epochInfos := app.EpochsKeeper.AllEpochInfos(ctx)
	for _, epochInfo := range epochInfos {
		app.EpochsKeeper.DeleteEpochInfo(ctx, epochInfo.Identifier)
	}

	now := time.Now()
	week := time.Hour * 24 * 7
	month := time.Hour * 24 * 30
	ctx = ctx.WithBlockHeight(1).WithBlockTime(now)

	epochs.InitGenesis(ctx, app.EpochsKeeper, types.GenesisState{
		Epochs: []types.EpochInfo{
			{
				Identifier:            "monthly",
				StartTime:             now.Add(month),
				Duration:              time.Hour * 24 * 30,
				CurrentEpoch:          0,
				CurrentEpochStartTime: time.Time{},
				EpochCountingStarted:  false,
				CurrentEpochEnded:     true,
			},
		},
	})

	// epoch not started yet
	epochInfo := app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	require.Equal(t, epochInfo.CurrentEpochStartTime, time.Time{})
	require.Equal(t, epochInfo.EpochCountingStarted, false)
	require.Equal(t, epochInfo.CurrentEpochEnded, true)

	// after 1 week
	ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(week))
	epochs.BeginBlocker(ctx, app.EpochsKeeper)
	epochs.EndBlocker(ctx, app.EpochsKeeper)

	// epoch not started yet
	epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	require.Equal(t, epochInfo.CurrentEpochStartTime, time.Time{})
	require.Equal(t, epochInfo.EpochCountingStarted, false)
	require.Equal(t, epochInfo.CurrentEpochEnded, true)

	// after 1 month
	ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(month))
	epochs.BeginBlocker(ctx, app.EpochsKeeper)
	epochs.EndBlocker(ctx, app.EpochsKeeper)

	// epoch started
	epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	require.Equal(t, epochInfo.CurrentEpochStartTime.UTC().String(), now.Add(month).UTC().String())
	require.Equal(t, epochInfo.EpochCountingStarted, true)
	require.Equal(t, epochInfo.CurrentEpochEnded, false)
}
