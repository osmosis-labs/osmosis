package simulation

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	simtypes "github.com/osmosis-labs/osmosis/v7/simulation/types"
)

const AverageBlockTime = 6 * time.Second

// initialize the chain for the simulation
func initChain(
	r *rand.Rand,
	params Params,
	accounts []simulation.Account,
	app *baseapp.BaseApp,
	appStateFn simulation.AppStateFn,
	config simulation.Config,
	cdc codec.JSONCodec,
) (mockValidators, time.Time, []simulation.Account, string) {

	appState, accounts, chainID, genesisTimestamp := appStateFn(r, accounts, config)
	consensusParams := randomConsensusParams(r, appState, cdc)
	req := abci.RequestInitChain{
		AppStateBytes:   appState,
		ChainId:         chainID,
		ConsensusParams: consensusParams,
		Time:            genesisTimestamp,
	}
	// Valid app version can only be zero on app initialization.
	req.ConsensusParams.Version.AppVersion = 0
	res := app.InitChain(req)
	validators := newMockValidators(r, res.Validators, params)

	return validators, genesisTimestamp, accounts, chainID
}

// SimulateFromSeed tests an application by running the provided
// operations, testing the provided invariants, but using the provided config.Seed.
// TODO: split this monster function up
func SimulateFromSeed(
	tb testing.TB,
	w io.Writer,
	app *baseapp.BaseApp,
	appStateFn simulation.AppStateFn,
	randAccFn simulation.RandomAccountFn,
	ops WeightedOperations,
	blockedAddrs map[string]bool,
	config simulation.Config,
	cdc codec.JSONCodec,
) (stopEarly bool, exportedParams Params, err error) {
	// in case we have to end early, don't os.Exit so that we can run cleanup code.
	testingMode, _, b := getTestingMode(tb)

	fmt.Fprintf(w, "Starting SimulateFromSeed with randomness created with seed %d\n", int(config.Seed))
	r := rand.New(rand.NewSource(config.Seed))
	simParams := RandomParams(r)
	fmt.Fprintf(w, "Randomized simulation params: \n%s\n", mustMarshalJSONIndent(simParams))

	accs := randAccFn(r, simParams.NumKeys())
	if len(accs) == 0 {
		return true, simParams, fmt.Errorf("must have greater than zero genesis accounts")
	}

	// Second variable to keep pending validator set
	// Recall that validator set updates are delayed one block by Tendermint,
	// e.g. that block validator update in end block of N, only goes into affect begin block of N + 2.
	// Initially this is the same as the initial validator set
	validators, genesisTimestamp, accs, chainID := initChain(r, simParams, accs, app, appStateFn, config, cdc)

	config.ChainID = chainID

	fmt.Printf(
		"Starting the simulation from time %v (unixtime %v)\n",
		genesisTimestamp.UTC().Format(time.UnixDate), genesisTimestamp.Unix(),
	)

	// remove module account address if they exist in accs
	var tmpAccs []simulation.Account

	for _, acc := range accs {
		if !blockedAddrs[acc.Address.String()] {
			tmpAccs = append(tmpAccs, acc)
		}
	}

	accs = tmpAccs
	initialHeader := tmproto.Header{
		ChainID:         config.ChainID,
		Height:          int64(config.InitialBlockHeight),
		Time:            genesisTimestamp,
		ProposerAddress: validators.randomProposer(r),
	}

	simState := newSimulatorState(simParams, initialHeader, tb, w, validators).WithLogParam(config.Lean)

	simCtx := simtypes.NewSimCtx(r, app, accs, simState.header.ChainID)

	// Setup code to catch SIGTERM's
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		receivedSignal := <-c
		fmt.Fprintf(w, "\nExiting early due to %s, on block %d, operation %d\n", receivedSignal, simState.header.Height, simState.opCount)
		err = fmt.Errorf("exited due to %s", receivedSignal)
		stopEarly = true
	}()

	blockSimulator := createBlockSimulator(testingMode, w, simParams, ops, simState, config)

	if !testingMode {
		b.ResetTimer()
	} else {
		// recover logs in case of panic
		defer func() {
			if r := recover(); r != nil {
				_, _ = fmt.Fprintf(w, "simulation halted due to panic on block %d\n", simState.header.Height)
				simState.logWriter.PrintLogs()
				panic(r)
			}
		}()
	}

	// set exported params to the initial state
	if config.ExportParamsPath != "" && config.ExportParamsHeight == 0 {
		exportedParams = simParams
	}

	// TODO: split up the contents of this for loop into new functions
	for height := config.InitialBlockHeight; height < config.NumBlocks+config.InitialBlockHeight && !stopEarly; height++ {
		simState.SimulateBlock(simCtx, blockSimulator)

		if config.Commit {
			simCtx.App.Commit()
		}

		// update the exported params
		// TODO: wtf is this for, the params don't update?
		if config.ExportParamsPath != "" && config.ExportParamsHeight == height {
			exportedParams = simParams
		}
	}

	if stopEarly {
		if config.ExportStatsPath != "" {
			fmt.Println("Exporting simulation statistics...")
			simState.eventStats.ExportJSON(config.ExportStatsPath)
		} else {
			simState.eventStats.Print(w)
		}

		return true, exportedParams, err
	}

	fmt.Fprintf(
		w,
		"\nSimulation complete; Final height (blocks): %d, final time (seconds): %v, operations ran: %d\n",
		simState.header.Height, simState.header.Time, simState.opCount,
	)

	if config.ExportStatsPath != "" {
		fmt.Println("Exporting simulation statistics...")
		simState.eventStats.ExportJSON(config.ExportStatsPath)
	} else {
		simState.eventStats.Print(w)
	}

	return false, exportedParams, nil
}

