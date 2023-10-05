package simtypes

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// TODO: Contemplate name better
//
//nolint:structcheck
type SimCtx struct {
	rm randManager

	Accounts []simulation.Account
	Cdc      codec.JSONCodec // application codec

	app     App
	chainID string

	txbuilder func(ctx sdk.Context, msg sdk.Msg, msgName string) (sdk.Tx, error)
}

func NewSimCtx(r *rand.Rand, app App, accounts []simulation.Account, chainID string) *SimCtx {
	sim := &SimCtx{
		rm:       newRandManager(r),
		app:      app,
		Accounts: accounts,
		chainID:  chainID,
	}
	sim.txbuilder = sim.defaultTxBuilder
	return sim
}

// TODO: Consider rename to Rand()
func (sim *SimCtx) GetRand() *rand.Rand {
	return sim.rm.GetRand()
}

// TODO: Consider rename to SeededRand()
// or DomainSeparatedRand
func (sim *SimCtx) GetSeededRand(seed string) *rand.Rand {
	return sim.rm.GetSeededRand(seed)
}

// WrapRand returns a new sim object and a cleanup function to write
// Accounts changes to the parent (and invalidate the prior)
func (sim *SimCtx) WrapRand(domainSeparator string) (wrappedSim *SimCtx, cleanup func()) {
	wrappedSim = &SimCtx{
		rm:        sim.rm.WrapRand(domainSeparator),
		app:       sim.app,
		Accounts:  sim.Accounts,
		Cdc:       sim.Cdc,
		chainID:   sim.chainID,
		txbuilder: sim.txbuilder,
	}
	cleanup = func() {
		sim.Accounts = wrappedSim.Accounts
		wrappedSim.Accounts = nil
	}
	return wrappedSim, cleanup
}

func (sim SimCtx) ChainID() string {
	return sim.chainID
}

func (sim SimCtx) BaseApp() *baseapp.BaseApp {
	return sim.app.GetBaseApp()
}

func (sim SimCtx) AppCodec() codec.Codec {
	return sim.app.AppCodec()
}

func (sim SimCtx) AccountKeeper() AccountKeeper {
	return sim.app.GetAccountKeeper()
}

func (sim SimCtx) BankKeeper() BankKeeper {
	return sim.app.GetBankKeeper()
}

func (sim SimCtx) StakingKeeper() stakingkeeper.Keeper {
	return sim.app.GetStakingKeeper()
}

func (sim SimCtx) PoolManagerKeeper() PoolManagerKeeper {
	return sim.app.GetPoolManagerKeeper()
}

// randManager is built to give API's for randomness access
// which allow the caller to avoid "butterfly effects".
// e.g. in the Simulator, I don't want adding one new rand call to a message
// to create an entirely new "run" shape.
type randManager struct {
	// TODO: delete this, once we cleanup simulator initialization logic,
	// and can then setup SimCtx with base seed.
	internalSeed int64
	rCounter     int64
	seededMap    map[string]*rand.Rand

	// if debug = true, we maintain a list of "seen" calls to Wrap,
	// to ensure no duplicates ever get made.
	// TODO: Find a way to expose this to executor.
	// Perhaps we move this to an internal package?
	debug     bool
	seenWraps map[string]bool
}

// TODO: Refactor to take in seed as API's improve
func newRandManager(r *rand.Rand) randManager {
	return randManager{
		internalSeed: r.Int63(),
		rCounter:     0,
		seededMap:    map[string]*rand.Rand{},
		debug:        false,
		seenWraps:    map[string]bool{},
	}
}

func stringToSeed(s string) int64 {
	// take first 8 bytes of the sha256 hash of s.
	// We use this for seeding our rand instances.
	// We use a hash just for convenience, we don't need cryptographic collision resistance,
	// just simple collisions being unlikely.
	bz := sha256.Sum256([]byte(s))
	seedInt := binary.BigEndian.Uint64(bz[:8])
	return int64(seedInt)
}

func (rm *randManager) WrapRand(domainSeparator string) randManager {
	if rm.debug {
		if _, found := rm.seenWraps[domainSeparator]; found {
			panic(fmt.Sprintf("domain separator %s reused!", domainSeparator))
		}
		rm.seenWraps[domainSeparator] = true
	}

	sepInt := stringToSeed(domainSeparator)
	newSeed := rm.internalSeed + sepInt
	r := rand.New(rand.NewSource(newSeed))
	return newRandManager(r)
}

func (rm *randManager) GetRand() *rand.Rand {
	rm.rCounter += 1
	r := rand.New(rand.NewSource(rm.internalSeed + rm.rCounter))
	return r
}

// TODO: Consider rename to DomainSeparatedRand
func (rm *randManager) GetSeededRand(seed string) *rand.Rand {
	// use value in map if present
	if r, ok := rm.seededMap[seed]; ok {
		return r
	}
	seedInt := stringToSeed(seed)
	newSeed := rm.internalSeed + seedInt
	r := rand.New(rand.NewSource(newSeed))
	rm.seededMap[seed] = r
	return r
}
