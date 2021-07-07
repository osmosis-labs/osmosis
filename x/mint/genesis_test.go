package mint_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/osmosis-labs/osmosis/app"
	"github.com/osmosis-labs/osmosis/x/mint"
	"github.com/osmosis-labs/osmosis/x/mint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var now = time.Now().UTC()
var acc1 = sdk.AccAddress([]byte("addr1---------------"))
var testGenesis = types.GenesisState{
	Minter: types.Minter{
		EpochProvisions: sdk.NewDec(100000000),
	},
	Params: types.Params{
		MintDenom:               "uosmo",
		GenesisEpochProvisions:  sdk.NewDec(100000000),
		EpochIdentifier:         "week",
		ReductionPeriodInEpochs: 100,
		ReductionFactor:         sdk.NewDecWithPrec(5, 1),
		DistributionProportions: types.DefaultParams().DistributionProportions,
		WeightedDeveloperRewardsReceivers: []types.WeightedAddress{
			{
				Address: acc1.String(),
				Weight:  sdk.NewDec(1),
			},
		},
		MintingRewardsDistributionStartEpoch: types.DefaultParams().MintingRewardsDistributionStartEpoch,
	},
	HalvenStartedEpoch: 0,
}

func TestInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	mint.InitGenesis(ctx, app.MintKeeper, app.AccountKeeper, &genesis)

	minter := app.MintKeeper.GetMinter(ctx)
	require.Equal(t, minter, genesis.Minter)

	params := app.MintKeeper.GetParams(ctx)
	require.Equal(t, params, genesis.Params)

	halvenStartedEpoch := app.MintKeeper.GetLastHalvenEpochNum(ctx)
	require.Equal(t, halvenStartedEpoch, genesis.HalvenStartedEpoch)
}

func TestExportGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	mint.InitGenesis(ctx, app.MintKeeper, app.AccountKeeper, &genesis)

	app.MintKeeper.SetLastHalvenEpochNum(ctx, 10)

	genesisExported := mint.ExportGenesis(ctx, app.MintKeeper)
	require.Equal(t, genesisExported.Params, genesis.Params)
	require.Equal(t, genesisExported.Minter, genesis.Minter)
	require.Equal(t, genesisExported.HalvenStartedEpoch, int64(10))
}

func TestMarshalUnmarshalGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	encodingConfig := simapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	am := mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper)

	genesis := testGenesis
	mint.InitGenesis(ctx, app.MintKeeper, app.AccountKeeper, &genesis)

	app.MintKeeper.SetLastHalvenEpochNum(ctx, 10)

	genesisExported := am.ExportGenesis(ctx, appCodec)
	assert.NotPanics(t, func() {
		app := simapp.Setup(false)
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}
