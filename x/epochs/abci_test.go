package epochs_test

import (
	"testing"
	"time"

	simapp "github.com/c-osmosis/osmosis/app"
	"github.com/c-osmosis/osmosis/x/epochs"
	"github.com/c-osmosis/osmosis/x/epochs/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestEpochInfoChangesBeginEndBlockersAndInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.EpochsKeeper.DeleteEpochInfo(ctx, "daily")
	app.EpochsKeeper.DeleteEpochInfo(ctx, "weekly")

	now := time.Now()
	ctx = ctx.WithBlockHeight(1)
	ctx = ctx.WithBlockTime(now)

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

	epochInfo := app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	require.Equal(t, epochInfo.Identifier, "monthly")
	require.Equal(t, epochInfo.StartTime.UTC().String(), now.UTC().String())
	require.Equal(t, epochInfo.Duration, time.Hour*24)
	require.Equal(t, epochInfo.CurrentEpoch, int64(0))
	require.Equal(t, epochInfo.CurrentEpochStartTime.UTC().String(), ctx.BlockTime().UTC().String())
	require.Equal(t, epochInfo.EpochCountingStarted, true)
	require.Equal(t, epochInfo.CurrentEpochEnded, true)

	// check beginblock
	ctx = ctx.WithBlockHeight(2)
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	epochs.BeginBlocker(ctx, app.EpochsKeeper)
	epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	require.Equal(t, epochInfo.StartTime.UTC().String(), now.UTC().String())
	require.Equal(t, epochInfo.Duration, time.Hour*24)
	require.Equal(t, epochInfo.CurrentEpoch, int64(1))
	require.Equal(t, epochInfo.CurrentEpochStartTime.UTC().String(), ctx.BlockTime().UTC().String())
	require.Equal(t, epochInfo.EpochCountingStarted, true)
	require.Equal(t, epochInfo.CurrentEpochEnded, false)

	// check endblock
	epochs.EndBlocker(ctx, app.EpochsKeeper)
	epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	require.Equal(t, epochInfo.StartTime.UTC().String(), now.UTC().String())
	require.Equal(t, epochInfo.Duration, time.Hour*24)
	require.Equal(t, epochInfo.CurrentEpoch, int64(1))
	require.Equal(t, epochInfo.CurrentEpochStartTime.UTC().String(), ctx.BlockTime().UTC().String())
	require.Equal(t, epochInfo.EpochCountingStarted, true)
	require.Equal(t, epochInfo.CurrentEpochEnded, false)

	// check beginblock
	ctx = ctx.WithBlockHeight(3)
	ctx = ctx.WithBlockTime(now.Add(time.Hour * 24 * 31))
	epochs.BeginBlocker(ctx, app.EpochsKeeper)
	epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	require.Equal(t, epochInfo.StartTime.UTC().String(), now.UTC().String())
	require.Equal(t, epochInfo.Duration, time.Hour*24)
	require.Equal(t, epochInfo.CurrentEpoch, int64(1))
	require.Equal(t, epochInfo.CurrentEpochStartTime.UTC().String(), now.Add(time.Second).UTC().String())
	require.Equal(t, epochInfo.EpochCountingStarted, true)
	require.Equal(t, epochInfo.CurrentEpochEnded, false)

	// check endblock
	epochs.EndBlocker(ctx, app.EpochsKeeper)
	epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	require.Equal(t, epochInfo.StartTime.UTC().String(), now.UTC().String())
	require.Equal(t, epochInfo.Duration, time.Hour*24)
	require.Equal(t, epochInfo.CurrentEpoch, int64(1))
	require.Equal(t, epochInfo.CurrentEpochStartTime.UTC().String(), now.Add(time.Second).UTC().String())
	require.Equal(t, epochInfo.EpochCountingStarted, true)
	require.Equal(t, epochInfo.CurrentEpochEnded, true)

	// check beginblock
	ctx = ctx.WithBlockHeight(4)
	ctx = ctx.WithBlockTime(now.Add(time.Hour * 24 * 32))
	epochs.BeginBlocker(ctx, app.EpochsKeeper)
	epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	require.Equal(t, epochInfo.StartTime.UTC().String(), now.UTC().String())
	require.Equal(t, epochInfo.Duration, time.Hour*24)
	require.Equal(t, epochInfo.CurrentEpoch, int64(2))
	require.Equal(t, epochInfo.CurrentEpochStartTime.UTC().String(), ctx.BlockTime().UTC().String())
	require.Equal(t, epochInfo.EpochCountingStarted, true)
	require.Equal(t, epochInfo.CurrentEpochEnded, false)

	// check endblock
	epochs.EndBlocker(ctx, app.EpochsKeeper)
	epochInfo = app.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	require.Equal(t, epochInfo.StartTime.UTC().String(), now.UTC().String())
	require.Equal(t, epochInfo.Duration, time.Hour*24)
	require.Equal(t, epochInfo.CurrentEpoch, int64(2))
	require.Equal(t, epochInfo.CurrentEpochStartTime.UTC().String(), ctx.BlockTime().UTC().String())
	require.Equal(t, epochInfo.EpochCountingStarted, true)
	require.Equal(t, epochInfo.CurrentEpochEnded, false)
}
