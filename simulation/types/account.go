package simulation

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
)

func (sim *SimCtx) RandomSimAccount() simulation.Account {
	r := sim.GetSeededRand("select random account")
	idx := r.Intn(len(sim.Accounts))
	return sim.Accounts[idx]
}

func (sim *SimCtx) RandomExistingAddress() sdk.AccAddress {
	acc := sim.RandomSimAccount()
	return acc.Address
}

func (sim *SimCtx) AddAccount(acc simulation.Account) {
	if _, found := sim.FindAccount(acc.Address); !found {
		sim.Accounts = append(sim.Accounts, acc)
	}
}

// FindAccount iterates over all the simulation accounts to find the one that matches
// the given address
// TODO: Benchmark time in here, we should probably just make a hashmap indexing this.
func (sim *SimCtx) FindAccount(address sdk.Address) (simulation.Account, bool) {
	for _, acc := range sim.Accounts {
		if acc.Address.Equals(address) {
			return acc, true
		}
	}

	return simulation.Account{}, false
}

// TODO: Thread in bank keeper here
// func (sim *SimCtx) SelAddrWithDenoms(denoms []string) (simulation.Account, bool) {
// 	filteredAddrs := []simulation.Account{}
// 	for _, acc := range sim.Accounts {
// 		if acc.Address.Equals(address) {
// 			return acc, true
// 		}
// 	}

// 	return simulation.Account{}, false
// }

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
