package simulation

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

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
	app simtypes.App,
	appStateFn simulation.AppStateFn,
	config *simulation.Config,
	cdc codec.JSONCodec,
) (mockValidators, time.Time, []simulation.Account) {
	// TODO: Cleanup the whole config dependency with appStateFn
	appState, accounts, chainID, genesisTimestamp := appStateFn(r, accounts, *config)
	consensusParams := randomConsensusParams(r, appState, cdc)
	req := abci.RequestInitChain{
		AppStateBytes:   appState,
		ChainId:         chainID,
		ConsensusParams: consensusParams,
		Time:            genesisTimestamp,
	}
	// Valid app version can only be zero on app initialization.
	req.ConsensusParams.Version.AppVersion = 0
	res := app.GetBaseApp().InitChain(req)
	validators := newMockValidators(r, res.Validators, params)

	// update config
	config.ChainID = chainID
	if config.InitialBlockHeight == 0 {
		config.InitialBlockHeight = 1
	}

	return validators, genesisTimestamp, accounts
}

// SimulateFromSeedLegacy tests an application by running the provided
// operations, testing the provided invariants, but using the provided config.Seed.
// TODO: Restore SimulateFromSeedLegacy by adding a wrapper that can take in
// func SimulateFromSeedLegacy(
// 	tb testing.TB,
// 	w io.Writer,
// 	app *baseapp.BaseApp,
// 	appStateFn simulation.AppStateFn,
// 	randAccFn simulation.RandomAccountFn,
// 	ops legacysimexec.WeightedOperations,
// 	blockedAddrs map[string]bool,
// 	config simulation.Config,
// 	cdc codec.JSONCodec,
// ) (stopEarly bool, exportedParams Params, err error) {
// 	actions := simtypes.ActionsFromWeightedOperations(ops)
// 	initFns := simtypes.InitFunctions{
// 		RandomAccountFn:   simtypes.WrapRandAccFnForResampling(randAccFn, blockedAddrs),
// 		AppInitialStateFn: appStateFn,
// 	}
// 	return SimulateFromSeed(tb, w, app, initFns, actions, config, cdc)
// }

