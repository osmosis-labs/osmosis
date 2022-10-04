package simtypes

import (
	"github.com/cosmos/cosmos-sdk/types/module"
)

// AppModuleSimulation defines the standard functions that every module should expose
// for the SDK blockchain simulator
type AppModuleSimulation interface {
	module.AppModule

	Actions() []Action
}

type AppModuleSimulationGenesis interface {
	AppModuleSimulation
	// TODO: Come back and improve SimulationState interface
	SimulatorGenesisState(*module.SimulationState, *SimCtx)
}

type AppModuleSimulationPropertyCheck interface {
	module.AppModule

	PropertyChecks() []PropertyCheck
}

type ModuleGenesisGenerator interface {
	GenerateGenesisStates(simState *module.SimulationState, sim *SimCtx)
}
