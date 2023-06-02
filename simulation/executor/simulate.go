package simulation

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"
	"testing"
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/osmosis-labs/osmosis/v16/simulation/executor/internal/executortypes"
	"github.com/osmosis-labs/osmosis/v16/simulation/executor/internal/stats"
	"github.com/osmosis-labs/osmosis/v16/simulation/simtypes"
)

const AverageBlockTime = 6 * time.Second

// SimulateFromSeed tests an application by running the provided
// operations, testing the provided invariants, but using the provided config.Seed.
// TODO: Inputs should be:
// * SimManager for module configs
// * Config file for params
// * whatever is needed for logging (tb + w rn)
// OR:
// * Could be a struct or something with options,
// to give caller ability to step through / instrument benchmarking if they
// wanted to, and add a cleanup function.
func SimulateFromSeed(
	tb testing.TB,
	w io.Writer,
	appCreator simtypes.AppCreator,
	initFunctions InitFunctions,
	config Config,
) (lastCommitId storetypes.CommitID, stopEarly bool, err error) {
	tb.Helper()
	// in case we have to end early, don't os.Exit so that we can run cleanup code.
	// TODO: Understand exit pattern, this is so screwed up. Then delete ^

	legacyInvariantPeriod := uint(10) // TODO: Make a better answer of what to do here, at minimum put into config
	app := appCreator(simulationHomeDir(), legacyInvariantPeriod, baseappOptionsFromConfig(config)...)
	simManager := executortypes.CreateSimulationManager(app)
	actions := simManager.Actions(config.Seed, app.AppCodec())

	// Set up sql table
	statsDb, err := stats.SetupStatsDb(config.ExportConfig)
	if err != nil {
		tb.Fatal(err)
	}
	defer statsDb.Cleanup()

	// Encapsulate the bizarre initialization logic that must be cleaned.
	simCtx, simState, simParams, err := cursedInitializationLogic(tb, w, app, simManager, initFunctions, &config)
	if err != nil {
		return storetypes.CommitID{}, true, err
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
	blockSimulator := createBlockSimulator(testingMode, w, simParams, actions, simState, config, statsDb)

	if !testingMode {
		b.ResetTimer()
	}
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

	stopEarly, err = simState.SimulateAllBlocks(w, simCtx, blockSimulator)

	simState.eventStats.ExportEvents(config.ExportConfig.ExportStatsPath, w)
	return storetypes.CommitID{}, stopEarly, err
}

func simulationHomeDir() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(userHomeDir, ".osmosis_simulation")
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
	simManager executortypes.Manager,
	initFunctions InitFunctions,
	config *Config,
) (*simtypes.SimCtx, *simState, Params, error) {
	tb.Helper()
	fmt.Fprintf(w, "Starting SimulateFromSeed with randomness created with seed %d\n", int(config.Seed))

	r := rand.New(rand.NewSource(config.Seed))
	simParams := RandomParams(r)
	fmt.Fprintf(w, "Randomized simulation params: \n%s\n", mustMarshalJSONIndent(simParams))

	accs := initFunctions.RandomAccountFn(r, simParams.NumKeys())
	if len(accs) == 0 {
		return nil, nil, simParams, fmt.Errorf("must have greater than zero genesis accounts")
	}

	validators, genesisTimestamp, accs, res := initChain(
		simManager, r, simParams, accs, app, initFunctions.InitChainFn, config)

	fmt.Printf(
		"Starting the simulation from time %v (unixtime %v)\n",
		genesisTimestamp.UTC().Format(time.UnixDate), genesisTimestamp.Unix(),
	)

	simCtx := simtypes.NewSimCtx(r, app, accs, config.InitializationConfig.ChainID)

	// TODO: Understand how this works better in Tendermint wrt
	// genesis timestamp and proposer for first block
	initialHeader := tmproto.Header{
		ChainID:         config.InitializationConfig.ChainID,
		Height:          int64(config.InitializationConfig.InitialBlockHeight),
		Time:            genesisTimestamp,
		ProposerAddress: validators.randomProposer(r).Address(),
		AppHash:         res.AppHash,
	}

	// must set version in order to generate hashes
	initialHeader.Version.Block = 11

	simState := newSimulatorState(tb, simParams, initialHeader, w, validators, *config)

	// TODO: If simulation has a param export path configured, export params here.

	return simCtx, simState, simParams, nil
}

