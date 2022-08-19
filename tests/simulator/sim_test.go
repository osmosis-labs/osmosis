package simapp

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/v11/app"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	osmosim "github.com/osmosis-labs/osmosis/v11/simulation/executor"
	"github.com/osmosis-labs/osmosis/v11/simulation/simtypes/simlogger"
)

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ github.com/osmosis-labs/osmosis/simapp -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	// -Enabled=true -NumBlocks=1000 -BlockSize=200 \
	// -Period=1 -Commit=true -Seed=57 -v -timeout 24h
	osmosim.FlagEnabledValue = true
	osmosim.FlagNumBlocksValue = 1000
	osmosim.FlagBlockSizeValue = 200
	osmosim.FlagCommitValue = true
	osmosim.FlagVerboseValue = true
	// osmosim.FlagPeriodValue = 1000
	fullAppSimulation(b, false)
}

func TestFullAppSimulation(t *testing.T) {
	// -Enabled=true -NumBlocks=1000 -BlockSize=200 \
	// -Period=1 -Commit=true -Seed=57 -v -timeout 24h
	osmosim.FlagEnabledValue = true
	osmosim.FlagNumBlocksValue = 200
	osmosim.FlagBlockSizeValue = 25
	osmosim.FlagCommitValue = true
	osmosim.FlagVerboseValue = true
	osmosim.FlagPeriodValue = 10
	osmosim.FlagSeedValue = 11
	fullAppSimulation(t, true)
}

func fullAppSimulation(tb testing.TB, is_testing bool) {
	config, db, dir, logger, _, err := osmosim.SetupSimulation("goleveldb-app-sim", "Simulation")
	if err != nil {
		tb.Fatalf("simulation setup failed: %s", err.Error())
	}
	logger = simlogger.NewSimLogger(logger)
	// This file is needed to provide the correct path
	// to reflect.wasm test file needed for wasmd simulation testing.
	config.ParamsFile = "params.json"

	defer func() {
		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			tb.Fatal(err)
		}
	}()

	// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
	// an IAVLStore for faster simulation speed.
	fauxMerkleModeOpt := func(bapp *baseapp.BaseApp) {
		if is_testing {
			bapp.SetFauxMerkleMode()
		}
	}

	osmosis := app.NewOsmosisApp(
		logger,
		db,
		nil,
		true, // load latest
		map[int64]bool{},
		app.DefaultNodeHome,
		osmosim.FlagPeriodValue,
		app.MakeEncodingConfig(),
		osmosim.EmptyAppOptions{},
		app.GetWasmEnabledProposals(),
		app.EmptyWasmOpts,
		interBlockCacheOpt(),
		fauxMerkleModeOpt)

	initFns := osmosim.InitFunctions{
		RandomAccountFn:   osmosim.WrapRandAccFnForResampling(simulation.RandomAccounts, osmosis.ModuleAccountAddrs()),
		AppInitialStateFn: AppStateFn(osmosis.AppCodec(), osmosis.SimulationManager()),
	}

	// Run randomized simulation:
	_, simErr := osmosim.SimulateFromSeed(
		tb,
		os.Stdout,
		osmosis,
		initFns,
		osmosis.SimulationManager().Actions(config.Seed, osmosis.AppCodec()), // Run all registered operations
		config,
	)

	if simErr != nil {
		tb.Fatal(simErr)
	}

	if config.Commit {
		osmosim.PrintStats(db)
	}
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

// TODO: Make another test for the fuzzer itself, which just has noOp txs
// and doesn't depend on the application.
func TestAppStateDeterminism(t *testing.T) {
	// if !osmosim.FlagEnabledValue {
	// 	t.Skip("skipping application simulation")
	// }

	config := osmosim.NewConfigFromFlags()
	config.ExportParamsPath = ""
	config.NumBlocks = 50
	config.BlockSize = 5
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = helpers.SimAppChainID
	config.Commit = true

	// This file is needed to provide the correct path
	// to reflect.wasm test file needed for wasmd simulation testing.
	config.ParamsFile = "params.json"

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]string, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			logger = simlogger.NewSimLogger(log.TestingLogger())
			// if osmosim.FlagVerboseValue {
			// 	logger = log.TestingLogger()
			// } else {
			// 	logger = log.NewNopLogger()
			// }

			db := dbm.NewMemDB()
			osmosis := app.NewOsmosisApp(
				logger,
				db,
				nil,
				true,
				map[int64]bool{},
				app.DefaultNodeHome,
				osmosim.FlagPeriodValue,
				app.MakeEncodingConfig(),
				osmosim.EmptyAppOptions{},
				app.GetWasmEnabledProposals(),
				app.EmptyWasmOpts,
				interBlockCacheOpt())

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			initFns := osmosim.InitFunctions{
				RandomAccountFn:   osmosim.WrapRandAccFnForResampling(simulation.RandomAccounts, osmosis.ModuleAccountAddrs()),
				AppInitialStateFn: AppStateFn(osmosis.AppCodec(), osmosis.SimulationManager()),
			}

			// Run randomized simulation:
			_, simErr := osmosim.SimulateFromSeed(
				t,
				os.Stdout,
				osmosis,
				initFns,
				osmosis.SimulationManager().Actions(config.Seed, osmosis.AppCodec()), // Run all registered operations
				config,
			)

			require.NoError(t, simErr)

			appHash := osmosis.LastCommitID().Hash
			appHashList[j] = fmt.Sprintf("%X", appHash)

			if j != 0 {
				require.Equal(
					t, appHashList[0], appHashList[j],
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}
