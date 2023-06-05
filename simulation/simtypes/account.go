package simtypes

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	sdkrand "github.com/osmosis-labs/osmosis/v16/simulation/simtypes/random"
)

func (sim *SimCtx) RandomSimAccount() simulation.Account {
	return sim.randomSimAccount(sim.Accounts)
}

func (sim *SimCtx) randomSimAccount(accs []simulation.Account) simulation.Account {
	r := sim.GetSeededRand("select random account")
	idx := r.Intn(len(accs))
	return accs[idx]
}

type SimAccountConstraint = func(account simulation.Account) bool

// returns acc, accExists := sim.RandomSimAccountWithConstraint(f)
// where acc is a uniformly sampled account from all accounts satisfying the constraint f
// a constraint is satisfied for an account `acc` if f(acc) = true
// accExists is false, if there is no such account.
func (sim *SimCtx) RandomSimAccountWithConstraint(f SimAccountConstraint) (simulation.Account, bool) {
	filteredAddrs := []simulation.Account{}
	for _, acc := range sim.Accounts {
		if f(acc) {
			filteredAddrs = append(filteredAddrs, acc)
		}
	}

	if len(filteredAddrs) == 0 {
		return simulation.Account{}, false
	}
	return sim.randomSimAccount(filteredAddrs), true
}

func (sim *SimCtx) RandomSimAccountWithMinCoins(ctx sdk.Context, coins sdk.Coins) (simulation.Account, error) {
	accHasMinCoins := func(acc simulation.Account) bool {
		spendableCoins := sim.BankKeeper().SpendableCoins(ctx, acc.Address)
		return spendableCoins.IsAllGTE(coins) && coins.DenomsSubsetOf(spendableCoins)
	}
	acc, found := sim.RandomSimAccountWithConstraint(accHasMinCoins)
	if !found {
		return simulation.Account{}, errors.New("no address with min balance found.")
	}
	return acc, nil
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

func (sim *SimCtx) RandomSimAccountWithBalance(ctx sdk.Context) (simulation.Account, error) {
	accHasBal := func(acc simulation.Account) bool {
		return len(sim.BankKeeper().SpendableCoins(ctx, acc.Address)) != 0
	}
	acc, found := sim.RandomSimAccountWithConstraint(accHasBal)
	if !found {
		return simulation.Account{}, errors.New("no address with balance found. Check simulator configuration, this should be very rare.")
	}
	return acc, nil
}

// Returns (account, randSubsetCoins, found), so if found = false, then no such address exists.
// randSubsetCoins is a random subset of the provided denoms, if the account is found.
// TODO: Write unit test
func (sim *SimCtx) SelAddrWithDenoms(ctx sdk.Context, denoms []string) (simulation.Account, sdk.Coins, bool) {
	accHasDenoms := func(acc simulation.Account) bool {
		for _, denom := range denoms {
			if sim.BankKeeper().GetBalance(ctx, acc.Address, denom).Amount.IsZero() {
				return false
			}
			// only return addr if it has spendable coins of requested denom
			coins := sim.BankKeeper().SpendableCoins(ctx, acc.Address)
			for _, coin := range coins {
				if denom == coin.Denom {
					return true
				}
			}
		}
		return true
	}

	acc, accExists := sim.RandomSimAccountWithConstraint(accHasDenoms)
	if !accExists {
		return acc, sdk.Coins{}, false
	}
	balance := sim.RandCoinSubset(ctx, acc.Address, denoms)
	return acc, balance.Sort(), true
}

// SelAddrWithDenom attempts to find an address with the provided denom. This function
// returns (account, randSubsetCoins, found), so if found = false, then no such address exists.
// randSubsetCoins is a random subset of the provided denoms, if the account is found.
// TODO: Write unit test
func (sim *SimCtx) SelAddrWithDenom(ctx sdk.Context, denom string) (simulation.Account, sdk.Coin, bool) {
	acc, subsetCoins, found := sim.SelAddrWithDenoms(ctx, []string{denom})
	if !found {
		return acc, sdk.Coin{}, found
	}
	return acc, subsetCoins[0], found
}

// GetRandSubsetOfKDenoms returns a random subset of coins of k unique denoms from the provided account
// TODO: Write unit test
func (sim *SimCtx) GetRandSubsetOfKDenoms(ctx sdk.Context, acc simulation.Account, k int) (sdk.Coins, bool) {
	// get all spendable coins from provided account
	coins := sim.BankKeeper().SpendableCoins(ctx, acc.Address)
	// ensure account coins are greater than or equal to the requested subset length
	if len(coins) < k {
		return sdk.Coins{}, false
	}
	// randomly remove a denom from the coins array until we reach desired length
	r := sim.GetSeededRand("select random seed")

	for len(coins) != k {
		index := r.Intn(len(coins) - 1)
		coins = RemoveIndex(coins, index)
	}
	// append random amount less than or equal to existing amount to new subset array
	subset := sdk.Coins{}
	for _, c := range coins {
		amt, err := simulation.RandPositiveInt(r, c.Amount)
		if err != nil {
			return sdk.Coins{}, false
		}
		subset = append(subset, sdk.NewCoin(c.Denom, amt))
	}

	// return nothing if the coin struct length is less than requested (sanity check)
	if len(subset) < k {
		return sdk.Coins{}, false
	}

	return subset.Sort(), true
}

// RandomSimAccountWithKDenoms returns an account that possesses k unique denoms
func (sim *SimCtx) RandomSimAccountWithKDenoms(ctx sdk.Context, k int) (simulation.Account, bool) {
	accHasBal := func(acc simulation.Account) bool {
		return len(sim.BankKeeper().SpendableCoins(ctx, acc.Address)) >= k
	}
	return sim.RandomSimAccountWithConstraint(accHasBal)
}

// RandGeometricCoin uniformly samples a denom from the addr's balances.
// Then it samples an Exponentially distributed amount of the addr's coins, with rate = 10.
// (Meaning that on average it samples 10% of the chosen balance)
// Pre-condition: Addr must have a spendable balance
func (sim *SimCtx) RandExponentialCoin(ctx sdk.Context, addr sdk.AccAddress) sdk.Coin {
	balances := sim.BankKeeper().SpendableCoins(ctx, addr)
	if len(balances) == 0 {
		panic("precondition for RandExponentialCoin broken: Addr has 0 spendable balance")
	}
	coin := RandSelect(sim, balances...)
	// TODO: Reconsider if this becomes problematic in the future, but currently thinking it
	// should be fine for simulation.
	r := sim.GetSeededRand("Exponential distribution")
	return sdkrand.RandExponentialCoin(r, coin)
}

func (sim *SimCtx) RandCoinSubset(ctx sdk.Context, addr sdk.AccAddress, denoms []string) sdk.Coins {
	subsetCoins := sdk.Coins{}
	for _, denom := range denoms {
		coins := sim.BankKeeper().SpendableCoins(ctx, addr)
		for _, coin := range coins {
			if denom == coin.Denom {
				amt := sim.RandPositiveInt(coin.Amount)
				subsetCoins = subsetCoins.Add(sdk.NewCoin(coin.Denom, amt))
			}
		}
	}
	return subsetCoins
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

	amt := sim.RandPositiveInt(randCoin.Amount)

	// Create a random fee and verify the fees are within the account's spendable
	// balance.
	fees := sdk.NewCoins(sdk.NewCoin(randCoin.Denom, amt))

	return fees, nil
}

func RemoveIndex(s sdk.Coins, index int) sdk.Coins {
	return append(s[:index], s[index+1:]...)
}
