package simulation

import (
	"encoding/json"
	"fmt"
	"testing"

	"io/ioutil"

	osmosim "github.com/osmosis-labs/osmosis/v11/simulation/simtypes"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func getTestingMode(tb testing.TB) (testingMode bool, t *testing.T, b *testing.B) {
	testingMode = false

	if _t, ok := tb.(*testing.T); ok {
		t = _t
		testingMode = true
	} else {
		//nolint:forcetypeassert
		b = tb.(*testing.B)
	}

	return testingMode, t, b
}

// getBlockSize returns a block size as determined from the transition matrix.
// It targets making average block size the provided parameter. The three
// states it moves between are:
//  - "over stuffed" blocks with average size of 2 * avgblocksize,
//  - normal sized blocks, hitting avgBlocksize on average,
//  - and empty blocks, with no txs / only txs scheduled from the past.
func getBlockSize(simCtx *osmosim.SimCtx, params Params, lastBlockSizeState, avgBlockSize int) (state, blockSize int) {
	r := simCtx.GetSeededRand("get block size")
	// TODO: Make default blocksize transition matrix actually make the average
	// blocksize equal to avgBlockSize.
	state = params.BlockSizeTransitionMatrix().NextState(r, lastBlockSizeState)

	switch state {
	case 0:
		blockSize = r.Intn(avgBlockSize * 4)

	case 1:
		blockSize = r.Intn(avgBlockSize * 2)

	default:
		blockSize = 0
	}

	return state, blockSize
}

func mustMarshalJSONIndent(o interface{}) []byte {
	bz, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("failed to JSON encode: %s", err))
	}

	return bz
}

// SetupSimulation creates the config, db (levelDB), temporary directory and logger for
// the simulation tests. If `FlagEnabledValue` is false it skips the current test.
// Returns error on an invalid db intantiation or temp dir creation.
func SetupSimulation(dirPrefix, dbName string) (simtypes.Config, dbm.DB, string, log.Logger, bool, error) {
	if !FlagEnabledValue {
		return simtypes.Config{}, nil, "", nil, true, nil
	}

	config := NewConfigFromFlags()
	config.ChainID = helpers.SimAppChainID

	var logger log.Logger
	if FlagVerboseValue {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}

	dir, err := ioutil.TempDir("", dirPrefix)
	if err != nil {
		return simtypes.Config{}, nil, "", nil, false, err
	}

	db, err := sdk.NewLevelDB(dbName, dir)
	if err != nil {
		return simtypes.Config{}, nil, "", nil, false, err
	}

	return config, db, dir, logger, false, nil
}

// SimulationOperations retrieves the simulation params from the provided file path
// and returns all the modules weighted operations
func SimulationOperations(app simapp.App, cdc codec.JSONCodec, config simtypes.Config) []simtypes.WeightedOperation {
	simState := module.SimulationState{
		AppParams: make(simtypes.AppParams),
		Cdc:       cdc,
	}

	if config.ParamsFile != "" {
		bz, err := ioutil.ReadFile(config.ParamsFile)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(bz, &simState.AppParams)
		if err != nil {
			panic(err)
		}
	}

	simState.ParamChanges = app.SimulationManager().GenerateParamChanges(config.Seed)
	simState.Contents = app.SimulationManager().GetProposalContents(simState)
	return app.SimulationManager().WeightedOperations(simState)
}

// CheckExportSimulation exports the app state and simulation parameters to JSON
// if the export paths are defined.
func CheckExportSimulation(
	app simapp.App, config simtypes.Config, params simtypes.Params,
) error {
	if config.ExportStatePath != "" {
		fmt.Println("exporting app state...")
		exported, err := app.ExportAppStateAndValidators(false, nil)
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(config.ExportStatePath, []byte(exported.AppState), 0600); err != nil {
			return err
		}
	}

	if config.ExportParamsPath != "" {
		fmt.Println("exporting simulation params...")
		paramsBz, err := json.MarshalIndent(params, "", " ")
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(config.ExportParamsPath, paramsBz, 0600); err != nil {
			return err
		}
	}
	return nil
}

// PrintStats prints the corresponding statistics from the app DB.
func PrintStats(db dbm.DB) {
	fmt.Println("\nLevelDB Stats")
	fmt.Println(db.Stats()["leveldb.stats"])
	fmt.Println("LevelDB cached block size", db.Stats()["leveldb.cachedblock"])
}

// GetSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func GetSimulationLog(storeName string, sdr sdk.StoreDecoderRegistry, kvAs, kvBs []kv.Pair) (log string) {
	for i := 0; i < len(kvAs); i++ {
		if len(kvAs[i].Value) == 0 && len(kvBs[i].Value) == 0 {
			// skip if the value doesn't have any bytes
			continue
		}

		decoder, ok := sdr[storeName]
		if ok {
			log += decoder(kvAs[i], kvBs[i])
		} else {
			log += fmt.Sprintf("store A %X => %X\nstore B %X => %X\n", kvAs[i].Key, kvAs[i].Value, kvBs[i].Key, kvBs[i].Value)
		}
	}

	return log
}
