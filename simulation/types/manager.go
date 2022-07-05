package simulation

import (
	"sort"

	"github.com/cosmos/cosmos-sdk/types/module"
	legacysimexec "github.com/cosmos/cosmos-sdk/x/simulation"
	"golang.org/x/exp/maps"
)

// AppModuleSimulation defines the standard functions that every module should expose
// for the SDK blockchain simulator
type AppModuleSimulationV2 interface {
	// randomized genesis states
	// TODO: Come back and improve SimulationState interface
	GenerateGenesisState(*module.SimulationState, *SimCtx)
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

func (mod v2wrapperOfV1Module) GenerateGenesisState(simState *module.SimulationState, sim *SimCtx) {
	mod.module.GenerateGenesisState(simState)
}

// SimulationManager defines a simulation manager that provides the high level utility
// for managing and executing simulation functionalities for a group of modules
type Manager struct {
	moduleManager module.Manager
	Modules       map[string]AppModuleSimulationV2 // array of app modules; we use an array for deterministic simulation tests
}

func NewSimulationManager(manager module.Manager, overrideModules map[string]module.AppModuleSimulation) Manager {
	if manager.OrderInitGenesis == nil {
		panic("manager.OrderInitGenesis is unset, needs to be set prior to creating simulation manager")
	}

	simModules := map[string]AppModuleSimulationV2{}
	appModuleNamesSorted := maps.Keys(manager.Modules)
	sort.Strings(appModuleNamesSorted)

	for _, moduleName := range appModuleNamesSorted {
		// for every module, see if we override it. If so, use override.
		// Else, if we can cast the app module into a simulation module add it.
		// otherwise no simulation module.
		if simModule, ok := overrideModules[moduleName]; ok {
			simModules[moduleName] = v2wrapperOfV1Module{module: simModule, name: moduleName}
		} else {
			appModule := manager.Modules[moduleName]
			if simModule, ok := appModule.(AppModuleSimulationV2); ok {
				simModules[moduleName] = simModule
			} else if simModule, ok := appModule.(module.AppModuleSimulation); ok {
				simModules[moduleName] = v2wrapperOfV1Module{module: simModule, name: moduleName}
			}
			// cannot cast, so we continue
		}
	}
	return Manager{moduleManager: manager, Modules: simModules}
}

// TODO: Fix this
// Unfortunately I'm temporarily giving up on fixing genesis logic, its very screwed up in the legacy designs
// and I want to move on to the more interesting goals of this simulation refactor.
// We do need to come back and un-screw up alot of this genesis work.
//
// Thankfully for Osmosis-custom modules, we don't really care about genesis logic. (yet)
// The architectural errors for future readers revolve around on the design of the
// * Design of the AppStateFn (just look at it, osmosis/simapp/state.go)
// 	 * Abstraction leaks overt amounts of code riddle it!
// * Configs being read key by key per module via AppParams, should be a typed config
// * Operation/Action weights being read from params, rather than from come generic config loading
// * every module not just returning a genesis struct, and instead mutating things in place
// The only error corrected in the genesis work over what was present in prior code is:
// better rand handling (simCtx), and calling genesis in the InitGenesis ordering.
func (m Manager) GenerateGenesisStates(simState *module.SimulationState, sim *SimCtx) {
	for _, moduleName := range m.moduleManager.OrderInitGenesis {
		if simModule, ok := m.Modules[moduleName]; ok {
			simModule.GenerateGenesisState(simState, sim)
		}
	}
}