type blockSimFn func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
	accounts []simulation.Account, header tmproto.Header) (opCount int)

// Returns a function to simulate blocks. Written like this to avoid constant
// parameters being passed everytime, to minimize memory overhead.
func createBlockSimulator(testingMode bool, w io.Writer, params Params, ops WeightedOperations,
	simState *simState, config simulation.Config) blockSimFn {

	lastBlockSizeState := 0 // state for [4 * uniform distribution]
	blocksize := 0
	selectOp := ops.getSelectOpFn()

	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accounts []simulation.Account, header tmproto.Header,
	) (opCount int) {
		_, _ = fmt.Fprintf(
			w, "\rSimulating... block %d/%d, operation %d/%d.",
			header.Height, config.NumBlocks, opCount, blocksize,
		)
		lastBlockSizeState, blocksize = getBlockSize(r, params, lastBlockSizeState, config.BlockSize)

		type opAndR struct {
			op   simulation.Operation
			rand *rand.Rand
		}

		opAndRz := make([]opAndR, 0, blocksize)

		// Predetermine the blocksize slice so that we can do things like block
		// out certain operations without changing the ops that follow.
		for i := 0; i < blocksize; i++ {
			opAndRz = append(opAndRz, opAndR{
				op:   selectOp(r),
				rand: simulation.DeriveRand(r),
			})
		}

		for i := 0; i < blocksize; i++ {
			// NOTE: the Rand 'r' should not be used here.
			opAndR := opAndRz[i]
			op, r2 := opAndR.op, opAndR.rand
			opMsg, futureOps, err := op(r2, app, ctx, accounts, config.ChainID)
			opMsg.LogEvent(simState.eventStats.Tally)

			if !config.Lean || opMsg.OK {
				simState.logWriter.AddEntry(MsgEntry(header.Height, int64(i), opMsg))
			}

			if err != nil {
				simState.logWriter.PrintLogs()
				simState.tb.Fatalf(`error on block  %d/%d, operation (%d/%d) from x/%s:
%v
Comment: %s`,
					header.Height, config.NumBlocks, opCount, blocksize, opMsg.Route, err, opMsg.Comment)
			}

			queueOperations(simState.operationQueue, simState.timeOperationQueue, futureOps)

			if testingMode && opCount%50 == 0 {
				fmt.Fprintf(w, "\rSimulating... block %d/%d, operation %d/%d. ",
					header.Height, config.NumBlocks, opCount, blocksize)
			}

			opCount++
		}

		return opCount
	}
}

// nolint: errcheck
func runQueuedOperations(queueOps map[int][]simulation.Operation,
	height int, tb testing.TB, r *rand.Rand, app *baseapp.BaseApp,
	ctx sdk.Context, accounts []simulation.Account, logWriter LogWriter,
	event func(route, op, evResult string), lean bool, chainID string) (numOpsRan int) {

	queuedOp, ok := queueOps[height]
	if !ok {
		return 0
	}

	numOpsRan = len(queuedOp)
	for i := 0; i < numOpsRan; i++ {

		// For now, queued operations cannot queue more operations.
		// If a need arises for us to support queued messages to queue more messages, this can
		// be changed.
		opMsg, _, err := queuedOp[i](r, app, ctx, accounts, chainID)
		opMsg.LogEvent(event)

		if !lean || opMsg.OK {
			logWriter.AddEntry((QueuedMsgEntry(int64(height), opMsg)))
		}

		if err != nil {
			logWriter.PrintLogs()
			tb.FailNow()
		}
	}
	delete(queueOps, height)

	return numOpsRan
}

func runQueuedTimeOperations(queueOps []simulation.FutureOperation,
	height int, currentTime time.Time, tb testing.TB, r *rand.Rand,
	app *baseapp.BaseApp, ctx sdk.Context, accounts []simulation.Account,
	logWriter LogWriter, event func(route, op, evResult string),
	lean bool, chainID string) (numOpsRan int) {

	numOpsRan = 0
	for len(queueOps) > 0 && currentTime.After(queueOps[0].BlockTime) {

		// For now, queued operations cannot queue more operations.
		// If a need arises for us to support queued messages to queue more messages, this can
		// be changed.
		opMsg, _, err := queueOps[0].Op(r, app, ctx, accounts, chainID)
		opMsg.LogEvent(event)

		if !lean || opMsg.OK {
			logWriter.AddEntry(QueuedMsgEntry(int64(height), opMsg))
		}

		if err != nil {
			logWriter.PrintLogs()
			tb.FailNow()
		}

		queueOps = queueOps[1:]
		numOpsRan++
	}

	return numOpsRan
}
