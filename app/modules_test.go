package app

import (
	"reflect"
	"testing"

	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func TestOrderEndBlockers_Determinism(t *testing.T) {
	db := dbm.NewMemDB()
	app := NewOsmosisApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, sims.EmptyAppOptions{}, EmptyWasmOpts)

	for i := 0; i < 1000; i++ {
		a := OrderEndBlockers(app.mm.ModuleNames())
		b := OrderEndBlockers(app.mm.ModuleNames())
		require.True(t, reflect.DeepEqual(a, b))
	}
}
