package apptesting

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/osmosis-labs/osmosis/v15/app"
)

func BenchmarkSetup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := new(KeeperTestHelper)
		s.Setup()
	}
}

// Run TestCreateGenesis to get a genesis file for use here.
func BenchmarkInitGenesis(b *testing.B) {
	file, err := os.Open("bench_genesis.json")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	genstate := tmtypes.GenesisDoc{}
	err = decoder.Decode(&genstate)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	abciRequest := abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: simapp.DefaultConsensusParams,
		AppStateBytes:   genstate.AppState,
	}

	for i := 0; i < b.N; i++ {
		db, cleanup := app.GetLevelDbInstance()
		osmoApp := app.NewOsmosisApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, "localtest", 0, NoCrisisAppOpts{}, app.GetWasmEnabledProposals(), app.EmptyWasmOpts)

		osmoApp.InitChain(abciRequest)
		osmoApp.Commit()
		cleanup()
	}
}

type NoCrisisAppOpts struct {
}

func (NoCrisisAppOpts) Get(o string) interface{} {
	if o == crisis.FlagSkipGenesisInvariants {
		return true
	}
	return nil
}
