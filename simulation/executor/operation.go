package simulation

import (
	"encoding/json"
	"math/rand"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	simtypes "github.com/osmosis-labs/osmosis/v7/simulation/types"

	legacysimexec "github.com/cosmos/cosmos-sdk/x/simulation"
	// simtypes "github.com/osmosis-labs/osmosis/simulation/types"
)

// entry kinds for use within OperationEntry
const (
	BeginBlockEntryKind = "begin_block"
	EndBlockEntryKind   = "end_block"
	MsgEntryKind        = "msg"
	QueuedMsgEntryKind  = "queued_msg"
)

// OperationEntry - an operation entry for logging (ex. BeginBlock, EndBlock, XxxMsg, etc)
type OperationEntry struct {
	EntryKind string          `json:"entry_kind" yaml:"entry_kind"`
	Height    int64           `json:"height" yaml:"height"`
	Order     int64           `json:"order" yaml:"order"`
	Operation json.RawMessage `json:"operation" yaml:"operation"`
}

// NewOperationEntry creates a new OperationEntry instance
func NewOperationEntry(entry string, height, order int64, op json.RawMessage) OperationEntry {
	return OperationEntry{
		EntryKind: entry,
		Height:    height,
		Order:     order,
		Operation: op,
	}
}

// BeginBlockEntry - operation entry for begin block
func BeginBlockEntry(height int64) OperationEntry {
	return NewOperationEntry(BeginBlockEntryKind, height, -1, nil)
}

// EndBlockEntry - operation entry for end block
func EndBlockEntry(height int64) OperationEntry {
	return NewOperationEntry(EndBlockEntryKind, height, -1, nil)
}

// MsgEntry - operation entry for standard msg
func MsgEntry(height, order int64, opMsg simulation.OperationMsg) OperationEntry {
	return NewOperationEntry(MsgEntryKind, height, order, opMsg.MustMarshal())
}

// QueuedMsgEntry creates an operation entry for a given queued message.
func QueuedMsgEntry(height int64, opMsg simulation.OperationMsg) OperationEntry {
	return NewOperationEntry(QueuedMsgEntryKind, height, -1, opMsg.MustMarshal())
}

// MustMarshal marshals the operation entry, panic on error.
func (oe OperationEntry) MustMarshal() json.RawMessage {
	out, err := json.Marshal(oe)
	if err != nil {
		panic(err)
	}

	return out
}

// OperationQueue defines an object for a queue of operations
type OperationQueue map[int][]simulation.Operation

// NewOperationQueue creates a new OperationQueue instance.
func NewOperationQueue() OperationQueue {
	return make(OperationQueue)
}

// queueOperations adds all future operations into the operation queue.
func queueOperations(queuedOps OperationQueue, queuedTimeOps []simulation.FutureOperation, futureOps []simulation.FutureOperation) {
	if futureOps == nil {
		return
	}

	for _, futureOp := range futureOps {
		futureOp := futureOp
		if futureOp.BlockHeight != 0 {
			if val, ok := queuedOps[futureOp.BlockHeight]; ok {
				queuedOps[futureOp.BlockHeight] = append(val, futureOp.Op)
			} else {
				queuedOps[futureOp.BlockHeight] = []simulation.Operation{futureOp.Op}
			}

			continue
		}

		// TODO: Replace with proper sorted data structure, so don't have the
		// copy entire slice
		index := sort.Search(
			len(queuedTimeOps),
			func(i int) bool {
				return queuedTimeOps[i].BlockTime.After(futureOp.BlockTime)
			},
		)

		queuedTimeOps = append(queuedTimeOps, simulation.FutureOperation{})
		copy(queuedTimeOps[index+1:], queuedTimeOps[index:])
		queuedTimeOps[index] = futureOp
	}
}

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
	Execute(*simtypes.SimCtx, sdk.Context) (
		OperationMsg simulation.OperationMsg, futureOps []simulation.FutureOperation, err error)
}

type weightedOperationAction struct {
	op simulation.WeightedOperation
}

func (a weightedOperationAction) Name() string { return "weighted_op" }
func (a weightedOperationAction) Weight() int  { return a.op.Weight() }
func (a weightedOperationAction) Execute(sim *simtypes.SimCtx, ctx sdk.Context) (
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

func totalWeight(actions []Action) int {
	totalWeight := 0
	for _, action := range actions {
		totalWeight += action.Weight()
	}

	return totalWeight
}

type selectActionFn func(r *rand.Rand) Action

func getSelectActionFn(actions []Action) selectActionFn {
	totalOpWeight := totalWeight(actions)

	return func(r *rand.Rand) Action {
		x := r.Intn(totalOpWeight)
		// TODO: Change to an accum list approach
		for i := 0; i < len(actions); i++ {
			if x <= actions[i].Weight() {
				return actions[i]
			}

			x -= actions[i].Weight()
		}
		// shouldn't happen
		return actions[0]
	}
}
