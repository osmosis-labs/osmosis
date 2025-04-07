package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"github.com/osmosis-labs/osmosis/osmomath"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/v27/x/treasury/types"
)

// Simulation parameter constants
const (
	taxPolicyKey               = "tax_policy"
	rewardPolicyKey            = "reward_policy"
	seigniorageBurdenTargetKey = "seigniorage_burden_target"
	miningIncrementKey         = "mining_increment"
	windowShortKey             = "window_short"
	windowLongKey              = "window_long"
	windowProbationKey         = "window_probation"
)

// GenWindowShort randomized WindowShort
func GenWindowShort(r *rand.Rand) uint64 {
	return uint64(1 + r.Intn(12))
}

// GenWindowLong randomized WindowLong
func GenWindowLong(r *rand.Rand) uint64 {
	return uint64(12 + r.Intn(24))
}

// GenWindowProbation randomized WindowProbation
func GenWindowProbation(r *rand.Rand) uint64 {
	return uint64(1 + r.Intn(6))
}

// RandomizedGenState generates a random GenesisState for gov
func RandomizedGenState(simState *module.SimulationState) {
	var windowShort uint64
	simState.AppParams.GetOrGenerate(
		windowShortKey, &windowShort, simState.Rand,
		func(r *rand.Rand) { windowShort = GenWindowShort(r) },
	)

	var windowLong uint64
	simState.AppParams.GetOrGenerate(
		windowLongKey, &windowLong, simState.Rand,
		func(r *rand.Rand) { windowLong = GenWindowLong(r) },
	)

	var windowProbation uint64
	simState.AppParams.GetOrGenerate(
		windowProbationKey, &windowProbation, simState.Rand,
		func(r *rand.Rand) { windowProbation = GenWindowProbation(r) },
	)

	treasuryGenesis := types.NewGenesisState(
		types.Params{
			ReserveAllowableOffset: osmomath.Dec{},
			MaxFeeMultiplier:       osmomath.Dec{},
			WindowShort:            windowShort,
			WindowLong:             windowLong,
			WindowProbation:        windowProbation,
		},
		osmomath.Dec{},
	)

	bz, err := json.MarshalIndent(&treasuryGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Selected randomly generated market parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(treasuryGenesis)
}
