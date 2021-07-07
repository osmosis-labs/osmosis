package pool_incentives_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/osmosis-labs/osmosis/app"
	pool_incentives "github.com/osmosis-labs/osmosis/x/pool-incentives"
	"github.com/osmosis-labs/osmosis/x/pool-incentives/types"
	"github.com/stretchr/testify/assert"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var now = time.Now().UTC()
var testGenesis = types.GenesisState{
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

func TestMarshalUnmarshalGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	encodingConfig := simapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	am := pool_incentives.NewAppModule(appCodec, app.PoolIncentivesKeeper)

	genesis := testGenesis
	app.PoolIncentivesKeeper.InitGenesis(ctx, &genesis)

	genesisExported := am.ExportGenesis(ctx, appCodec)
	assert.NotPanics(t, func() {
		app := simapp.Setup(false)
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := pool_incentives.NewAppModule(appCodec, app.PoolIncentivesKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}
