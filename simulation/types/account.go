package simulation

import (
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
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

// Returns (account, randSubsetCoins, found), so if found = false, then no such address exists.
// randSubsetCoins is a random subset of the provided denoms, if the account is found.
// TODO: Write unit test
func (sim *SimCtx) SelAddrWithDenoms(ctx sdk.Context, denoms []string) (simulation.Account, sdk.Coins, bool) {
	accHasDenoms := func(acc simulation.Account) bool {
		for _, denom := range denoms {
			if sim.App.GetBankKeeper().GetBalance(ctx, acc.Address, denom).Amount.IsZero() {
				return false
			}
		}
		return true
	}

	acc, accExists := sim.RandomSimAccountWithConstraint(accHasDenoms)
	if !accExists {
		return acc, sdk.Coins{}, false
	}
	balance := sim.RandCoinSubset(ctx, acc.Address, denoms)
	return acc, balance, true
}

// RandGeometricCoin uniformly samples a denom from the addr's balances.
// Then it samples an Exponentially distributed amount of the addr's coins, with rate = 10.
// (Meaning that on average it samples 10% of the chosen balance)
func (sim *SimCtx) RandExponentialCoin(ctx sdk.Context, addr sdk.AccAddress) sdk.Coin {
	balances := sim.App.GetBankKeeper().GetAllBalances(ctx, addr)
	coin := RandSelect(sim, balances...)
	// TODO: Reconsider if this becomes problematic in the future, but currently thinking it
	// should be fine for simulation.
	r := sim.GetSeededRand("Exponential distribution")
	lambda := float64(10)
	sample := r.ExpFloat64() / lambda
	// truncate exp at 1, which will only be reached in .0045% of the time.
	// .000045 ~= (1 - CDF(1, Exp[\lambda=10])) = e^{-10}
	sample = math.Min(1, sample)
	// Do some hacky scaling to get this into an SDK decimal,
	// were going to treat it as an integer in the range [0, 10000]
	maxRange := int64(10000)
	intSample := int64(math.Round(sample * float64(maxRange)))
	newAmount := coin.Amount.MulRaw(intSample).QuoRaw(maxRange)
	return sdk.NewCoin(coin.Denom, newAmount)
}

func (sim *SimCtx) RandCoinSubset(ctx sdk.Context, addr sdk.AccAddress, denoms []string) sdk.Coins {
	subsetCoins := sdk.Coins{}
	for _, denom := range denoms {
		bal := sim.App.GetBankKeeper().GetBalance(ctx, addr, denom)
		amt, err := sim.RandPositiveInt(bal.Amount)
		if err != nil {
			panic(err)
		}
		subsetCoins = subsetCoins.Add(sdk.NewCoin(bal.Denom, amt))
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

	amt, err := sim.RandPositiveInt(randCoin.Amount)
	if err != nil {
		return nil, err
	}

	// Create a random fee and verify the fees are within the account's spendable
	// balance.
	fees := sdk.NewCoins(sdk.NewCoin(randCoin.Denom, amt))

	return fees, nil
}
