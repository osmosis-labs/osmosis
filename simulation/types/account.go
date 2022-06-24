package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Account contains a privkey, pubkey, address tuple
// eventually more useful data can be placed in here.
// (e.g. number of coins)
type Account struct {
	PrivKey cryptotypes.PrivKey
	PubKey  cryptotypes.PubKey
	Address sdk.AccAddress
	ConsKey cryptotypes.PrivKey
}

// Equals returns true if two accounts are equal
func (acc Account) Equals(acc2 Account) bool {
	return acc.Address.Equals(acc2.Address)
}

func (sim *SimCtx) RandomSimAccount() Account {
	r := sim.GetSeededRand("select random account")
	idx := r.Intn(len(sim.Accounts))
	return sim.Accounts[idx]
}

func (sim *SimCtx) RandomExistingAddress() sdk.AccAddress {
	acc := sim.RandomSimAccount()
	return acc.Address
}

func (sim *SimCtx) AddAccount(acc Account) {
	if _, found := sim.FindAccount(acc.Address); !found {
		sim.Accounts = append(sim.Accounts, acc)
	}
}

// RandomAccounts generates n random accounts
func RandomAccounts(r *rand.Rand, n int) []Account {
	accs := make([]Account, n)

	for i := 0; i < n; i++ {
		// don't need that much entropy for simulation
		privkeySeed := make([]byte, 15)
		r.Read(privkeySeed)

		accs[i].PrivKey = secp256k1.GenPrivKeyFromSecret(privkeySeed)
		accs[i].PubKey = accs[i].PrivKey.PubKey()
		accs[i].Address = sdk.AccAddress(accs[i].PubKey.Address())

		accs[i].ConsKey = ed25519.GenPrivKeyFromSecret(privkeySeed)
	}

	return accs
}

// FindAccount iterates over all the simulation accounts to find the one that matches
// the given address
// TODO: Benchmark time in here, we should probably just make a hashmap indexing this.
func (sim *SimCtx) FindAccount(address sdk.Address) (Account, bool) {
	for _, acc := range sim.Accounts {
		if acc.Address.Equals(address) {
			return acc, true
		}
	}

	return Account{}, false
}

// RandomFees returns a random fee by selecting a random coin denomination and
// amount from the account's available balance. If the user doesn't have enough
// funds for paying fees, it returns empty coins.
func (sim *SimCtx) RandomFees(ctx sdk.Context, spendableCoins sdk.Coins) (sdk.Coins, error) {
	if spendableCoins.Empty() {
		return nil, nil
	}

	// TODO: Revisit this
	r := sim.GetRand()
	perm := r.Perm(len(spendableCoins))
	var randCoin sdk.Coin
	for _, index := range perm {
		randCoin = spendableCoins[index]
		if !randCoin.Amount.IsZero() {
			break
		}
	}

	if randCoin.Amount.IsZero() {
		return nil, fmt.Errorf("no coins found for random fees")
	}

	amt, err := sim.RandPositiveInt(randCoin.Amount)
	if err != nil {
		return nil, err
	}

	// Create a random fee and verify the fees are within the account's spendable
	// balance.
	fees := sdk.NewCoins(sdk.NewCoin(randCoin.Denom, amt))

	return fees, nil
}
