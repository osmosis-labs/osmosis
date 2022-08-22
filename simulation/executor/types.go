package simulation

import (
	"encoding/json"
	"math/rand"
	"time"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// AppStateFn returns the app state json bytes and the genesis accounts
type AppStateFn func(r *rand.Rand, accs []simtypes.Account, config Config) (
	appState json.RawMessage, accounts []simtypes.Account, chainId string, genesisTimestamp time.Time,
)

// RandomAccountFn returns a slice of n random simulation accounts
type RandomAccountFn func(r *rand.Rand, n int) []simtypes.Account
