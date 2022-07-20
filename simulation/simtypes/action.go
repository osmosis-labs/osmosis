package simtypes

import (
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	legacysimexec "github.com/cosmos/cosmos-sdk/x/simulation"
)

// Action represents a simulator action.
// The details of this struct are internal,
// we currently plan on maintaining 2 constructors for it.
// * weightedOperationAction - for legacy simulator compatibility
// * msgBasedAction - An easy API to go from creating a message via simctx to something simulator can deal with
type Action interface {
	Name() string
	// I envision this frequency being provided from a config.
	// Module providers can optionally provide a default from an enum,
	// but this should not be the default.
	Frequency() Frequency
	Execute(*SimCtx, sdk.Context) (
		OperationMsg simulation.OperationMsg, futureOps []simulation.FutureOperation, err error)
	WithFrequency(w Frequency) Action
}

type selectActionFn func(r *rand.Rand) Action

type weightedOperationAction struct {
	moduleName string
	frequency  Frequency
	op         simulation.WeightedOperation
}

func (a weightedOperationAction) WithFrequency(w Frequency) Action { a.frequency = w; return a }
func (a weightedOperationAction) Name() string                     { return fmt.Sprintf("%s: weighted_op", a.moduleName) }
func (a weightedOperationAction) Frequency() Frequency {
	return Frequency(mapFrequencyFromInt(a.op.Weight()))
}
func (a weightedOperationAction) Execute(sim *SimCtx, ctx sdk.Context) (
	simulation.OperationMsg, []simulation.FutureOperation, error,
) {
	return a.op.Op()(sim.GetRand(), sim.BaseApp(), ctx, sim.Accounts, sim.ChainID())
}

func ActionsFromWeightedOperations(ops legacysimexec.WeightedOperations) []Action {
	return actionsFromWeightedOperations("no module name", ops)
}

func actionsFromWeightedOperations(moduleName string, ops legacysimexec.WeightedOperations) []Action {
	actions := make([]Action, 0, len(ops))
	for _, op := range ops {
		actions = append(actions, weightedOperationAction{moduleName: moduleName, op: op})
	}
	return actions
}

var _ Action = msgBasedAction{}

func NewKeeperlessMsgBasedAction[M sdk.Msg](actionName string, msgGenerator func(sim *SimCtx, ctx sdk.Context) (M, error)) Action {
	wrappedMsgGen := func(sim *SimCtx, ctx sdk.Context) (sdk.Msg, error) {
		return msgGenerator(sim, ctx)
	}
	// TODO: This likely won't work, and we need to instead make a mock sim ctx and ctx to get this.
	// TODO: Also do we even want this? Isn't the goal to write simulation event hooks based on Name
	// var sampleMsg M
	// msgName := osmoutils.GetType(sampleMsg)
	return msgBasedAction{name: actionName, frequency: Common, msgGenerator: wrappedMsgGen}
}

func CurryMsgGenerator[K interface{}, M sdk.Msg](k K, f func(K, *SimCtx, sdk.Context) (M, error)) func(*SimCtx, sdk.Context) (M, error) {
	return func(sim *SimCtx, ctx sdk.Context) (M, error) {
		return f(k, sim, ctx)
	}
}

func NewMsgBasedAction[K interface{}, M sdk.Msg](actionName string, k K, f func(K, *SimCtx, sdk.Context) (M, error)) Action {
	msgGenerator := CurryMsgGenerator(k, f)
	return NewKeeperlessMsgBasedAction(actionName, msgGenerator)
}

// TODO: make API in simulator action collection interface
// to add a 'modulename' to many actions

type msgBasedAction struct {
	name         string
	frequency    Frequency
	msgGenerator func(sim *SimCtx, ctx sdk.Context) (sdk.Msg, error)
}

func (m msgBasedAction) WithFrequency(w Frequency) Action { m.frequency = w; return m }
func (m msgBasedAction) Name() string                     { return m.name }
func (m msgBasedAction) Frequency() Frequency             { return m.frequency }
func (m msgBasedAction) Execute(sim *SimCtx, ctx sdk.Context) (
	OperationMsg simulation.OperationMsg, futureOps []simulation.FutureOperation, err error,
) {
	msg, err := m.msgGenerator(sim, ctx)
	if err != nil {
		return simulation.NoOpMsg(m.name, m.name, fmt.Sprintf("unable to build msg due to: %v", err)), nil, nil
	}
	tx, err := sim.txbuilder(ctx, msg, m.name)
	if err != nil {
		return simulation.NoOpMsg(m.name, m.name, fmt.Sprintf("unable to build tx due to: %v", err)), nil, err
	}
	return sim.deliverTx(tx, msg, m.name)
}

func totalFrequency(actions []Action) int {
	totalFrequency := 0
	for _, action := range actions {
		totalFrequency += mapFrequencyFromString(action.Frequency())
	}

	return totalFrequency
}

func GetSelectActionFn(actions []Action) selectActionFn {
	totalOpFrequency := totalFrequency(actions)

	return func(r *rand.Rand) Action {
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
