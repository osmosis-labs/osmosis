package simulation

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"golang.org/x/exp/maps"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
)

// AppModuleSimulation defines the standard functions that every module should expose
// for the SDK blockchain simulator
type AppModuleSimulationV2 interface {
	module.AppModule

	Actions() []Action
	// PropertyTests()
}

type AppModuleSimulationV2WithRandGenesis interface {
	AppModuleSimulationV2
	// TODO: Come back and improve SimulationState interface
	GenerateGenesisState(*module.SimulationState, *SimCtx)
}

// SimulationManager defines a simulation manager that provides the high level utility
// for managing and executing simulation functionalities for a group of modules
type Manager struct {
	moduleManager module.Manager
	Modules       map[string]AppModuleSimulationV2      // map of all non-legacy app modules;
	legacyModules map[string]module.AppModuleSimulation // legacy app modules
}

func NewSimulationManager(manager module.Manager, overrideModules map[string]module.AppModuleSimulation) Manager {
	if manager.OrderInitGenesis == nil {
		panic("manager.OrderInitGenesis is unset, needs to be set prior to creating simulation manager")
	}

	simModules := map[string]AppModuleSimulationV2{}
	legacySimModules := map[string]module.AppModuleSimulation{}
	appModuleNamesSorted := maps.Keys(manager.Modules)
	sort.Strings(appModuleNamesSorted)

	for _, moduleName := range appModuleNamesSorted {
		// for every module, see if we override it. If so, use override.
		// Else, if we can cast the app module into a simulation module add it.
		// otherwise no simulation module.
		if simModule, ok := overrideModules[moduleName]; ok {
			legacySimModules[moduleName] = simModule
		} else {
			appModule := manager.Modules[moduleName]
			if simModule, ok := appModule.(AppModuleSimulationV2); ok {
				simModules[moduleName] = simModule
			} else if simModule, ok := appModule.(module.AppModuleSimulation); ok {
				legacySimModules[moduleName] = simModule
			}
			// cannot cast, so we continue
		}
	}
	return Manager{moduleManager: manager, legacyModules: legacySimModules, Modules: simModules}
}

func loadAppParamsForWasm(path string) simulation.AppParams {
	bz, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	appParams := make(simulation.AppParams)
	err = json.Unmarshal(bz, &appParams)
	if err != nil {
		panic(err)
	}
	return appParams
}

func (m Manager) legacyActions(seed int64, cdc codec.JSONCodec) []Action {
	// We do not support the legacy simulator config format, and just (unfortunately)
	// hardcode this one filepath for wasm.
	// TODO: Clean this up / make a better plan

	simState := module.SimulationState{
		AppParams:    loadAppParamsForWasm("params.json"),
		ParamChanges: []simulation.ParamChange{},
		Contents:     []simulation.WeightedProposalContent{},
		Cdc:          cdc,
	}

	r := rand.New(rand.NewSource(seed))
	// first pass generate randomized params + proposal contents
	for _, moduleName := range m.moduleManager.OrderInitGenesis {
		if simModule, ok := m.legacyModules[moduleName]; ok {
			simState.ParamChanges = append(simState.ParamChanges, simModule.RandomizedParams(r)...)
			simState.Contents = append(simState.Contents, simModule.ProposalContents(simState)...)
		}
	}
	// second pass generate actions
	actions := []Action{}
	for _, moduleName := range m.moduleManager.OrderInitGenesis {
		if simModule, ok := m.legacyModules[moduleName]; ok {
			weightedOps := simModule.WeightedOperations(simState)
			actions = append(actions, actionsFromWeightedOperations(moduleName, weightedOps)...)
		}
	}
	return actions
}

// TODO: Can we use sim here instead? Perhaps by passing in the simulation module manager to the simulator.
func (m Manager) Actions(seed int64, cdc codec.JSONCodec) []Action {
	actions := m.legacyActions(seed, cdc)
	moduleKeys := maps.Keys(m.Modules)
	osmoutils.SortSlice(moduleKeys)
	for _, simModuleName := range moduleKeys {
		actions = append(actions, m.Modules[simModuleName].Actions()...)
	}
	return actions
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
			// if we define a random genesis function use it, otherwise use default genesis
			if mod, ok := simModule.(AppModuleSimulationV2WithRandGenesis); ok {
				mod.GenerateGenesisState(simState, sim)
			} else {
				simState.GenState[simModule.Name()] = simModule.DefaultGenesis(simState.Cdc)
			}
		}
		if simModule, ok := m.legacyModules[moduleName]; ok {
			simModule.GenerateGenesisState(simState)
		}
	}
}