// initialize the chain for the simulation
func initChain(
	simManager executortypes.Manager,
	r *rand.Rand,
	params Params,
	accounts []simulation.Account,
	app simtypes.App,
	initChainFn InitChainFn,
	config *Config,
) (mockValidators, time.Time, []simulation.Account, abci.ResponseInitChain) {
	// TODO: Cleanup the whole config dependency with appStateFn
	accounts, req := initChainFn(simManager, r, accounts, config.InitializationConfig)
	// Valid app version can only be zero on app initialization.
	req.ConsensusParams.Version.AppVersion = 0
	res := app.GetBaseApp().InitChain(req)
	validators := newMockValidators(r, res.Validators, params)

	// update config
	config.InitializationConfig.ChainID = req.ChainId
	if config.InitializationConfig.InitialBlockHeight == 0 {
		config.InitializationConfig.InitialBlockHeight = 1
	}

	return validators, req.Time, accounts, res
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

type blockSimFn func(simCtx *simtypes.SimCtx, ctx sdk.Context, header tmproto.Header) (opCount int, err error)

// Returns a function to simulate blocks. Written like this to avoid constant
// parameters being passed everytime, to minimize memory overhead.
func createBlockSimulator(testingMode bool, w io.Writer, params Params, actions []simtypes.ActionsWithMetadata,
	simState *simState, config Config, stats stats.StatsDb,
) blockSimFn {
	lastBlockSizeState := 0 // state for [4 * uniform distribution]
	blocksize := 0
	selectAction := executortypes.GetSelectActionFn(actions)

	return func(
		simCtx *simtypes.SimCtx, ctx sdk.Context, header tmproto.Header,
	) (opCount int, err error) {
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
			opMsg, futureOps, resultData, err := action.Execute(actionSimCtx, ctx)

			// add execution result to block's data storage
			simState.Data = append(simState.Data, resultData)
			opMsg.Route = action.ModuleName
			cleanup()

			err = simState.logActionResult(header, i, opMsg, resultData, stats, err)
			if err != nil {
				return opCount, fmt.Errorf("error on block  %d/%d, operation (%d/%d): %w",
					header.Height, config.NumBlocks, i, blocksize, err)
			}

			simState.queueOperations(futureOps)

			if testingMode && i%50 == 0 {
				fmt.Fprintf(w, "\rSimulating... block %d/%d, operation %d/%d. ",
					header.Height, config.NumBlocks, i, blocksize)
			}
		}

		return blocksize, nil
	}
}

// This is inheriting old functionality. We should break this as part of making logging be usable / make sense.
func (simState *simState) logActionResult(
	header tmproto.Header, actionIndex int,
	opMsg simulation.OperationMsg, resultData []byte, stats stats.StatsDb, actionErr error,
) error {
	opMsg.LogEvent(simState.eventStats.Tally)
	err := stats.LogActionResult(header, opMsg, resultData)
	if err != nil {
		return err
	}

	if !simState.config.Lean || opMsg.OK {
		simState.logWriter.AddEntry(MsgEntry(header.Height, int64(actionIndex), opMsg))
	}

	if actionErr != nil {
		simState.logWriter.PrintLogs()
		return fmt.Errorf(`error from x/%s:
%v
Comment: %s`, opMsg.Route, actionErr, opMsg.Comment)
	}
	return nil
}

// TODO: We need to cleanup queued operations, to instead make it queued action + have code re-use with prior code
func (simState *simState) runQueuedOperations(simCtx *simtypes.SimCtx, ctx sdk.Context) (numOpsRan int, err error) {
	height := int(simState.header.Height)
	queuedOp, ok := simState.operationQueue[height]
	if !ok {
		return 0, nil
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

		if !simState.config.Lean || opMsg.OK {
			simState.logWriter.AddEntry((QueuedMsgEntry(int64(height), opMsg)))
		}

		if err != nil {
			simState.logWriter.PrintLogs()
			return 0, fmt.Errorf(`error on block  %d, height queued operation (%d/%d) from x/%s:
%v
Comment: %s`,
				simState.header.Height, i, numOpsRan, opMsg.Route, err, opMsg.Comment)
		}
	}
	delete(simState.operationQueue, height)

	return numOpsRan, nil
}
