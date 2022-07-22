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

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/osmosis-labs/osmosis/v10/simulation/simtypes"
)

const AverageBlockTime = 6 * time.Second

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
// OR: Could be a struct or something with options,
//     to give caller ability to step through / instrument benchmarking if they wanted to, and add a cleanup function.
func SimulateFromSeed(
	tb testing.TB,
	w io.Writer,
	app simtypes.App,
	initFunctions simtypes.InitFunctions,
	actions []simtypes.ActionsWithMetadata,
	config simulation.Config,
) (stopEarly bool, err error) {
	// in case we have to end early, don't os.Exit so that we can run cleanup code.
	// TODO: Understand exit pattern, this is so screwed up. Then delete ^

	// Encapsulate the bizarre initialization logic that must be cleaned.
	simCtx, simState, simParams, err := cursedInitializationLogic(tb, w, app, initFunctions, &config)
	if err != nil {
		return true, err
	}

	// Setup code to catch SIGTERM's
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		receivedSignal := <-c
		fmt.Fprintf(w, "\nExiting early due to %s, on block %d, operation %d\n", receivedSignal, simState.header.Height, simState.opCount)
		err = fmt.Errorf("exited due to %s", receivedSignal)
		stopEarly = true
	}()

	testingMode, _, b := getTestingMode(tb)
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

	stopEarly = simState.SimulateAllBlocks(w, simCtx, blockSimulator, config)

	simState.eventStats.exportEvents(config.ExportStatsPath, w)
	return stopEarly, nil
}

// The goal of this function is to group the extremely badly abstracted genesis logic,
// into a single function we can target continuing to improve / abstract better.
// It outputs SimCtx and SimState which are "cleaner" interface abstractions for the rest of the simulator.
// It also outputs SimParams which is not great.
// It also can modify config.
func cursedInitializationLogic(
	tb testing.TB,
	w io.Writer,
	app simtypes.App,
	initFunctions simtypes.InitFunctions,
	config *simulation.Config) (*simtypes.SimCtx, *simState, Params, error) {
	fmt.Fprintf(w, "Starting SimulateFromSeed with randomness created with seed %d\n", int(config.Seed))

	r := rand.New(rand.NewSource(config.Seed))
	simParams := RandomParams(r)
	fmt.Fprintf(w, "Randomized simulation params: \n%s\n", mustMarshalJSONIndent(simParams))

	accs := initFunctions.RandomAccountFn(r, simParams.NumKeys())
	if len(accs) == 0 {
		return nil, nil, simParams, fmt.Errorf("must have greater than zero genesis accounts")
	}

	validators, genesisTimestamp, accs := initChain(r, simParams, accs, app, initFunctions.AppInitialStateFn, config)

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

	// TODO: If simulation has a param export path configured, export params here.

	return simCtx, simState, simParams, nil
}

// initialize the chain for the simulation
func initChain(
	r *rand.Rand,
	params Params,
	accounts []simulation.Account,
	app simtypes.App,
	appStateFn simulation.AppStateFn,
	config *simulation.Config,
) (mockValidators, time.Time, []simulation.Account) {
	// TODO: Cleanup the whole config dependency with appStateFn
	appState, accounts, chainID, genesisTimestamp := appStateFn(r, accounts, *config)
	consensusParams := randomConsensusParams(r, appState, app.AppCodec())
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
func createBlockSimulator(testingMode bool, w io.Writer, params Params, actions []simtypes.ActionsWithMetadata,
	simState *simState, config simulation.Config,
) blockSimFn {
	lastBlockSizeState := 0 // state for [4 * uniform distribution]
	blocksize := 0
	selectAction := simtypes.GetSelectActionFn(actions)

	return func(
		simCtx *simtypes.SimCtx, ctx sdk.Context, header tmproto.Header,
	) (opCount int) {
		_, _ = fmt.Fprintf(
			w, "\rSimulating... block %d/%d, operation 0/%d.",
			header.Height, config.NumBlocks, blocksize,
		)
		lastBlockSizeState, blocksize = getBlockSize(simCtx, params, lastBlockSizeState, config.BlockSize)

		blockNumStr := fmt.Sprintf("block %d", header.Height)
		for i := 0; i < blocksize; i++ {
			// Sample and execute every action using independent randomness.
			// Thus any change within one action's randomness won't waterfall
			// to every other action and the overall order of txs.
			// We can also use this to limit which operations we run, in debugging a simulator run.
			actionSeed := fmt.Sprintf("%s operation %d", blockNumStr, i)
			actionSimCtx, cleanup := simCtx.WrapRand(actionSeed)

			// Select and execute tx
			action := selectAction(actionSimCtx.GetSeededRand("action select"))
			opMsg, futureOps, err := action.Execute(actionSimCtx, ctx)
			opMsg.Route = action.ModuleName
			cleanup()

			simState.logActionResult(header, i, config, blocksize, opMsg, err)

			simState.queueOperations(futureOps)

			if testingMode && i%50 == 0 {
				fmt.Fprintf(w, "\rSimulating... block %d/%d, operation %d/%d. ",
					header.Height, config.NumBlocks, i, blocksize)
			}
		}

		return blocksize
	}
}

// This is inheriting old functionality. We should break this as part of making logging be usable / make sense.
func (simState *simState) logActionResult(
	header tmproto.Header, actionIndex int, config simulation.Config, blocksize int,
	opMsg simulation.OperationMsg, actionErr error) {
	opMsg.LogEvent(simState.eventStats.Tally)
	if !simState.leanLogs || opMsg.OK {
		simState.logWriter.AddEntry(MsgEntry(header.Height, int64(actionIndex), opMsg))
	}

	if actionErr != nil {
		simState.logWriter.PrintLogs()
		simState.tb.Fatalf(`error on block  %d/%d, operation (%d/%d) from x/%s:
%v
Comment: %s`,
			header.Height, config.NumBlocks, actionIndex, blocksize, opMsg.Route, actionErr, opMsg.Comment)
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
		opMsg, _, err := queuedOp[i](r, simCtx.BaseApp(), ctx, simCtx.Accounts, simCtx.ChainID())
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
