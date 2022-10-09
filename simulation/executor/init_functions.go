package simulation

// TODO: Move this entire file to simtypes

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
)

// TODO: Consider adding consensus parameters / simulator params / tendermint params to this.
type InitFunctions struct {
	// Why does this take in Numkeys / why isn't this part of the initial state function / config to decide?
	RandomAccountFn RandomAccountFn
	InitChainFn     InitChainFn
}

// TODO: cleanup args in the future, should ideally just be a slice.
func DefaultSimInitFunctions(moduleAccountAddresses map[string]bool) InitFunctions {
	return InitFunctions{
		RandomAccountFn: WrapRandAccFnForResampling(RandomSimAccounts, moduleAccountAddresses),
	}
}

func WrapRandAccFnForResampling(randFn RandomAccountFn, blockList map[string]bool) RandomAccountFn {
	// TODO: do resampling
	return func(r *rand.Rand, n int) []simulation.Account {
		initAccs := randFn(r, n)
		cleanedAccs := make([]simulation.Account, 0, n)
		for _, acc := range initAccs {
			if _, contains := blockList[acc.Address.String()]; !contains {
				cleanedAccs = append(cleanedAccs, acc)
			}
		}
		return cleanedAccs
	}
}

// RandomAccounts generates n random accounts
func RandomSimAccounts(r *rand.Rand, n int) []simulation.Account {
	accs := make([]simulation.Account, n)

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
