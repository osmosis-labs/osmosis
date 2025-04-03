package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/osmosis-labs/osmosis/v27/x/treasury/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(*rand.Rand) []simtypes.LegacyParamChange {
	return []simtypes.LegacyParamChange{
		simulation.NewSimLegacyParamChange(types.ModuleName, string(types.KeyWindowShort),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenWindowShort(r))
			},
		),
		simulation.NewSimLegacyParamChange(types.ModuleName, string(types.KeyWindowLong),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenWindowLong(r))
			},
		),
		simulation.NewSimLegacyParamChange(types.ModuleName, string(types.KeyWindowProbation),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenWindowProbation(r))
			},
		),
	}
}
