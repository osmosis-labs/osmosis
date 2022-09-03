package simapp

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	db "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/v11/app"
	"github.com/osmosis-labs/osmosis/v11/simulation/simtypes"
)

func OsmosisAppCreator(logger log.Logger, db db.DB) simtypes.AppCreator {
	return func(homepath string, legacyInvariantPeriod uint, baseappOptions ...func(*baseapp.BaseApp)) simtypes.App {
		return app.NewOsmosisApp(
			logger,
			db,
			nil,
			true, // load latest
			map[int64]bool{},
			homepath,
			legacyInvariantPeriod,
			emptyAppOptions{},
			app.GetWasmEnabledProposals(),
			app.EmptyWasmOpts,
			baseappOptions...)
	}
}

// EmptyAppOptions is a stub implementing AppOptions
type emptyAppOptions struct{}

// Get implements AppOptions
func (ao emptyAppOptions) Get(o string) interface{} {
	return nil
}
