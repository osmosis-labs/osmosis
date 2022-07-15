package simulation

// TODO: Figure out how to delete everything in this file OR migrate to improved logging concepts

import (
	"encoding/json"
	"math/rand"
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/types/simulation"

	simtypes "github.com/osmosis-labs/osmosis/v7/simulation/types"
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
// TODO: Change FutureOperation to FutureAction
func (simState *simState) queueOperations(futureOps []simulation.FutureOperation) {
	if len(futureOps) == 0 {
		return
	}

	for _, futureOp := range futureOps {
		futureOp := futureOp
		if futureOp.BlockHeight != 0 {
			if val, ok := simState.operationQueue[futureOp.BlockHeight]; ok {
				simState.operationQueue[futureOp.BlockHeight] = append(val, futureOp.Op)
			} else {
				simState.operationQueue[futureOp.BlockHeight] = []simulation.Operation{futureOp.Op}
			}

			continue
		}

		if (futureOp.BlockTime == time.Time{}) {
			continue
		}
		// TODO: Replace with proper sorted data structure, so don't have the
		// copy entire slice
		index := sort.Search(
			len(simState.timeOperationQueue),
			func(i int) bool {
				return simState.timeOperationQueue[i].BlockTime.After(futureOp.BlockTime)
			},
		)

		simState.timeOperationQueue = append(simState.timeOperationQueue, simulation.FutureOperation{})
		copy(simState.timeOperationQueue[index+1:], simState.timeOperationQueue[index:])
		simState.timeOperationQueue[index] = futureOp
	}
}

func totalWeight(actions []simtypes.Action) int {
	totalWeight := 0
	for _, action := range actions {
		totalWeight += int(action.Weight())
	}

	return totalWeight
}

type selectActionFn func(r *rand.Rand) simtypes.Action

func getSelectActionFn(actions []simtypes.Action) selectActionFn {
	totalOpWeight := totalWeight(actions)

	return func(r *rand.Rand) simtypes.Action {
		x := r.Intn(totalOpWeight)
		// TODO: Change to an accum list approach
		for i := 0; i < len(actions); i++ {
			if x <= int(actions[i].Weight()) {
				return actions[i]
			}

			x -= int(actions[i].Weight())
		}
		// shouldn't happen
		return actions[0]
	}
}
