package simulation

import (
	"encoding/json"
	"fmt"

<<<<<<< HEAD
	"github.com/osmosis-labs/osmosis/v18/x/superfluid/types"
=======
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/superfluid/types"
>>>>>>> ca75f4c3 (refactor(deps): switch to cosmossdk.io/math from fork math (#6238))

	"github.com/cosmos/cosmos-sdk/types/module"
)

// RandomizedGenState generates a random GenesisState for staking.
func RandomizedGenState(simState *module.SimulationState) {
	superfluidGenesis := &types.GenesisState{
		Params: types.Params{
			MinimumRiskFactor: osmomath.NewDecWithPrec(5, 2), // 5%
		},
		SuperfluidAssets:          []types.SuperfluidAsset{},
		OsmoEquivalentMultipliers: []types.OsmoEquivalentMultiplierRecord{},
	}

	bz, err := json.MarshalIndent(&superfluidGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated superfluid parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(superfluidGenesis)
}
