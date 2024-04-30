package txfees_test

import (
	"encoding/json"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	osmosisapp "github.com/osmosis-labs/osmosis/v25/app"

	simapp "github.com/osmosis-labs/osmosis/v25/app"
	mempool1559 "github.com/osmosis-labs/osmosis/v25/x/txfees/keeper/mempool-1559"
)

func TestSetBaseDenomOnInitBlock(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	genesisState := osmosisapp.GenesisStateWithValSet(app)
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(err)
	}

	app.InitChain(
		abcitypes.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: sims.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
			ChainId:         "osmosis-1",
		},
	)

	baseDenom, err := app.TxFeesKeeper.GetBaseDenom(ctx)
	require.Nil(t, err)
	require.NotEmpty(t, baseDenom)
}

func TestBeginBlock(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "osmosis-1", Height: 1})

	genesisState := osmosisapp.GenesisStateWithValSet(app)
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(err)
	}

	app.InitChain(
		abcitypes.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: sims.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
			ChainId:         "osmosis-1",
		},
	)

	// Begin block hasn't happened yet, target gas should be equal to hard coded default value
	hardCodedGasTarget := int64(187_500_000)
	require.Equal(t, hardCodedGasTarget, mempool1559.TargetGas)

	// Run begin block
	ctx = RunBeginBlock(ctx, app)

	// Target gas should be updated to the value set in InitChain
	defaultBlockTargetGas := mempool1559.TargetBlockSpacePercent.Mul(sdk.NewDec(sims.DefaultConsensusParams.Block.MaxGas)).TruncateInt().Int64()
	require.Equal(t, defaultBlockTargetGas, mempool1559.TargetGas)

	// Run begin block again, should not update target gas
	ctx = RunBeginBlock(ctx, app)
	require.Equal(t, defaultBlockTargetGas, mempool1559.TargetGas)

	// Update the consensus params
	newDefaultBlockMaxGas := int64(300_000_000)
	newConsensusParams := *sims.DefaultConsensusParams
	newConsensusParams.Block.MaxGas = newDefaultBlockMaxGas
	app.ConsensusParamsKeeper.Set(ctx, &newConsensusParams)

	// Ensure that the consensus params have not been updated yet
	require.Equal(t, defaultBlockTargetGas, mempool1559.TargetGas)

	// Run begin block again, should update target gas
	RunBeginBlock(ctx, app)
	expectedNewBlockTargetGas := mempool1559.TargetBlockSpacePercent.Mul(sdk.NewDec(newDefaultBlockMaxGas)).TruncateInt().Int64()
	require.Equal(t, expectedNewBlockTargetGas, mempool1559.TargetGas)
}

func RunBeginBlock(ctx sdk.Context, app *simapp.OsmosisApp) sdk.Context {
	oldHeight := ctx.BlockHeight()
	oldHeader := ctx.BlockHeader()
	app.Commit()
	newHeader := tmproto.Header{Height: oldHeight + 1, ChainID: oldHeader.ChainID, Time: oldHeader.Time.Add(time.Second)}
	app.BeginBlock(abci.RequestBeginBlock{Header: newHeader})
	ctx = app.GetBaseApp().NewContext(false, newHeader)
	return ctx
}
