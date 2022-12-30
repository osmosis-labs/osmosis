package concentrated_liquidity_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	osmoapp "github.com/osmosis-labs/osmosis/v13/app"
	cl "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

var (
	testGenesis = types.GenesisState{
		Params: types.Params{AuthorizedTickSpacing: []uint64{1, 10, 50}},
	}
)

// TestInitGenesis tests the InitGenesis function of the ConcentratedLiquidityKeeper.
// It checks that the state is initialized correctly based on the provided genesis.
func TestInitGenesis(t *testing.T) {

	t.Skip("TODO: re-enable this when CL state-breakage PR is merged.")

	// Set up the app and context
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	now := ctx.BlockTime()
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	// Initialize the state with the provided genesis
	app.ConcentratedLiquidityKeeper.InitGenesis(ctx, testGenesis)

	// Check that the state was initialized correctly
	clParamsAfterInitialization := app.ConcentratedLiquidityKeeper.GetParams(ctx)
	require.Equal(t, testGenesis.Params.String(), clParamsAfterInitialization.String())
}

// TestExportGenesis tests the ExportGenesis function of the ConcentratedLiquidityKeeper.
// It checks that the correct genesis state is returned.
func TestExportGenesis(t *testing.T) {

	t.Skip("TODO: re-enable this when CL state-breakage PR is merged.")

	// Set up the app and context
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	now := ctx.BlockTime()
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	// Initialize the state with the provided genesis
	app.ConcentratedLiquidityKeeper.InitGenesis(ctx, testGenesis)

	// Export the genesis state and check that it is correct
	genesisExported := app.ConcentratedLiquidityKeeper.ExportGenesis(ctx)
	require.Equal(t, testGenesis.Params.String(), genesisExported.Params.String())
}

// TestMarshalUnmarshalGenesis tests the MarshalUnmarshalGenesis functions of the ConcentratedLiquidityKeeper.
// It checks that the exported genesis can be marshaled and unmarshaled without panicking.
func TestMarshalUnmarshalGenesis(t *testing.T) {

	t.Skip("TODO: re-enable this when CL state-breakage PR is merged.")

	// Set up the app and context
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	now := ctx.BlockTime()
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	// Create an app module for the ConcentratedLiquidityKeeper
	encodingConfig := osmoapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	appModule := cl.NewAppModule(appCodec, *app.ConcentratedLiquidityKeeper)

	// Export the genesis state
	genesisExported := appModule.ExportGenesis(ctx, appCodec)

	// Test that the exported genesis can be marshaled and unmarshaled without panicking
	assert.NotPanics(t, func() {
		app := osmoapp.Setup(false)
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := cl.NewAppModule(appCodec, *app.ConcentratedLiquidityKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}
