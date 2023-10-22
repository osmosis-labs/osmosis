package app

import (
	"encoding/json"
	"os"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymd "github.com/dymensionxyz/dymension/app"
)

var defaultGenesisBz []byte

func getDefaultGenesisStateBytes(cdc codec.JSONCodec) []byte {
	if len(defaultGenesisBz) == 0 {
		genesisState := dymd.NewDefaultGenesisState(cdc)
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}
		defaultGenesisBz = stateBytes
	}
	return defaultGenesisBz
}

// Setup initializes a new OsmosisApp.
func Setup(isCheckTx bool) *dymd.App {
	db := dbm.NewMemDB()
	encCdc := dymd.MakeEncodingConfig()
	app := dymd.New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, dymd.DefaultNodeHome, 0, encCdc, simapp.EmptyAppOptions{})

	if !isCheckTx {
		stateBytes := getDefaultGenesisStateBytes(encCdc.Codec)

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
func SetupTestingAppWithLevelDb(isCheckTx bool) (app *dymd.App, cleanupFn func()) {
	dir, err := os.MkdirTemp(os.TempDir(), "osmosis_leveldb_testing")
	if err != nil {
		panic(err)
	}
	db, err := sdk.NewLevelDB("osmosis_leveldb_testing", dir)
	if err != nil {
		panic(err)
	}
	encCdc := dymd.MakeEncodingConfig()
	app = dymd.New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, dymd.DefaultNodeHome, 5, encCdc, simapp.EmptyAppOptions{})

	if !isCheckTx {
		genesisState := dymd.NewDefaultGenesisState(encCdc.Codec)
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: simapp.DefaultConsensusParams,
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
