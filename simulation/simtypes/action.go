package simtypes

import (
	"fmt"

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
	// TODO: Maybe we make a "WithWeight" fn?
	Weight() Weight
	Execute(*SimCtx, sdk.Context) (
		OperationMsg simulation.OperationMsg, futureOps []simulation.FutureOperation, err error)
}

type weightedOperationAction struct {
	moduleName string
	op         simulation.WeightedOperation
}

func (a weightedOperationAction) Name() string   { return fmt.Sprintf("%s: weighted_op", a.moduleName) }
func (a weightedOperationAction) Weight() Weight { return Weight(a.op.Weight()) }
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
	return msgBasedAction{name: actionName, weight: Normal, msgGenerator: wrappedMsgGen}
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
	weight       Weight
	msgGenerator func(sim *SimCtx, ctx sdk.Context) (sdk.Msg, error)
}

func (m msgBasedAction) WithWeight(weight Weight) msgBasedAction {
	m.weight = weight
	return m
}

func (m msgBasedAction) Name() string   { return m.name }
func (m msgBasedAction) Weight() Weight { return m.weight }
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
