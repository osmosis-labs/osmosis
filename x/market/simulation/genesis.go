package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"github.com/osmosis-labs/osmosis/osmomath"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/v27/x/market/types"
)

// Simulation parameter constants
const (
	basePoolKey           = "base_pool"
	poolRecoveryPeriodKey = "pool_recovery_period"
	minStabilitySpreadKey = "min_spread"
)

// GenBasePool randomized MintBasePool
func GenBasePool(r *rand.Rand) osmomath.Dec {
	return osmomath.NewDec(50000000000000).Add(osmomath.NewDec(int64(r.Intn(10000000000))))
}

// GenPoolRecoveryPeriod randomized PoolRecoveryPeriod
func GenPoolRecoveryPeriod(r *rand.Rand) uint64 {
	return uint64(100 + r.Intn(10000000000))
}

// GenMinSpread randomized MinSpread
func GenMinSpread(r *rand.Rand) osmomath.Dec {
	return osmomath.NewDecWithPrec(1, 2).Add(osmomath.NewDecWithPrec(int64(r.Intn(100)), 3))
}

// RandomizedGenState generates a random GenesisState for gov
func RandomizedGenState(simState *module.SimulationState) {
	var poolRecoveryPeriod uint64
	simState.AppParams.GetOrGenerate(
		poolRecoveryPeriodKey, &poolRecoveryPeriod, simState.Rand,
		func(r *rand.Rand) { poolRecoveryPeriod = GenPoolRecoveryPeriod(r) },
	)

	var minStabilitySpread osmomath.Dec
	simState.AppParams.GetOrGenerate(
		minStabilitySpreadKey, &minStabilitySpread, simState.Rand,
		func(r *rand.Rand) { minStabilitySpread = GenMinSpread(r) },
	)

	marketGenesis := types.NewGenesisState(
		types.Params{
			MinStabilitySpread: minStabilitySpread,
		},
	)

	bz, err := json.MarshalIndent(&marketGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Selected randomly generated market parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(marketGenesis)
}
