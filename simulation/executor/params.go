package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	markov "github.com/osmosis-labs/osmosis/v12/simulation/simtypes/transitionmatrix"
)

const (
	// Minimum time per block
	minTimePerBlock int64 = 10000 / 2

	// Maximum time per block
	maxTimePerBlock int64 = 10000
)

// TODO: explain transitional matrix usage
var (
	// Currently there are 3 different liveness types,
	// fully online, spotty connection, offline.
	defaultLivenessTransitionMatrix, _ = markov.CreateTransitionMatrix([][]int{
		{90, 20, 1},
		{10, 50, 5},
		{0, 10, 1000},
	})

	// 3 states: rand in range [0, 4*provided blocksize],
	// rand in range [0, 2 * provided blocksize], 0
	defaultBlockSizeTransitionMatrix, _ = markov.CreateTransitionMatrix([][]int{
		{85, 5, 0},
		{15, 92, 40},
		{0, 3, 60},
	})
)

// Params define the parameters necessary for running the simulations
type Params struct {
	pastEvidenceFraction      float64
	numKeys                   int
	evidenceFraction          float64
	initialLivenessWeightings []int
	livenessTransitionMatrix  simulation.TransitionMatrix
	blockSizeTransitionMatrix simulation.TransitionMatrix
}

func (p Params) PastEvidenceFraction() float64 {
	return p.pastEvidenceFraction
}

func (p Params) NumKeys() int {
	return p.numKeys
}

func (p Params) EvidenceFraction() float64 {
	return p.evidenceFraction
}

func (p Params) InitialLivenessWeightings() []int {
	return p.initialLivenessWeightings
}

func (p Params) LivenessTransitionMatrix() simulation.TransitionMatrix {
	return p.livenessTransitionMatrix
}

func (p Params) BlockSizeTransitionMatrix() simulation.TransitionMatrix {
	return p.blockSizeTransitionMatrix
}

// RandomParams returns random simulation parameters
func RandomParams(r *rand.Rand) Params {
	return Params{
		pastEvidenceFraction:      r.Float64(),
		numKeys:                   simulation.RandIntBetween(r, 50, 2500), // number of accounts created for the simulation
		evidenceFraction:          r.Float64() / 2,
		initialLivenessWeightings: []int{simulation.RandIntBetween(r, 1, 80), r.Intn(10), r.Intn(10)},
		livenessTransitionMatrix:  defaultLivenessTransitionMatrix,
		blockSizeTransitionMatrix: defaultBlockSizeTransitionMatrix,
	}
}

// randomConsensusParams returns random simulation consensus parameters, it extracts the Evidence from the Staking genesis state.
func randomConsensusParams(r *rand.Rand, appState json.RawMessage, cdc codec.JSONCodec) *abci.ConsensusParams {
	var genesisState map[string]json.RawMessage
	err := json.Unmarshal(appState, &genesisState)
	if err != nil {
		panic(err)
	}

	stakingGenesisState := stakingtypes.GetGenesisStateFromAppState(cdc, genesisState)
	consensusParams := &abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxBytes: int64(simulation.RandIntBetween(r, 20000000, 30000000)),
			MaxGas:   -1,
		},
		Validator: &tmproto.ValidatorParams{
			PubKeyTypes: []string{types.ABCIPubKeyTypeEd25519},
		},
		Evidence: &tmproto.EvidenceParams{
			MaxAgeNumBlocks: int64(stakingGenesisState.Params.UnbondingTime / AverageBlockTime),
			MaxAgeDuration:  stakingGenesisState.Params.UnbondingTime,
		},
		Version: &tmproto.VersionParams{
			AppVersion: uint64(simulation.RandIntBetween(r, 0, 10000)),
		},
	}

	bz, err := json.MarshalIndent(&consensusParams, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated consensus parameters:\n%s\n", bz)

	return consensusParams
}
