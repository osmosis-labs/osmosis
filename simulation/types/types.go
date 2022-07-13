package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/simulation"
)

//nolint:structcheck
//TODO: Contemplate name better
type SimCtx struct {
	r *rand.Rand
	// TODO: delete this, once we cleanup simulator initialization logic,
	// and can then setup SimCtx with base seed.
	internalSeed int64
	rCounter     int64
	seededMap    map[string]*rand.Rand

	App      App
	Accounts []simulation.Account
	Cdc      codec.JSONCodec // application codec
	ChainID  string

	txbuilder func(ctx sdk.Context, msg sdk.Msg, msgName string) (sdk.Tx, error)
}

func NewSimCtx(r *rand.Rand, app App, accounts []simulation.Account, chainID string) *SimCtx {
	sim := &SimCtx{
		r:            r,
		internalSeed: r.Int63(),
		rCounter:     0,
		seededMap:    map[string]*rand.Rand{},

		App:      app,
		Accounts: accounts,
		ChainID:  chainID,
	}
	sim.txbuilder = sim.defaultTxBuilder
	return sim
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
