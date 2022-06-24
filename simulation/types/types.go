package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

//nolint:structcheck
type SimCtx struct {
	r *rand.Rand
	//nolint:unused
	rCounter uint64
	App      *baseapp.BaseApp
	Accounts []Account
	ChainID  string
}

func NewSimCtx(r *rand.Rand, app *baseapp.BaseApp, accounts []Account, chainID string) SimCtx {
	return SimCtx{
		r:        r,
		rCounter: 0,
		App:      app,
		Accounts: accounts,
		ChainID:  chainID,
	}
}

// TODO: Refactor to eventually seed a new prng from rCounter
func (sim *SimCtx) GetRand() *rand.Rand {
	return sim.r
}

// TODO: Refactor to eventually seed a new prng from seed
// and maintain a cache of seed -> rand
func (sim *SimCtx) GetSeededRand(seed string) *rand.Rand {
	return sim.r
}
