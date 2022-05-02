package superfluid_test

import (
	"testing"
	"time"

	simapp "github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var now = time.Now().UTC()

var testGenesis = types.GenesisState{
	Params: types.Params{
		MinimumRiskFactor: sdk.NewDecWithPrec(5, 1), // 50%
	},
	SuperfluidAssets: []types.SuperfluidAsset{
		{
			Denom:     "gamm/pool/1",
			AssetType: types.SuperfluidAssetTypeLPShare,
		},
	},
	OsmoEquivalentMultipliers: []types.OsmoEquivalentMultiplierRecord{
		{
			EpochNumber: 1,
			Denom:       "gamm/pool/1",
			Multiplier:  sdk.NewDec(1000),
		},
	},
	IntermediaryAccounts: []types.SuperfluidIntermediaryAccount{
		{
			Denom:   "gamm/pool/1",
			ValAddr: "osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n",
			GaugeId: 1,
		},
	},
	IntemediaryAccountConnections: []types.LockIdIntermediaryAccountConnection{
		{
			LockId:              1,
			IntermediaryAccount: "osmo1hpgapnfl3thkevvl0jp3wqtk8jw7mpqumuuc2f",
		},
	},
}

func TestMarshalUnmarshalGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	encodingConfig := simapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	am := superfluid.NewAppModule(appCodec, *app.SuperfluidKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.LockupKeeper, app.GAMMKeeper, app.EpochsKeeper)

	genesis := testGenesis
	superfluid.InitGenesis(ctx, *app.SuperfluidKeeper, genesis)

	genesisExported := am.ExportGenesis(ctx, appCodec)
	assert.NotPanics(t, func() {
		app := simapp.Setup(false)
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := superfluid.NewAppModule(appCodec, *app.SuperfluidKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.LockupKeeper, app.GAMMKeeper, app.EpochsKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}

func TestInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	superfluid.InitGenesis(ctx, *app.SuperfluidKeeper, genesis)

	params := app.SuperfluidKeeper.GetParams(ctx)
	require.Equal(t, params, genesis.Params)

	assets := app.SuperfluidKeeper.GetAllSuperfluidAssets(ctx)
	require.Equal(t, assets, genesis.SuperfluidAssets)

	multipliers := app.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(ctx)
	require.Equal(t, multipliers, genesis.OsmoEquivalentMultipliers)

	accounts := app.SuperfluidKeeper.GetAllIntermediaryAccounts(ctx)
	require.Equal(t, accounts, genesis.IntermediaryAccounts)

	connections := app.SuperfluidKeeper.GetAllLockIdIntermediaryAccountConnections(ctx)
	require.Equal(t, connections, genesis.IntemediaryAccountConnections)
}

func TestExportGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	superfluid.InitGenesis(ctx, *app.SuperfluidKeeper, genesis)

	asset := types.SuperfluidAsset{
		Denom:     "gamm/pool/2",
		AssetType: types.SuperfluidAssetTypeLPShare,
	}
	app.SuperfluidKeeper.SetSuperfluidAsset(ctx, asset)
	savedAsset := app.SuperfluidKeeper.GetSuperfluidAsset(ctx, "gamm/pool/2")
	require.Equal(t, savedAsset, asset)

	genesisExported := superfluid.ExportGenesis(ctx, *app.SuperfluidKeeper)
	require.Equal(t, genesisExported.Params, genesis.Params)
	require.Equal(t, genesisExported.SuperfluidAssets, append(genesis.SuperfluidAssets, asset))
	require.Equal(t, genesis.OsmoEquivalentMultipliers, genesis.OsmoEquivalentMultipliers)
	require.Equal(t, genesis.IntermediaryAccounts, genesis.IntermediaryAccounts)
	require.Equal(t, genesis.IntemediaryAccountConnections, genesis.IntemediaryAccountConnections)
}
