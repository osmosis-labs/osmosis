package app

import (
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func TestOrderEndBlockers_Determinism(t *testing.T) {
	db := dbm.NewMemDB()
	app := NewOsmosisApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, simapp.EmptyAppOptions{}, GetWasmEnabledProposals(), EmptyWasmOpts)

	for i := 0; i < 1000; i++ {
		a := OrderEndBlockers(app.mm.ModuleNames())
		b := OrderEndBlockers(app.mm.ModuleNames())
		require.True(t, reflect.DeepEqual(a, b))
	}
}
