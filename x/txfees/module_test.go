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
	osmosisapp "github.com/osmosis-labs/osmosis/v32/app"

	simapp "github.com/osmosis-labs/osmosis/v32/app"
	"github.com/osmosis-labs/osmosis/v32/x/txfees"
	mempool1559 "github.com/osmosis-labs/osmosis/v32/x/txfees/keeper/mempool-1559"
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

// TestCheckAndSetTargetGas tests that the target gas is only ever set to a positive value, both on the first
// block after node start and when the block max gas changes while the node is running. A block max gas that
// yields no positive target gas must leave the target gas as is, and must not prevent later changes from
// being applied.
func TestCheckAndSetTargetGas(t *testing.T) {
	const (
		initialMaxGas = int64(120_000_000)
		laterMaxGas   = int64(300_000_000)
	)
	targetGasOf := func(maxGas int64) int64 {
		return mempool1559.TargetBlockSpacePercent.Mul(osmomath.NewDec(maxGas)).TruncateInt().Int64()
	}
	blockParams := func(maxGas int64) *tmproto.BlockParams {
		return &tmproto.BlockParams{MaxBytes: 200_000, MaxGas: maxGas}
	}

	tests := map[string]struct {
		block *tmproto.BlockParams
		// expectedTargetGas is the target gas the block params must yield. Zero means that the target gas
		// must be left unchanged.
		expectedTargetGas int64
	}{
		"unlimited block gas, max gas -1":         {block: blockParams(-1)},
		"unlimited block gas, max gas 0":          {block: blockParams(0)},
		"unset block params":                      {block: nil},
		"max gas 1 truncates to a zero target":    {block: blockParams(1)},
		"smallest max gas with a positive target": {block: blockParams(2), expectedTargetGas: 1},
		"regular max gas":                         {block: blockParams(200_000_000), expectedTargetGas: 125_000_000},
	}

	// A home directory of its own is needed, as the wasm VM locks the home directory it runs in.
	app := simapp.SetupWithCustomHome(false, t.TempDir())
	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{ChainID: "osmosis-1", Height: 1})
	am := txfees.NewAppModule(*app.TxFeesKeeper)

	// The target gas and the cached consensus params are process wide, restore them for other tests.
	originalTargetGas := mempool1559.TargetGas
	t.Cleanup(func() {
		mempool1559.TargetGas = originalTargetGas
		txfees.ResetCachedConsParams()
	})

	setBlockParams := func(t *testing.T, block *tmproto.BlockParams) {
		t.Helper()
		consParams := *sims.DefaultConsensusParams
		consParams.Block = block
		require.NoError(t, app.ConsensusParamsKeeper.ParamsStore.Set(ctx, consParams))
	}
	// requireTargetGas runs the check twice, as a second block with the same params must not change the outcome.
	requireTargetGas := func(t *testing.T, expectedTargetGas int64) {
		t.Helper()
		for range 2 {
			require.NotPanics(t, func() { am.CheckAndSetTargetGas(ctx) })
			require.Equal(t, expectedTargetGas, mempool1559.TargetGas)
		}
	}

	for name, tc := range tests {
		expectedTargetGas := func(currentTargetGas int64) int64 {
			if tc.expectedTargetGas == 0 {
				return currentTargetGas
			}
			return tc.expectedTargetGas
		}

		t.Run("node start: "+name, func(t *testing.T) {
			txfees.ResetCachedConsParams()
			mempool1559.TargetGas = originalTargetGas

			setBlockParams(t, tc.block)
			requireTargetGas(t, expectedTargetGas(originalTargetGas))

			// A later change of the max gas must be applied.
			setBlockParams(t, blockParams(laterMaxGas))
			requireTargetGas(t, targetGasOf(laterMaxGas))
		})

		t.Run("changed params: "+name, func(t *testing.T) {
			txfees.ResetCachedConsParams()
			mempool1559.TargetGas = originalTargetGas

			setBlockParams(t, blockParams(initialMaxGas))
			requireTargetGas(t, targetGasOf(initialMaxGas))

			setBlockParams(t, tc.block)
			requireTargetGas(t, expectedTargetGas(targetGasOf(initialMaxGas)))

			// A later change of the max gas must be applied.
			setBlockParams(t, blockParams(laterMaxGas))
			requireTargetGas(t, targetGasOf(laterMaxGas))
		})
	}
}

func RunFinalizeBlock(ctx sdk.Context, app *simapp.OsmosisApp) sdk.Context {
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
