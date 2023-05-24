package app

import (
	"encoding/json"
	"os"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var defaultGenesisBz []byte
var DefaultValidatorStr = sdk.MustAccAddressFromBech32("osmovalcons1yhauul02y90hamrq3yuu59mvkcn0a24xdnkc33")

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

// Setup initializes a new OsmosisApp.
func Setup(isCheckTx bool) *OsmosisApp {
	db := dbm.NewMemDB()
	app := NewOsmosisApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, simapp.EmptyAppOptions{}, EmptyWasmOpts)
	if !isCheckTx {
		stateBytes := getDefaultGenesisStateBytes()

		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: simapp.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

// SetupTestingAppWithLevelDb initializes a new OsmosisApp intended for testing,
// with LevelDB as a db.
func SetupTestingAppWithLevelDb(isCheckTx bool) (app *OsmosisApp, reloadApp func() (app *OsmosisApp, cleanupFn func()), cleanupFn func()) {
	dir, err := os.MkdirTemp(os.TempDir(), "osmosis_leveldb_testing")
	if err != nil {
		panic(err)
	}
	reloadApp = func() (app *OsmosisApp, cleanupFn func()) {
		newDb, err := sdk.NewLevelDB("osmosis_leveldb_testing", dir)
		if err != nil {
			panic(err)
		}
		newApp := NewOsmosisApp(log.NewNopLogger(), newDb, nil, true, map[int64]bool{}, DefaultNodeHome, 5, simapp.EmptyAppOptions{}, EmptyWasmOpts)
		cleanupFn = func() {
			newDb.Close()
			err = os.RemoveAll(dir)
			if err != nil {
				panic(err)
			}
		}
		return newApp, cleanupFn
	}
	app, cleanupFn = reloadApp()
	if !isCheckTx {
		genesisState := NewDefaultGenesisState()
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		app.InitChain(
			abci.RequestInitChain{
				Validators: []abci.ValidatorUpdate{
					{Address: DefaultValidatorStr.Bytes(), Power: 100},
				},
				ConsensusParams: simapp.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app, reloadApp, cleanupFn
}
