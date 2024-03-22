package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/osmosis-labs/osmosis/v23/x/market/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(*rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyBasePool),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenBasePool(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyPoolRecoveryPeriod),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenPoolRecoveryPeriod(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMinStabilitySpread),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenMinSpread(r))
			},
		),
	}
}
