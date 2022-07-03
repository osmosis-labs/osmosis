package simulation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	legacysimexec "github.com/cosmos/cosmos-sdk/x/simulation"
)

// Action represents a simulator action.
// The details of this struct are internal,
// we currently plan on maintaining 3 constructors for it.
// * ActionFromWeightedOperation - for legacy simulator compatibility
// * ActionFromMsg - An easy API to go from creating a message via simctx to something simulator can deal with
// * ActionFromDependentMessages - API for defining a series of messages that depend on one another, and should satisfy
//   some properties post-execution.
type Action interface {
	Name() string
	// I envision this weight being provided from a config.
	// Module providers can optionally provide a default from an enum,
	// but this should not be the default.
	Weight() int
	Execute(*SimCtx, sdk.Context) (
		OperationMsg simulation.OperationMsg, futureOps []simulation.FutureOperation, err error)
}

type weightedOperationAction struct {
	op simulation.WeightedOperation
}

func (a weightedOperationAction) Name() string { return "weighted_op" }
func (a weightedOperationAction) Weight() int  { return a.op.Weight() }
func (a weightedOperationAction) Execute(sim *SimCtx, ctx sdk.Context) (
	simulation.OperationMsg, []simulation.FutureOperation, error) {
	return a.op.Op()(sim.GetRand(), sim.App, ctx, sim.Accounts, sim.ChainID)
}

func ActionsFromWeightedOperations(ops legacysimexec.WeightedOperations) []Action {
	actions := make([]Action, 0, len(ops))
	for _, op := range ops {
		actions = append(actions, weightedOperationAction{op: op})
	}
	return actions
}
