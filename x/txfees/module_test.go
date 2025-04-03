package txfees_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"

	coreheader "cosmossdk.io/core/header"
	abci "github.com/cometbft/cometbft/abci/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	osmosisapp "github.com/osmosis-labs/osmosis/v27/app"

	simapp "github.com/osmosis-labs/osmosis/v27/app"
	mempool1559 "github.com/osmosis-labs/osmosis/v27/x/txfees/keeper/mempool-1559"
)

func TestSetBaseDenomOnInitBlock(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})

	genesisState := osmosisapp.GenesisStateWithValSet(app)
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(err)
	}

	app.InitChain(
		&abcitypes.RequestInitChain{
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
	dirName := fmt.Sprintf("%d", rand.Int())
	app := simapp.SetupWithCustomHome(false, dirName)

	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{ChainID: "osmosis-1", Height: 1})

	genesisState := osmosisapp.GenesisStateWithValSet(app)
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(err)
	}

	app.InitChain(
		&abcitypes.RequestInitChain{
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
	ctx = RunFinalizeBlock(ctx, app)

	// Target gas should be updated to the value set in InitChain
	defaultBlockTargetGas := mempool1559.TargetBlockSpacePercent.Mul(osmomath.NewDec(sims.DefaultConsensusParams.Block.MaxGas)).TruncateInt().Int64()
	require.Equal(t, defaultBlockTargetGas, mempool1559.TargetGas)

	// Run begin block again, should not update target gas
	ctx = RunFinalizeBlock(ctx, app)
	require.Equal(t, defaultBlockTargetGas, mempool1559.TargetGas)

	// Update the consensus params
	newDefaultBlockMaxGas := int64(300_000_000)
	newConsensusParams := *sims.DefaultConsensusParams
	newConsensusParams.Block.MaxGas = newDefaultBlockMaxGas
	err = app.ConsensusParamsKeeper.ParamsStore.Set(ctx, newConsensusParams)
	if err != nil {
		panic(err)
	}

	// Ensure that the consensus params have not been updated yet
	require.Equal(t, defaultBlockTargetGas, mempool1559.TargetGas)

	// Run begin block again, should update target gas
	RunFinalizeBlock(ctx, app)
	expectedNewBlockTargetGas := mempool1559.TargetBlockSpacePercent.Mul(osmomath.NewDec(newDefaultBlockMaxGas)).TruncateInt().Int64()
	require.Equal(t, expectedNewBlockTargetGas, mempool1559.TargetGas)

	os.RemoveAll(dirName)
}

func RunFinalizeBlock(ctx sdk.Context, app *simapp.SymphonyApp) sdk.Context {
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: ctx.BlockHeight(), Time: ctx.BlockTime()})
	if err != nil {
		panic(err)
	}
	_, err = app.Commit()
	if err != nil {
		panic(err)
	}
	header := ctx.BlockHeader()
	header.Time = ctx.BlockTime()
	header.Height++

	ctx = app.GetBaseApp().NewUncachedContext(false, header).WithHeaderInfo(coreheader.Info{
		Height: header.Height,
		Time:   header.Time,
	})
	return ctx
}
