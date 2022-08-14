package simulation

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
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
	flag.BoolVar(&FlagCommitValue, "Commit", false, "have the simulation commit")
	flag.BoolVar(&FlagOnOperationValue, "SimulateEveryOperation", false, "run slow invariants every operation")
	flag.BoolVar(&FlagAllInvariantsValue, "PrintAllInvariants", false, "print all invariants if a broken invariant is found")

	// simulation flags
	flag.BoolVar(&FlagEnabledValue, "Enabled", false, "enable the simulation")
	flag.BoolVar(&FlagVerboseValue, "Verbose", false, "verbose log output")
	flag.UintVar(&FlagPeriodValue, "Period", 0, "run slow invariants only once every period assertions")
	flag.Int64Var(&FlagGenesisTimeValue, "GenesisTime", 0, "override genesis UNIX time instead of using a random UNIX time")
}

// NewConfigFromFlags creates a simulation from the retrieved values of the flags.
func NewConfigFromFlags() Config {
	return Config{
		GenesisFile:        FlagGenesisFileValue,
		ParamsFile:         FlagParamsFileValue,
		ExportParamsPath:   FlagExportParamsPathValue,
		ExportParamsHeight: FlagExportParamsHeightValue,
		ExportStatePath:    FlagExportStatePathValue,
		ExportStatsPath:    FlagExportStatsPathValue,
		Seed:               FlagSeedValue,
		InitialBlockHeight: FlagInitialBlockHeightValue,
		NumBlocks:          FlagNumBlocksValue,
		BlockSize:          FlagBlockSizeValue,
		Lean:               FlagLeanValue,
		Commit:             FlagCommitValue,
		OnOperation:        FlagOnOperationValue,
		AllInvariants:      FlagAllInvariantsValue,
	}
}

// SetupSimulation creates the config, db (levelDB), temporary directory and logger for
// the simulation tests. If `FlagEnabledValue` is false it skips the current test.
// Returns error on an invalid db intantiation or temp dir creation.
func SetupSimulation(dirPrefix, dbName string) (Config, dbm.DB, string, log.Logger, bool, error) {
	if !FlagEnabledValue {
		return Config{}, nil, "", nil, true, nil
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
		return Config{}, nil, "", nil, false, err
	}

	db, err := sdk.NewLevelDB(dbName, dir)
	if err != nil {
		return Config{}, nil, "", nil, false, err
	}

	return config, db, dir, logger, false, nil
}

// PrintStats prints the corresponding statistics from the app DB.
func PrintStats(db dbm.DB) {
	fmt.Println("\nLevelDB Stats")
	fmt.Println(db.Stats()["leveldb.stats"])
	fmt.Println("LevelDB cached block size", db.Stats()["leveldb.cachedblock"])
}

type Config struct {
	GenesisFile string // custom simulation genesis file; cannot be used with params file
	ParamsFile  string // custom simulation params file which overrides any random params; cannot be used with genesis

	ExportParamsPath   string // custom file path to save the exported params JSON
	ExportParamsHeight int    // height to which export the randomly generated params
	ExportStatePath    string // custom file path to save the exported app state JSON
	ExportStatsPath    string // custom file path to save the exported simulation statistics JSON

	Seed               int64  // simulation random seed
	InitialBlockHeight int    // initial block to start the simulation
	NumBlocks          int    // number of new blocks to simulate from the initial block height
	BlockSize          int    // operations per block
	ChainID            string // chain-id used on the simulation

	Lean   bool // lean simulation log output
	Commit bool // have the simulation commit

	OnOperation   bool // run slow invariants every operation
	AllInvariants bool // print all failed invariants if a broken invariant is found
}
