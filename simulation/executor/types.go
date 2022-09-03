package simulation

import (
	"encoding/json"
	"math/rand"
	"time"

	legacysim "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/osmosis-labs/osmosis/v11/simulation/simtypes"
)

// AppStateFn returns the app state json bytes and the genesis accounts
type AppStateFn func(simManager *simtypes.Manager, r *rand.Rand, accs []legacysim.Account, config InitializationConfig) (
	appState json.RawMessage, accounts []legacysim.Account, chainId string, genesisTimestamp time.Time,
)

// RandomAccountFn returns a slice of n random simulation accounts
type RandomAccountFn func(r *rand.Rand, n int) []legacysim.Account
