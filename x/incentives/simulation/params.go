package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const (
	keyDistrEpochIdentifier = "DistrEpochIdentifier"
)

// TODO: remove hardcoded params
// refer x/epochs/simulation/genesis.go.
var epochIdentifiers = []string{"day", "hour"}

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation.
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyDistrEpochIdentifier,
			func(r *rand.Rand) string {
				return fmt.Sprintf(`{"%s: %s"}`, ParamsDistrEpochIdentifier, GenParamsDistrEpochIdentifier(r))
			},
		),
	}
}

func GenParamsDistrEpochIdentifier(r *rand.Rand) string {
	return epochIdentifiers[r.Intn(len(epochIdentifiers))]
}
