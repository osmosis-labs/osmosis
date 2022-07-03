package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/simulation"
)

//nolint:structcheck
type SimCtx struct {
	r *rand.Rand
	// TODO: delete this, once we cleanup simulator initialization logic,
	// and can then setup SimCtx with base seed.
	internalSeed int64
	rCounter     int64
	seededMap    map[string]*rand.Rand

	App      *baseapp.BaseApp
	Accounts []simulation.Account
	Cdc      codec.JSONCodec // application codec
	ChainID  string
}

func NewSimCtx(r *rand.Rand, app *baseapp.BaseApp, accounts []simulation.Account, chainID string) *SimCtx {
	return &SimCtx{
		r:            r,
		internalSeed: r.Int63(),
		rCounter:     0,
		seededMap:    map[string]*rand.Rand{},

		App:      app,
		Accounts: accounts,
		ChainID:  chainID,
	}
}

func (sim *SimCtx) GetRand() *rand.Rand {
	sim.rCounter += 1
	r := rand.New(rand.NewSource(sim.internalSeed + sim.rCounter))
	return r
}

// TODO: Refactor to eventually seed a new prng from seed
// and maintain a cache of seed -> rand
func (sim *SimCtx) GetSeededRand(seed string) *rand.Rand {
	return sim.r
}
