package simapp

import (
	"github.com/cometbft/cometbft/libs/log"

	db "github.com/cometbft/cometbft-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	simexec "github.com/osmosis-labs/osmosis/v27/simulation/executor"

	"github.com/osmosis-labs/osmosis/v27/app"
	"github.com/osmosis-labs/osmosis/v27/simulation/simtypes"
)

func SymphonyAppCreator(logger log.Logger, db db.DB) simtypes.AppCreator {
	return func(homepath string, legacyInvariantPeriod uint, baseappOptions ...func(*baseapp.BaseApp)) simtypes.App {
		return app.NewSymphonyApp(
			logger,
			db,
			nil,
			true, // load latest
			map[int64]bool{},
			homepath,
			legacyInvariantPeriod,
			emptyAppOptions{},
			app.EmptyWasmOpts,
			baseappOptions...)
	}
}

var SymphonyInitFns = simexec.InitFunctions{
	RandomAccountFn: simexec.WrapRandAccFnForResampling(simulation.RandomAccounts, app.ModuleAccountAddrs()),
	InitChainFn:     InitChainFn(),
}

// EmptyAppOptions is a stub implementing AppOptions
type emptyAppOptions struct{}

// Get implements AppOptions
func (ao emptyAppOptions) Get(o string) interface{} {
	return nil
}