// SimulateFromSeed tests an application by running the provided
// operations, testing the provided invariants, but using the provided config.Seed.
// TODO: Inputs should be:
// * SimManager for module configs
// * Config file for params
// * whatever is needed for logging (tb + w rn)
func SimulateFromSeed(
	tb testing.TB,
	w io.Writer,
	app simtypes.App,
	initFunctions simtypes.InitFunctions,
	actions []simtypes.Action,
	config simulation.Config,
	cdc codec.JSONCodec,
) (stopEarly bool, exportedParams Params, err error) {
	// in case we have to end early, don't os.Exit so that we can run cleanup code.
	testingMode, _, b := getTestingMode(tb)

	fmt.Fprintf(w, "Starting SimulateFromSeed with randomness created with seed %d\n", int(config.Seed))
	r := rand.New(rand.NewSource(config.Seed))
	simParams := RandomParams(r)
	fmt.Fprintf(w, "Randomized simulation params: \n%s\n", mustMarshalJSONIndent(simParams))

	accs := initFunctions.RandomAccountFn(r, simParams.NumKeys())
	if len(accs) == 0 {
		return true, simParams, fmt.Errorf("must have greater than zero genesis accounts")
	}

	validators, genesisTimestamp, accs := initChain(r, simParams, accs, app, initFunctions.AppInitialStateFn, &config, cdc)

	fmt.Printf(
		"Starting the simulation from time %v (unixtime %v)\n",
		genesisTimestamp.UTC().Format(time.UnixDate), genesisTimestamp.Unix(),
	)

	simCtx := simtypes.NewSimCtx(r, app, accs, config.ChainID)

	initialHeader := tmproto.Header{
		ChainID:         config.ChainID,
		Height:          int64(config.InitialBlockHeight),
		Time:            genesisTimestamp,
		ProposerAddress: validators.randomProposer(r),
	}

	simState := newSimulatorState(simParams, initialHeader, tb, w, validators).WithLogParam(config.Lean)

	// Setup code to catch SIGTERM's
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		receivedSignal := <-c
		fmt.Fprintf(w, "\nExiting early due to %s, on block %d, operation %d\n", receivedSignal, simState.header.Height, simState.opCount)
		err = fmt.Errorf("exited due to %s", receivedSignal)
		stopEarly = true
	}()

	blockSimulator := createBlockSimulator(testingMode, w, simParams, actions, simState, config)

	if !testingMode {
		b.ResetTimer()
	} else {
		// recover logs in case of panic
		defer func() {
			if r := recover(); r != nil {
				// TODO: Come back and cleanup the entire panic recovery logging.
				// printPanicRecoveryError(r)
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

	for height := config.InitialBlockHeight; height < config.NumBlocks+config.InitialBlockHeight && !stopEarly; height++ {
		stopEarly = simState.SimulateBlock(simCtx, blockSimulator)
		if stopEarly {
			break
		}

		if config.Commit {
			simCtx.App.GetBaseApp().Commit()
		}
	}

	if !stopEarly {
		fmt.Fprintf(
			w,
			"\nSimulation complete; Final height (blocks): %d, final time (seconds): %v, operations ran: %d\n",
			simState.header.Height, simState.header.Time, simState.opCount,
		)
		simState.logWriter.PrintLogs()
	}

	simState.eventStats.exportEvents(config.ExportStatsPath, w)
	return stopEarly, exportedParams, nil
}

//nolint:deadcode,unused
func printPanicRecoveryError(recoveryError interface{}) {
	errStackTrace := string(debug.Stack())
	switch e := recoveryError.(type) {
	case string:
		fmt.Println("Recovering from (string) panic: " + e)
	case runtime.Error:
		fmt.Println("recovered (runtime.Error) panic: " + e.Error())
	case error:
		fmt.Println("recovered (error) panic: " + e.Error())
	default:
		fmt.Println("recovered (default) panic. Could not capture logs in ctx, see stdout")
		fmt.Println("Recovering from panic ", recoveryError)
		debug.PrintStack()
		return
	}
	fmt.Println("stack trace: " + errStackTrace)
}

type blockSimFn func(simCtx *simtypes.SimCtx, ctx sdk.Context, header tmproto.Header) (opCount int)

// Returns a function to simulate blocks. Written like this to avoid constant
// parameters being passed everytime, to minimize memory overhead.
func createBlockSimulator(testingMode bool, w io.Writer, params Params, actions []simtypes.Action,
	simState *simState, config simulation.Config,
) blockSimFn {
	lastBlockSizeState := 0 // state for [4 * uniform distribution]
	blocksize := 0
	selectAction := getSelectActionFn(actions)

	return func(
		simCtx *simtypes.SimCtx, ctx sdk.Context, header tmproto.Header,
	) (opCount int) {
		_, _ = fmt.Fprintf(
			w, "\rSimulating... block %d/%d, operation %d/%d.",
			header.Height, config.NumBlocks, opCount, blocksize,
		)
		lastBlockSizeState, blocksize = getBlockSize(simCtx, params, lastBlockSizeState, config.BlockSize)

		// TODO: Fix according to the r plans
		// Predetermine the blocksize slice so that we can do things like block
		// out certain operations without changing the ops that follow.
		// NOTE: This is poor mans seeding, it will improve in our simctx plans =)
		blockActions := make([]simtypes.Action, 0, blocksize)
		for i := 0; i < blocksize; i++ {
			blockActions = append(blockActions, selectAction(simCtx.GetRand()))
		}

		for i := 0; i < blocksize; i++ {
			action := blockActions[i]
			// TODO: We need to make a simCtx.WithSeededRand, that replaces the rand map internally
			// but allows updates to accounts.
			opMsg, futureOps, err := action.Execute(simCtx, ctx)
			opMsg.LogEvent(simState.eventStats.Tally)

			if !simState.leanLogs || opMsg.OK {
				simState.logWriter.AddEntry(MsgEntry(header.Height, int64(i), opMsg))
			}

			if err != nil {
				simState.logWriter.PrintLogs()
				simState.tb.Fatalf(`error on block  %d/%d, operation (%d/%d) from x/%s:
%v
Comment: %s`,
					header.Height, config.NumBlocks, opCount, blocksize, opMsg.Route, err, opMsg.Comment)
			}

			simState.queueOperations(futureOps)

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
func (simState *simState) runQueuedOperations(simCtx *simtypes.SimCtx, ctx sdk.Context) (numOpsRan int) {
	height := int(simState.header.Height)
	queuedOp, ok := simState.operationQueue[height]
	if !ok {
		return 0
	}

	numOpsRan = len(queuedOp)
	for i := 0; i < numOpsRan; i++ {
		// TODO: Fix according to the r plans
		r := simCtx.GetRand()

		// For now, queued operations cannot queue more operations.
		// If a need arises for us to support queued messages to queue more messages, this can
		// be changed.
		opMsg, _, err := queuedOp[i](r, simCtx.App.GetBaseApp(), ctx, simCtx.Accounts, simCtx.ChainID)
		opMsg.LogEvent(simState.eventStats.Tally)

		if !simState.leanLogs || opMsg.OK {
			simState.logWriter.AddEntry((QueuedMsgEntry(int64(height), opMsg)))
		}

		if err != nil {
			simState.logWriter.PrintLogs()
			simState.tb.Fatalf(`error on block  %d, height queued operation (%d/%d) from x/%s:
%v
Comment: %s`,
				simState.header.Height, i, numOpsRan, opMsg.Route, err, opMsg.Comment)
			simState.tb.FailNow()
		}
	}
	delete(simState.operationQueue, height)

	return numOpsRan
}

func (simState *simState) runQueuedTimeOperations(simCtx *simtypes.SimCtx, ctx sdk.Context) (
	numOpsRan int,
) {
	// TODO: Refactor this to gather time queue ops, then execute them.
	queueOps := simState.timeOperationQueue
	currentTime := simState.header.Time
	numOpsRan = 0
	for len(queueOps) > 0 && currentTime.After(queueOps[0].BlockTime) {
		// TODO: Fix according to the r plans
		r := simCtx.GetRand()

		// For now, queued operations cannot queue more operations.
		// If a need arises for us to support queued messages to queue more messages, this can
		// be changed.
		opMsg, _, err := queueOps[0].Op(r, simCtx.App.GetBaseApp(), ctx, simCtx.Accounts, simCtx.ChainID)
		opMsg.LogEvent(simState.eventStats.Tally)

		if !simState.leanLogs || opMsg.OK {
			simState.logWriter.AddEntry(QueuedMsgEntry(simState.header.Height, opMsg))
		}

		if err != nil {
			simState.logWriter.PrintLogs()
			simState.tb.Fatalf(`error on block  %d, time queued operation (x/x) from x/%s:
			%v
			Comment: %s`,
				simState.header.Height, opMsg.Route, err, opMsg.Comment)
			simState.tb.FailNow()
		}

		queueOps = queueOps[1:]
		numOpsRan++
	}
	simState.timeOperationQueue = queueOps
	return numOpsRan
}
