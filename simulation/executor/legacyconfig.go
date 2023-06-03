package simulation

import (
	"flag"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/v16/simulation/executor/internal/stats"
	"github.com/osmosis-labs/osmosis/v16/simulation/simtypes/simlogger"
)

// List of available flags for the simulator
var (
	FlagGenesisFileValue        string
	FlagParamsFileValue         string
	FlagExportParamsPathValue   string
	FlagExportParamsHeightValue int
	FlagExportStatePathValue    string
	FlagExportStatsPathValue    string
	FlagSeedValue               int64
	FlagInitialBlockHeightValue int
	FlagNumBlocksValue          int
	FlagBlockSizeValue          int
	FlagLeanValue               bool
	FlagCommitValue             bool
	FlagOnOperationValue        bool // TODO: Remove in favor of binary search for invariant violation
	FlagAllInvariantsValue      bool
	FlagWriteStatsToDB          bool

	FlagEnabledValue     bool
	FlagVerboseValue     bool
	FlagPeriodValue      uint
	FlagGenesisTimeValue int64
)

// EmptyAppOptions is a stub implementing AppOptions
type EmptyAppOptions struct{}

// Get implements AppOptions
func (ao EmptyAppOptions) Get(o string) interface{} {
	return nil
}

// GetSimulatorFlags gets the values of all the available simulation flags
func GetSimulatorFlags() {
	// config fields
	flag.StringVar(&FlagGenesisFileValue, "Genesis", "", "custom simulation genesis file; cannot be used with params file")
	flag.StringVar(&FlagParamsFileValue, "Params", "", "custom simulation params file which overrides any random params; cannot be used with genesis")
	flag.StringVar(&FlagExportParamsPathValue, "ExportParamsPath", "", "custom file path to save the exported params JSON")
	flag.IntVar(&FlagExportParamsHeightValue, "ExportParamsHeight", 0, "height to which export the randomly generated params")
	flag.StringVar(&FlagExportStatePathValue, "ExportStatePath", "", "custom file path to save the exported app state JSON")
	flag.StringVar(&FlagExportStatsPathValue, "ExportStatsPath", "", "custom file path to save the exported simulation statistics JSON")
	flag.Int64Var(&FlagSeedValue, "Seed", 42, "simulation random seed")
	flag.IntVar(&FlagInitialBlockHeightValue, "InitialBlockHeight", 1, "initial block to start the simulation")
	flag.IntVar(&FlagNumBlocksValue, "NumBlocks", 500, "number of new blocks to simulate from the initial block height")
	flag.IntVar(&FlagBlockSizeValue, "BlockSize", 200, "operations per block")
	flag.BoolVar(&FlagLeanValue, "Lean", false, "lean simulation log output")
	flag.BoolVar(&FlagOnOperationValue, "SimulateEveryOperation", false, "run slow invariants every operation")
	flag.BoolVar(&FlagAllInvariantsValue, "PrintAllInvariants", false, "print all invariants if a broken invariant is found")
	flag.BoolVar(&FlagWriteStatsToDB, "WriteStatsToDB", false, "write stats to a local sqlite3 database")

	// simulation flags
	flag.BoolVar(&FlagEnabledValue, "Enabled", false, "enable the simulation")
	flag.BoolVar(&FlagVerboseValue, "Verbose", false, "verbose log output")
	flag.UintVar(&FlagPeriodValue, "Period", 0, "run slow invariants only once every period assertions")
	flag.Int64Var(&FlagGenesisTimeValue, "GenesisTime", 0, "override genesis UNIX time instead of using a random UNIX time")
}

// NewConfigFromFlags creates a simulation from the retrieved values of the flags.
func NewConfigFromFlags() Config {
	return Config{
		InitializationConfig: NewInitializationConfigFromFlags(),
		ExportConfig:         NewExportConfigFromFlags(),
		ExecutionDbConfig:    NewExecutionDbConfigFromFlags(),
		Seed:                 FlagSeedValue,
		NumBlocks:            FlagNumBlocksValue,
		BlockSize:            FlagBlockSizeValue,
		Lean:                 FlagLeanValue,
		OnOperation:          FlagOnOperationValue,
		AllInvariants:        FlagAllInvariantsValue,
	}
}

