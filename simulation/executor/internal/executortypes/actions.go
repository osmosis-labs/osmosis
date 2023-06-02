package executortypes

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/simulation"
	legacysimexec "github.com/cosmos/cosmos-sdk/x/simulation"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/simulation/simtypes"
)

type selectActionFn func(r *rand.Rand) simtypes.ActionsWithMetadata

func GetSelectActionFn(actions []simtypes.ActionsWithMetadata) selectActionFn {
	totalOpFrequency := totalFrequency(actions)

	return func(r *rand.Rand) simtypes.ActionsWithMetadata {
		x := r.Intn(totalOpFrequency)
		// TODO: Change to an accum list approach
		for i := 0; i < len(actions); i++ {
			if x <= mapFrequencyFromString(actions[i].Frequency()) {
				return actions[i]
			}

			x -= mapFrequencyFromString(actions[i].Frequency())
		}
		// shouldn't happen
		return actions[0]
	}
}

func ActionsFromWeightedOperations(ops legacysimexec.WeightedOperations) []simtypes.Action {
	return actionsFromWeightedOperations("no module name", ops)
}

func actionsFromWeightedOperations(moduleName string, ops legacysimexec.WeightedOperations) []simtypes.Action {
	actions := make([]simtypes.Action, 0, len(ops))
	for _, op := range ops {
		actions = append(actions, weightedOperationAction{moduleName: moduleName, op: op})
	}
	return actions
}

type weightedOperationAction struct {
	moduleName string
	frequency  simtypes.Frequency
	op         simulation.WeightedOperation
}

func (a weightedOperationAction) WithFrequency(w simtypes.Frequency) simtypes.Action {
	a.frequency = w
	return a
}
func (a weightedOperationAction) Name() string { return fmt.Sprintf("%s: weighted_op", a.moduleName) }
func (a weightedOperationAction) Frequency() simtypes.Frequency {
	return mapFrequencyFromInt(a.op.Weight())
}

func (a weightedOperationAction) Execute(sim *simtypes.SimCtx, ctx sdk.Context) (
	simulation.OperationMsg, []simulation.FutureOperation, []byte, error,
) {
	op, futureOp, err := a.op.Op()(sim.GetRand(), sim.BaseApp(), ctx, sim.Accounts, sim.ChainID())
	// pass in an empty byte array for response data for all weighted operations
	return op, futureOp, []byte{}, err
}
