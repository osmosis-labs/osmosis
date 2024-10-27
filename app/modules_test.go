package app

import (
	"reflect"
	"testing"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"
)

func TestOrderEndBlockers_Determinism(t *testing.T) {
	db := dbm.NewMemDB()
	app := NewOsmosisApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, sims.EmptyAppOptions{}, nil, baseapp.SetChainID("osmosis-1"))

	for i := 0; i < 1000; i++ {
		a := OrderEndBlockers(app.mm.ModuleNames())
		b := OrderEndBlockers(app.mm.ModuleNames())
		require.True(t, reflect.DeepEqual(a, b))
	}
}