func NewExportConfigFromFlags() ExportConfig {
	return ExportConfig{
		ExportParamsPath:   FlagExportParamsPathValue,
		ExportParamsHeight: FlagExportParamsHeightValue,
		ExportStatePath:    FlagExportStatePathValue,
		ExportStatsPath:    FlagExportStatsPathValue,
		WriteStatsToDB:     FlagWriteStatsToDB,
	}
}

func NewInitializationConfigFromFlags() InitializationConfig {
	return InitializationConfig{
		GenesisFile:        FlagGenesisFileValue,
		ParamsFile:         FlagParamsFileValue,
		InitialBlockHeight: FlagInitialBlockHeightValue,
	}
}

func NewExecutionDbConfigFromFlags() ExecutionDbConfig {
	return ExecutionDbConfig{
		UseMerkleTree: true,
	}
}

// SetupSimulation creates the config, db (levelDB), temporary directory and logger for
// the simulation tests. If `FlagEnabledValue` is false it skips the current test.
// Returns error on an invalid db intantiation or temp dir creation.
func SetupSimulation(dirPrefix, dbName string) (cfg Config, db dbm.DB, logger log.Logger, cleanup func(), err error) {
	if !FlagEnabledValue {
		return Config{}, nil, nil, func() {}, nil
	}

	config := NewConfigFromFlags()
	config.InitializationConfig.ChainID = helpers.SimAppChainID

	if FlagVerboseValue {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}
	logger = simlogger.NewSimLogger(logger)

	dir, err := os.MkdirTemp("", dirPrefix)
	if err != nil {
		return Config{}, nil, nil, func() {}, err
	}

	db, err = sdk.NewLevelDB(dbName, dir)
	if err != nil {
		return Config{}, nil, nil, func() {}, err
	}

	cleanup = func() {
		db.Close()
		err = os.RemoveAll(dir)
	}

	return config, db, logger, cleanup, nil
}

// PrintStats prints the corresponding statistics from the app DB.
func PrintStats(db dbm.DB) {
	fmt.Println("\nLevelDB Stats")
	fmt.Println(db.Stats()["leveldb.stats"])
	fmt.Println("LevelDB cached block size", db.Stats()["leveldb.cachedblock"])
}

func baseappOptionsFromConfig(config Config) []func(*baseapp.BaseApp) {
	// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
	// an IAVLStore for faster simulation speed.
	fauxMerkleModeOpt := func(bapp *baseapp.BaseApp) {
		if config.ExecutionDbConfig.UseMerkleTree {
			bapp.SetFauxMerkleMode()
		}
	}
	return []func(*baseapp.BaseApp){interBlockCacheOpt(), fauxMerkleModeOpt}
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

type Config struct {
	InitializationConfig InitializationConfig
	ExportConfig         ExportConfig
	ExecutionDbConfig    ExecutionDbConfig

	Seed int64 // simulation random seed

	NumBlocks int // number of new blocks to simulate from the initial block height
	BlockSize int // operations per block

	Lean bool // lean simulation log output

	OnOperation   bool // run slow invariants every operation
	AllInvariants bool // print all failed invariants if a broken invariant is found
}

// Config for how to initialize the simulator state
type InitializationConfig struct {
	GenesisFile        string // custom simulation genesis file; cannot be used with params file
	ParamsFile         string // custom simulation params file which overrides any random params; cannot be used with genesis
	InitialBlockHeight int    // initial block to start the simulation
	ChainID            string // chain-id used on the simulation
}

type ExportConfig = stats.ExportConfig

type ExecutionDbConfig struct {
	UseMerkleTree bool // Use merkle tree underneath, vs using a "fake" merkle tree
}
