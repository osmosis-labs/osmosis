package simulation

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/types/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/merkle"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/osmosis-labs/osmosis/v16/simulation/executor/internal/stats"
	"github.com/osmosis-labs/osmosis/v16/simulation/simtypes"
)

type simState struct {
	// TODO: Fix things so this can be interface
	simParams Params
	header    tmproto.Header

	// These are operations which have been queued by previous operations
	// TODO: Replace with new action syntax
	operationQueue map[int][]simulation.Operation

	curValidators  mockValidators
	nextValidators mockValidators

	// block tx result data storage, reset after each block
	Data [][]byte

	// We technically have to store past block times for every block within the unbonding period.
	// For simplicity, we take the RAM overhead and store all past times.
	pastTimes     []time.Time
	pastVoteInfos [][]abci.VoteInfo

	logWriter LogWriter
	w         io.Writer

	// eventStats is an obviously bad design, but for now we leave it as future
	// work for us to clean up and architect well.
	// We should be collecting this raw data, and able to stream it out into a database.
	// Its fine to keep some basic aggregate statistics, but not where it should end.
	eventStats stats.EventStats
	opCount    int

	config Config
}

func newSimulatorState(tb testing.TB, simParams Params, initialHeader tmproto.Header, w io.Writer, validators mockValidators, config Config) *simState {
	tb.Helper()
	return &simState{
		simParams:      simParams,
		header:         initialHeader,
		operationQueue: NewOperationQueue(),
		curValidators:  validators.Clone(),
		nextValidators: validators.Clone(),
		pastTimes:      []time.Time{},
		pastVoteInfos:  [][]abci.VoteInfo{},
		logWriter:      NewLogWriter(tb),
		w:              w,
		eventStats:     stats.NewEventStats(),
		opCount:        0,
		config:         config,
	}
}

func (simState *simState) SimulateAllBlocks(
	w io.Writer,
	simCtx *simtypes.SimCtx,
	blockSimulator blockSimFn,
) (stopEarly bool, err error) {
	stopEarly = false
	initialHeight := simState.config.InitializationConfig.InitialBlockHeight
	numBlocks := simState.config.NumBlocks
	for height := initialHeight; height < numBlocks+initialHeight && !stopEarly; height++ {
		stopEarly, err = simState.SimulateBlock(simCtx, blockSimulator)
		if stopEarly {
			break
		}

		simCtx.BaseApp().Commit()
	}

	if !stopEarly {
		fmt.Fprintf(
			w,
			"\nSimulation complete; Final height (blocks): %d, final time (seconds): %v, operations ran: %d\n",
			simState.header.Height, simState.header.Time, simState.opCount,
		)
		simState.logWriter.PrintLogs()
	}
	return stopEarly, err
}

// simulate a block, update state
func (simState *simState) SimulateBlock(simCtx *simtypes.SimCtx, blockSimulator blockSimFn) (stopEarly bool, err error) {
	if simState.header.ProposerAddress == nil {
		fmt.Fprintf(simState.w, "\nSimulation stopped early as all validators have been unbonded; nobody left to propose a block!\n")
		return true, nil
	}

	requestBeginBlock := simState.beginBlock(simCtx)
	ctx := simCtx.BaseApp().NewContext(false, simState.header).WithBlockTime(simState.header.Time)

	// Run queued operations. Ignores blocksize if blocksize is too small
	numQueuedOpsRan, err := simState.runQueuedOperations(simCtx, ctx)
	if err != nil {
		return true, err
	}
	// numQueuedTimeOpsRan := simState.runQueuedTimeOperations(simCtx, ctx)

	// run standard operations
	// TODO: rename blockSimulator arg
	operations, err := blockSimulator(simCtx, ctx, simState.header)
	simState.opCount += operations + numQueuedOpsRan // + numQueuedTimeOpsRan
	if err != nil {
		return true, err
	}

	responseEndBlock := simState.endBlock(simCtx)

	err = simState.prepareNextSimState(simCtx, requestBeginBlock, responseEndBlock)
	if err != nil {
		return true, err
	}
	return false, nil
}

func (simState *simState) beginBlock(simCtx *simtypes.SimCtx) abci.RequestBeginBlock {
	// Generate a random RequestBeginBlock with the current validator set
	requestBeginBlock := RandomRequestBeginBlock(simCtx.GetRand(), simState.simParams, simState.curValidators, simState.pastTimes, simState.pastVoteInfos, simState.eventStats.Tally, simState.header)
	// Run the BeginBlock handler
	simState.logWriter.AddEntry(BeginBlockEntry(simState.header.Height))
	simCtx.BaseApp().BeginBlock(requestBeginBlock)
	return requestBeginBlock
}

func (simState *simState) endBlock(simCtx *simtypes.SimCtx) abci.ResponseEndBlock {
	res := simCtx.BaseApp().EndBlock(abci.RequestEndBlock{})
	simState.logWriter.AddEntry(EndBlockEntry(simState.header.Height))
	return res
}

func (simState *simState) prepareNextSimState(simCtx *simtypes.SimCtx, req abci.RequestBeginBlock, res abci.ResponseEndBlock) error {
	// Log the current block's header time for future lookup
	simState.pastTimes = append(simState.pastTimes, simState.header.Time)
	simState.pastVoteInfos = append(simState.pastVoteInfos, req.LastCommitInfo.Votes)

	// increase header height by one
	simState.header.Height++

	// set header time
	timeDiff := maxTimePerBlock - minTimePerBlock
	simState.header.Time = simState.header.Time.Add(
		time.Duration(minTimePerBlock) * time.Second)
	simState.header.Time = simState.header.Time.Add(
		time.Duration(int64(simCtx.GetRand().Intn(int(timeDiff)))) * time.Second)

	// Draw the block proposer from proposers for n+1
	proposerPubKey := simState.nextValidators.randomProposer(simCtx.GetRand())
	simState.header.ProposerAddress = proposerPubKey.Address()
	// find N + 2 valset
	nPlus2Validators, err := updateValidators(simCtx.GetRand(), simState.simParams, simState.nextValidators, res.ValidatorUpdates, simState.eventStats.Tally)
	if err != nil {
		return err
	}

	// now set variables in perspective of block n+1
	simState.curValidators = simState.nextValidators
	simState.nextValidators = nPlus2Validators

	// utilize proposer public key and generate validator hash
	// then, with completed block header, generate app hash
	// see https://github.com/tendermint/tendermint/blob/v0.34.x/spec/core/data_structures.md#header for more info on block header hashes
	return simState.constructHeaderHashes(proposerPubKey)
}

func (simState *simState) constructHeaderHashes(proposerPubKey crypto.PubKey) error {
	currentValSet, err := simState.curValidators.toTmProtoValidators(proposerPubKey)
	if err != nil {
		return err
	}

	// generate a hash from the validatorSet type and set it to the validators hash
	simState.header.ValidatorsHash = currentValSet.Hash()

	// create a header type from the tmproto
	realHeader, err := tmtypes.HeaderFromProto(&simState.header)
	if err != nil {
		return err
	}

	// TODO: We don't fill out the nextValidatorSet, we should eventually but not priority
	// It is unclear to me if this means we need to choose a proposer prior to that block occurring

	// hash all the collected tx results and add as the block header data hash
	dataHash := merkle.HashFromByteSlices(simState.Data)
	simState.header.DataHash = dataHash

	// reset simState data to blank
	simState.Data = [][]byte{}

	// generate an apphash from this header and set this value
	simState.header.AppHash = realHeader.Hash()
	return nil
}
