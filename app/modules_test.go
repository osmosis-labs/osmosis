package app

import (
	"reflect"
	"testing"

	"cosmossdk.io/simapp"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

func TestOrderEndBlockers_Determinism(t *testing.T) {
	db := dbm.NewMemDB()
	app := NewOsmosisApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, simapp.EmptyAppOptions{}, EmptyWasmOpts)

	for i := 0; i < 1000; i++ {
		a := OrderEndBlockers(app.mm.ModuleNames())
		b := OrderEndBlockers(app.mm.ModuleNames())
		require.True(t, reflect.DeepEqual(a, b))
	}
}
