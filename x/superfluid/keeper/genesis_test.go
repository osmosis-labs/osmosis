package keeper_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	simapp "github.com/osmosis-labs/osmosis/v27/app"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

var now = time.Now().UTC()

var testGenesis = types.GenesisState{
	Params: types.Params{
		MinimumRiskFactor: osmomath.NewDecWithPrec(5, 1), // 50%
	},
	SuperfluidAssets: []types.SuperfluidAsset{
		{
			Denom:     DefaultGammAsset,
			AssetType: types.SuperfluidAssetTypeLPShare,
		},
	},
	OsmoEquivalentMultipliers: []types.OsmoEquivalentMultiplierRecord{
		{
			EpochNumber: 1,
			Denom:       DefaultGammAsset,
			Multiplier:  osmomath.NewDec(1000),
		},
	},
	IntermediaryAccounts: []types.SuperfluidIntermediaryAccount{
		{
			Denom:   DefaultGammAsset,
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
	dirName := fmt.Sprintf("%d", rand.Int())
	app := simapp.SetupWithCustomHome(false, dirName)

	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	encodingConfig := simapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	am := superfluid.NewAppModule(*app.SuperfluidKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.LockupKeeper, app.GAMMKeeper, app.EpochsKeeper, app.ConcentratedLiquidityKeeper)
	genesis := testGenesis
	app.SuperfluidKeeper.InitGenesis(ctx, genesis)

	genesisExported := am.ExportGenesis(ctx, appCodec)
	os.RemoveAll(dirName)

	assert.NotPanics(t, func() {
		app := simapp.Setup(false)
		ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := superfluid.NewAppModule(*app.SuperfluidKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.LockupKeeper, app.GAMMKeeper, app.EpochsKeeper, app.ConcentratedLiquidityKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}

func TestInitGenesis(t *testing.T) {
	dirName := fmt.Sprintf("%d", rand.Int())
	app := simapp.SetupWithCustomHome(false, dirName)

	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	app.SuperfluidKeeper.InitGenesis(ctx, genesis)

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
	os.RemoveAll(dirName)
}

func TestExportGenesis(t *testing.T) {
	dirName := fmt.Sprintf("%d", rand.Int())
	app := simapp.SetupWithCustomHome(false, dirName)

	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	app.SuperfluidKeeper.InitGenesis(ctx, genesis)

	asset := types.SuperfluidAsset{
		Denom:     "gamm/pool/2",
		AssetType: types.SuperfluidAssetTypeLPShare,
	}
	app.SuperfluidKeeper.SetSuperfluidAsset(ctx, asset)
	savedAsset, err := app.SuperfluidKeeper.GetSuperfluidAsset(ctx, "gamm/pool/2")
	require.NoError(t, err)
	require.Equal(t, savedAsset, asset)

	genesisExported := app.SuperfluidKeeper.ExportGenesis(ctx)
	require.Equal(t, genesisExported.Params, genesis.Params)
	require.Equal(t, genesisExported.SuperfluidAssets, append(genesis.SuperfluidAssets, asset))
	require.Equal(t, genesis.OsmoEquivalentMultipliers, genesis.OsmoEquivalentMultipliers)
	require.Equal(t, genesis.IntermediaryAccounts, genesis.IntermediaryAccounts)
	require.Equal(t, genesis.IntemediaryAccountConnections, genesis.IntemediaryAccountConnections)

	os.RemoveAll(dirName)
}
