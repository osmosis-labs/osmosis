package concentrated_liquidity_test

import (
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	osmoapp "github.com/osmosis-labs/osmosis/v15/app"
	clmodule "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/clmodule"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var (
	testGenesisPools = []model.Pool{}
	testGenesis      = types.GenesisState{
		Params: types.Params{
			AuthorizedTickSpacing: []uint64{1, 10, 50},
			AuthorizedSwapFees:    []sdk.Dec{sdk.MustNewDecFromStr("0.0001"), sdk.MustNewDecFromStr("0.0003"), sdk.MustNewDecFromStr("0.0005")}},
		Pools: []*codectypes.Any{},
	}
)

func init() {
	pool1, err := model.NewConcentratedLiquidityPool(1, "uosmo", "uatom", 5, sdk.NewInt(-4), DefaultZeroSwapFee)
	if err != nil {
		panic(err)
	}
	testGenesisPools = append(testGenesisPools, pool1)
	pool2, err := model.NewConcentratedLiquidityPool(7, "uusdc", "uatom", 4, sdk.NewInt(-2), sdk.MustNewDecFromStr("0.01"))
	if err != nil {
		panic(err)
	}
	testGenesisPools = append(testGenesisPools, pool2)
	for _, pool := range testGenesisPools {
		poolCopy := pool
		poolAny, err := codectypes.NewAnyWithValue(&poolCopy)
		if err != nil {
			panic(err)
		}
		testGenesis.Pools = append(testGenesis.Pools, poolAny)
	}
}

// TestInitGenesis tests the InitGenesis function of the ConcentratedLiquidityKeeper.
// It checks that the state is initialized correctly based on the provided genesis.
func TestInitGenesis(t *testing.T) {
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
	clPoolsAfterInitialization, err := app.ConcentratedLiquidityKeeper.GetAllPools(ctx)
	require.NoError(t, err)
	require.Equal(t, len(clPoolsAfterInitialization), 2)
	for i := 0; i < len(clPoolsAfterInitialization); i++ {
		require.Equal(t, &testGenesisPools[i], clPoolsAfterInitialization[i])
	}
}

// TestExportGenesis tests the ExportGenesis function of the ConcentratedLiquidityKeeper.
// It checks that the correct genesis state is returned.
func TestExportGenesis(t *testing.T) {
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
	require.Len(t, genesisExported.Pools, 2)
	require.Equal(t, testGenesis.Pools, genesisExported.Pools)
}

// TestMarshalUnmarshalGenesis tests the MarshalUnmarshalGenesis functions of the ConcentratedLiquidityKeeper.
// It checks that the exported genesis can be marshaled and unmarshaled without panicking.
func TestMarshalUnmarshalGenesis(t *testing.T) {
	// Set up the app and context
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	now := ctx.BlockTime()
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	// Create an app module for the ConcentratedLiquidityKeeper
	encodingConfig := osmoapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	appModule := clmodule.NewAppModule(appCodec, *app.ConcentratedLiquidityKeeper)

	// Export the genesis state
	genesisExported := appModule.ExportGenesis(ctx, appCodec)

	// Test that the exported genesis can be marshaled and unmarshaled without panicking
	assert.NotPanics(t, func() {
		app := osmoapp.Setup(false)
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := clmodule.NewAppModule(appCodec, *app.ConcentratedLiquidityKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}
