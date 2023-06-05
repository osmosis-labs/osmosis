package executortypes

import (
	"encoding/json"
	"math/rand"
	"os"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"golang.org/x/exp/maps"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v16/simulation/simtypes"
)

// Manager defines a simulation manager that provides the high level utility
// for managing and executing simulation functionalities for a group of modules
type Manager struct {
	moduleManager module.Manager
	Modules       map[string]simtypes.AppModuleSimulation // map of all non-legacy app modules;
	legacyModules map[string]module.AppModuleSimulation   // legacy app modules
}

// createSimulationManager returns a simulation manager
// must be ran after modulemanager.SetInitGenesisOrder
func CreateSimulationManager(
	app simtypes.App,
) Manager {
	appCodec := app.AppCodec()
	ak, ok := app.GetAccountKeeper().(*authkeeper.AccountKeeper)
	if !ok {
		panic("account keeper typecast fail")
	}
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(appCodec, *ak, authsims.RandomGenesisAccounts),
	}
	simulationManager := newSimulationManager(app.ModuleManager(), overrideModules)
	return simulationManager
}

func newSimulationManager(manager module.Manager, overrideModules map[string]module.AppModuleSimulation) Manager {
	if manager.OrderInitGenesis == nil {
		panic("manager.OrderInitGenesis is unset, needs to be set prior to creating simulation manager")
	}

	simModules := map[string]simtypes.AppModuleSimulation{}
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
			if simModule, ok := appModule.(simtypes.AppModuleSimulation); ok {
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
	bz, err := os.ReadFile(path)
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

func (m Manager) legacyActions(seed int64, cdc codec.JSONCodec) []simtypes.ActionsWithMetadata {
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
	actions := []simtypes.ActionsWithMetadata{}
	for _, moduleName := range m.moduleManager.OrderInitGenesis {
		// wasmd simulation has txfee assumptions that don't work with Osmosis.
		// TODO: Make an issue / PR on their repo
		if moduleName == "wasm" {
			continue
		}
		if simModule, ok := m.legacyModules[moduleName]; ok {
			weightedOps := simModule.WeightedOperations(simState)
			for _, action := range actionsFromWeightedOperations(moduleName, weightedOps) {
				var actionWithMetaData simtypes.ActionsWithMetadata
				actionWithMetaData.Action = action
				actionWithMetaData.ModuleName = moduleName
				actions = append(actions, actionWithMetaData)
			}
		}
	}
	return actions
}

// TODO: Can we use sim here instead? Perhaps by passing in the simulation module manager to the simulator.
func (m Manager) Actions(seed int64, cdc codec.JSONCodec) []simtypes.ActionsWithMetadata {
	actions := m.legacyActions(seed, cdc)
	moduleKeys := maps.Keys(m.Modules)
	osmoutils.SortSlice(moduleKeys)
	for _, simModuleName := range moduleKeys {
		for _, action := range m.Modules[simModuleName].Actions() {
			var actionWithMetaData simtypes.ActionsWithMetadata
			actionWithMetaData.Action = action
			actionWithMetaData.ModuleName = simModuleName
			actions = append(actions, actionWithMetaData)
		}
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
//   - Abstraction leaks overt amounts of code riddle it!
//
// * Configs being read key by key per module via AppParams, should be a typed config
// * Operation/Action weights being read from params, rather than from come generic config loading
// * every module not just returning a genesis struct, and instead mutating things in place
// The only error corrected in the genesis work over what was present in prior code is:
// better rand handling (simCtx), and calling genesis in the InitGenesis ordering.
func (m Manager) GenerateGenesisStates(simState *module.SimulationState, sim *simtypes.SimCtx) {
	for _, moduleName := range m.moduleManager.OrderInitGenesis {
		if simModule, ok := m.Modules[moduleName]; ok {
			// if we define a random genesis function use it, otherwise use default genesis
			if mod, ok := simModule.(simtypes.AppModuleSimulationGenesis); ok {
				mod.SimulatorGenesisState(simState, sim)
			} else {
				simState.GenState[simModule.Name()] = simModule.DefaultGenesis(simState.Cdc)
			}
		}
		if simModule, ok := m.legacyModules[moduleName]; ok {
			simModule.GenerateGenesisState(simState)
		}
	}
}
