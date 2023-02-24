package simapp

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	db "github.com/tendermint/tm-db"

	simexec "github.com/osmosis-labs/osmosis/v15/simulation/executor"

	"github.com/osmosis-labs/osmosis/v15/app"
	"github.com/osmosis-labs/osmosis/v15/simulation/simtypes"
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

var OsmosisInitFns = simexec.InitFunctions{
	RandomAccountFn: simexec.WrapRandAccFnForResampling(simulation.RandomAccounts, app.ModuleAccountAddrs()),
	InitChainFn:     InitChainFn(),
}

// EmptyAppOptions is a stub implementing AppOptions
type emptyAppOptions struct{}

// Get implements AppOptions
func (ao emptyAppOptions) Get(o string) interface{} {
	return nil
}
