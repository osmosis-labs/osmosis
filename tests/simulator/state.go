package simapp

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app"
	osmosim "github.com/osmosis-labs/osmosis/v27/simulation/executor"
	osmosimtypes "github.com/osmosis-labs/osmosis/v27/simulation/simtypes"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// InitChainFn returns the initial application state using a genesis or the simulation parameters.
// It panics if the user provides files for both of them.
// If a file is not given for the genesis or the sim params, it creates a randomized one.
func InitChainFn() osmosim.InitChainFn {
	cdc := app.MakeEncodingConfig().Marshaler
	return func(simManager osmosimtypes.ModuleGenesisGenerator, r *rand.Rand, accs []simtypes.Account, config osmosim.InitializationConfig,
	) (simAccs []simtypes.Account, req abci.RequestInitChain) {
		// N.B.: wasmd has the following check in its simulator:
		// https://github.com/osmosis-labs/wasmd/blob/c2ec9092d086b5ac6dd367f33ce8b5cce8e4c5f5/x/wasm/types/types.go#L261-L264
		// As a result, it is easy to overflow and become negative if seconds are set too large.
		genesisTime := time.Unix(0, r.Int63())

		appParams := make(simtypes.AppParams)
		if config.ParamsFile != "" {
			bz, err := os.ReadFile(config.ParamsFile)
			if err != nil {
				panic(err)
			}

			err = json.Unmarshal(bz, &appParams)
			if err != nil {
				panic(err)
			}
		}
		appState, simAccs := AppStateRandomizedFn(simManager, r, cdc, accs, genesisTime, appParams)
		appState = updateStakingAndBankState(appState, cdc)

		req = abci.RequestInitChain{
			Time:            genesisTime,
			ChainId:         config.ChainID,
			ConsensusParams: osmosim.DefaultRandomConsensusParams(r, appState, cdc),
			// Validators: ...,
			AppStateBytes: appState,
			// InitialHeight: ...,
		}
		return simAccs, req
	}
}

func updateStakingAndBankState(appState json.RawMessage, cdc codec.JSONCodec) json.RawMessage {
	rawState := make(map[string]json.RawMessage)
	err := json.Unmarshal(appState, &rawState)
	if err != nil {
		panic(err)
	}

	stakingStateBz, ok := rawState[stakingtypes.ModuleName]
	if !ok {
		panic("staking genesis state is missing")
	}

	stakingState := new(stakingtypes.GenesisState)
	err = cdc.UnmarshalJSON(stakingStateBz, stakingState)
	if err != nil {
		panic(err)
	}
	// compute not bonded balance
	notBondedTokens := osmomath.ZeroInt()
	for _, val := range stakingState.Validators {
		if val.Status != stakingtypes.Unbonded {
			continue
		}
		notBondedTokens = notBondedTokens.Add(val.GetTokens())
	}
	notBondedCoins := sdk.NewCoin(stakingState.Params.BondDenom, notBondedTokens)
	// edit bank state to make it have the not bonded pool tokens
	bankStateBz, ok := rawState[banktypes.ModuleName]
	if !ok {
		panic("bank genesis state is missing")
	}
	bankState := new(banktypes.GenesisState)
	err = cdc.UnmarshalJSON(bankStateBz, bankState)
	if err != nil {
		panic(err)
	}

	stakingAddr := authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName).String()
	var found bool
	for _, balance := range bankState.Balances {
		if balance.Address == stakingAddr {
			found = true
			break
		}
	}
	if !found {
		bankState.Balances = append(bankState.Balances, banktypes.Balance{
			Address: stakingAddr,
			Coins:   sdk.NewCoins(notBondedCoins),
		})
	}

	// change appState back
	rawState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(stakingState)
	rawState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankState)

	// replace appstate
	appState, err = json.Marshal(rawState)
	if err != nil {
		panic(err)
	}
	return appState
}

// AppStateRandomizedFn creates calls each module's GenesisState generator function
// and creates the simulation params.
func AppStateRandomizedFn(
	simManager osmosimtypes.ModuleGenesisGenerator, r *rand.Rand, cdc codec.JSONCodec,
	accs []simtypes.Account, genesisTimestamp time.Time, appParams simtypes.AppParams,
) (json.RawMessage, []simtypes.Account) {
	numAccs := int64(len(accs))
	genesisState := app.NewDefaultGenesisState()

	// generate a random amount of initial stake coins and a random initial
	// number of bonded accounts
	initialStake := osmomath.NewInt(r.Int63n(1e12))
	// Don't allow 0 validators to start off with
	numInitiallyBonded := int64(r.Intn(299)) + 1

	if numInitiallyBonded > numAccs {
		numInitiallyBonded = numAccs
	}

	log.Printf(
		`Selected randomly generated parameters for simulated genesis:
{
  stake_per_account: "%d",
  initially_bonded_validators: "%d"
}
`, initialStake, numInitiallyBonded,
	)

	simState := &module.SimulationState{
		AppParams:    appParams,
		Cdc:          cdc,
		Rand:         r,
		GenState:     genesisState,
		Accounts:     accs,
		InitialStake: initialStake,
		NumBonded:    numInitiallyBonded,
		GenTimestamp: genesisTimestamp,
	}

	simManager.GenerateGenesisStates(simState, &osmosimtypes.SimCtx{})

	appState, err := json.Marshal(genesisState)
	if err != nil {
		panic(err)
	}

	return appState, accs
}
