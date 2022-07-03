package simulation

import (
	"sort"

	"github.com/cosmos/cosmos-sdk/types/module"
	legacysimexec "github.com/cosmos/cosmos-sdk/x/simulation"
)

// AppModuleSimulation defines the standard functions that every module should expose
// for the SDK blockchain simulator
type AppModuleSimulationV2 interface {
	// randomized genesis states
	GenerateGenesisState(*SimCtx)
	// simulation operations (i.e msgs) with their respective weight
	Actions() []Action
}

type v2wrapperOfV1Module struct {
	module module.AppModuleSimulation
	name   string
}

func (mod v2wrapperOfV1Module) Actions() []Action {
	weightedOps := legacysimexec.WeightedOperations(mod.module.WeightedOperations(module.SimulationState{}))
	return ActionsFromWeightedOperations(weightedOps)
}

func (mod v2wrapperOfV1Module) GenerateGenesisState(sim *SimCtx) {
	v1SimState := module.SimulationState{
		Rand:     sim.GetSeededRand("genesis " + mod.name),
		Accounts: sim.Accounts,
	}
	mod.module.GenerateGenesisState(&v1SimState)
}

// SimulationManager defines a simulation manager that provides the high level utility
// for managing and executing simulation functionalities for a group of modules
type Manager struct {
	moduleManager module.Manager
	Modules       []AppModuleSimulationV2 // array of app modules; we use an array for deterministic simulation tests
}

func NewSimulationManager(modules []module.AppModule, overrideModules map[string]module.AppModuleSimulation) Manager {
	simModules := []AppModuleSimulationV2{}
	appModuleNamesSorted := make([]string, 0, len(modules))
	modulesMap := make(map[string]module.AppModule, len(modules))
	for _, module := range modules {
		modulesMap[module.Name()] = module
		appModuleNamesSorted = append(appModuleNamesSorted, module.Name())
	}

	sort.Strings(appModuleNamesSorted)

	for _, moduleName := range appModuleNamesSorted {
		// for every module, see if we override it. If so, use override.
		// Else, if we can cast the app module into a simulation module add it.
		// otherwise no simulation module.
		if simModule, ok := overrideModules[moduleName]; ok {
			simModules = append(simModules, v2wrapperOfV1Module{module: simModule, name: moduleName})
		} else {
			appModule := modulesMap[moduleName]
			if simModule, ok := appModule.(AppModuleSimulationV2); ok {
				simModules = append(simModules, simModule)
			} else if simModule, ok := appModule.(module.AppModuleSimulation); ok {
				simModules = append(simModules, v2wrapperOfV1Module{module: simModule, name: moduleName})
			}
			// cannot cast, so we continue
		}
	}
	return Manager{Modules: simModules}
}
