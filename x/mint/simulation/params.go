package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/c-osmosis/osmosis/x/mint/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

const (
	keyMaxRewardPerEpoch = "MaxRewardPerEpoch"
	keyMinRewardPerEpoch = "MinRewardPerEpoch"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyMaxRewardPerEpoch,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenMaxRewardPerEpoch(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyMinRewardPerEpoch,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenMinRewardPerEpoch(r))
			},
		),
	}
}
