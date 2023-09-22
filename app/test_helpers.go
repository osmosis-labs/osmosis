package app

import (
	"encoding/json"
	"os"

	cometbftdb "github.com/cometbft/cometbft-db"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"

	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
)

var defaultGenesisBz []byte

func getDefaultGenesisStateBytes() []byte {
	if len(defaultGenesisBz) == 0 {
		genesisState := NewDefaultGenesisState()
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}
		defaultGenesisBz = stateBytes
	}
	return defaultGenesisBz
}

// SetupWithCustomHome initializes a new OsmosisApp with a custom home directory
func SetupWithCustomHome(isCheckTx bool, dir string) *OsmosisApp {
	db := dbm.NewMemDB()
	app := NewOsmosisApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, dir, 0, sims.EmptyAppOptions{}, EmptyWasmOpts)
	if !isCheckTx {
		stateBytes := getDefaultGenesisStateBytes()

		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: sims.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

// Setup initializes a new OsmosisApp.
func Setup(isCheckTx bool) *OsmosisApp {
	return SetupWithCustomHome(isCheckTx, DefaultNodeHome)
}

// SetupTestingAppWithLevelDb initializes a new OsmosisApp intended for testing,
// with LevelDB as a db.
func SetupTestingAppWithLevelDb(isCheckTx bool) (app *OsmosisApp, cleanupFn func()) {
	dir, err := os.MkdirTemp(os.TempDir(), "osmosis_leveldb_testing")
	if err != nil {
		panic(err)
	}
	db, err := cometbftdb.NewGoLevelDB("osmosis_leveldb_testing", dir)
	if err != nil {
		panic(err)
	}
	app = NewOsmosisApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, sims.EmptyAppOptions{}, EmptyWasmOpts)
	if !isCheckTx {
		genesisState := NewDefaultGenesisState()
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: sims.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	cleanupFn = func() {
		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			panic(err)
		}
	}

	return app, cleanupFn
}
