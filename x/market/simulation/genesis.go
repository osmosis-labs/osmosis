package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/v23/x/market/types"
)

// Simulation parameter constants
const (
	basePoolKey           = "base_pool"
	poolRecoveryPeriodKey = "pool_recovery_period"
	minStabilitySpreadKey = "min_spread"
)

// GenBasePool randomized MintBasePool
func GenBasePool(r *rand.Rand) sdk.Dec {
	return sdk.NewDec(50000000000000).Add(sdk.NewDec(int64(r.Intn(10000000000))))
}

// GenPoolRecoveryPeriod randomized PoolRecoveryPeriod
func GenPoolRecoveryPeriod(r *rand.Rand) uint64 {
	return uint64(100 + r.Intn(10000000000))
}

// GenMinSpread randomized MinSpread
func GenMinSpread(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(100)), 3))
}

// RandomizedGenState generates a random GenesisState for gov
func RandomizedGenState(simState *module.SimulationState) {
	var basePool sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, basePoolKey, &basePool, simState.Rand,
		func(r *rand.Rand) { basePool = GenBasePool(r) },
	)

	var poolRecoveryPeriod uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, poolRecoveryPeriodKey, &poolRecoveryPeriod, simState.Rand,
		func(r *rand.Rand) { poolRecoveryPeriod = GenPoolRecoveryPeriod(r) },
	)

	var minStabilitySpread sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, minStabilitySpreadKey, &minStabilitySpread, simState.Rand,
		func(r *rand.Rand) { minStabilitySpread = GenMinSpread(r) },
	)

	marketGenesis := types.NewGenesisState(
		types.Params{
			BasePool:           basePool,
			PoolRecoveryPeriod: poolRecoveryPeriod,
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
