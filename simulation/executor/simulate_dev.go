package simulation

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/types/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simtypes "github.com/osmosis-labs/osmosis/v7/simulation/types"
)

type simState struct {
	// TODO: Fix things so this can be interface
	simParams Params
	header    tmproto.Header

	// These are operations which have been queued by previous operations
	// TODO: Replace with new action syntax
	operationQueue     map[int][]simulation.Operation
	timeOperationQueue []simulation.FutureOperation

	curValidators  mockValidators
	nextValidators mockValidators

	// TODO: Figure out why we need this???
	// Probably should get removed
	tb testing.TB

	// We technically have to store past block times for every block within the unbonding period.
	// For simplicity, we take the RAM overhead and store all past times.
	pastTimes     []time.Time
	pastVoteInfos [][]abci.VoteInfo

	leanLogs  bool
	logWriter LogWriter
	w         io.Writer

	// eventStats is an obviously bad design, but for now we leave it as future
	// work for us to clean up and architect well.
	// We should be collecting this raw data, and able to stream it out into a database.
	// Its fine to keep some basic aggregate statistics, but not where it should end.
	eventStats EventStats
	opCount    int
}

func newSimulatorState(simParams Params, initialHeader tmproto.Header, tb testing.TB, w io.Writer, validators mockValidators) *simState {
	return &simState{
		simParams:          simParams,
		header:             initialHeader,
		operationQueue:     NewOperationQueue(),
		timeOperationQueue: []simulation.FutureOperation{},
		curValidators:      validators.Clone(),
		nextValidators:     validators.Clone(),
		tb:                 tb,
		pastTimes:          []time.Time{},
		pastVoteInfos:      [][]abci.VoteInfo{},
		logWriter:          NewLogWriter(tb),
		w:                  w,
		eventStats:         NewEventStats(),
		opCount:            0,
	}
}

func (s *simState) WithLogParam(leanLogs bool) *simState {
	s.leanLogs = leanLogs
	return s
}

// simulate a block, update state
func (simState *simState) SimulateBlock(simCtx *simtypes.SimCtx, blockSimulator blockSimFn) (stopEarly bool) {
	if simState.header.ProposerAddress == nil {
		fmt.Fprintf(simState.w, "\nSimulation stopped early as all validators have been unbonded; nobody left to propose a block!\n")
		return true
	}

	requestBeginBlock := simState.beginBlock(simCtx)
	ctx := simCtx.App.NewContext(false, simState.header)

	// Run queued operations. Ignores blocksize if blocksize is too small
	numQueuedOpsRan := simState.runQueuedOperations(simCtx, ctx)
	numQueuedTimeOpsRan := simState.runQueuedTimeOperations(simCtx, ctx)

	// run standard operations
	operations := blockSimulator(simCtx, ctx, simState.header)
	simState.opCount += operations + numQueuedOpsRan + numQueuedTimeOpsRan

	responseEndBlock := simState.endBlock(simCtx)

	simState.prepareNextSimState(simCtx, requestBeginBlock, responseEndBlock)

	return false
}

func (simState *simState) beginBlock(simCtx *simtypes.SimCtx) abci.RequestBeginBlock {
	// Generate a random RequestBeginBlock with the current validator set
	requestBeginBlock := RandomRequestBeginBlock(simCtx.GetRand(), simState.simParams, simState.curValidators, simState.pastTimes, simState.pastVoteInfos, simState.eventStats.Tally, simState.header)
	// Run the BeginBlock handler
	simState.logWriter.AddEntry(BeginBlockEntry(simState.header.Height))
	simCtx.App.BeginBlock(requestBeginBlock)
	return requestBeginBlock
}

func (simState *simState) endBlock(simCtx *simtypes.SimCtx) abci.ResponseEndBlock {
	res := simCtx.App.EndBlock(abci.RequestEndBlock{})
	simState.logWriter.AddEntry(EndBlockEntry(simState.header.Height))
	return res
}

func (simState *simState) prepareNextSimState(simCtx *simtypes.SimCtx, req abci.RequestBeginBlock, res abci.ResponseEndBlock) {
	// Log the current block's header time for future lookup
	simState.pastTimes = append(simState.pastTimes, simState.header.Time)
	simState.pastVoteInfos = append(simState.pastVoteInfos, req.LastCommitInfo.Votes)

	simState.header.Height++

	timeDiff := maxTimePerBlock - minTimePerBlock
	simState.header.Time = simState.header.Time.Add(
		time.Duration(minTimePerBlock) * time.Second)
	simState.header.Time = simState.header.Time.Add(
		time.Duration(int64(simCtx.GetRand().Intn(int(timeDiff)))) * time.Second)

	// Draw the block proposer from proposers for n+1
	simState.header.ProposerAddress = simState.nextValidators.randomProposer(simCtx.GetRand())
	// find N + 2 valset
	nPlus2Validators := updateValidators(simState.tb, simCtx.GetRand(), simState.simParams, simState.nextValidators, res.ValidatorUpdates, simState.eventStats.Tally)

	// now set variables in perspective of block n+1
	simState.curValidators = simState.nextValidators
	simState.nextValidators = nPlus2Validators
}
