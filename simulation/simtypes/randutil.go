package simtypes

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/constraints"

	sdkrand "github.com/osmosis-labs/osmosis/v16/simulation/simtypes/random"
)

func RandLTBound[T constraints.Integer](sim *SimCtx, upperbound T) T {
	return RandLTEBound(sim, upperbound-T(1))
}

func RandLTEBound[T constraints.Integer](sim *SimCtx, upperbound T) T {
	max := int(upperbound)
	r := sim.RandIntBetween(0, max+1)
	t := T(r)
	return t
}

func RandSelect[T interface{}](sim *SimCtx, args ...T) T {
	choice := RandLTBound(sim, len(args))
	return args[choice]
}

// RandStringOfLength generates a random string of a particular length
func (sim *SimCtx) RandStringOfLength(n int) string {
	r := sim.GetSeededRand("random string of bounded length")
	return sdkrand.RandStringOfLength(r, n)
}

// RandPositiveInt get a rand positive sdk.Int
func (sim *SimCtx) RandPositiveInt(max sdk.Int) sdk.Int {
	r := sim.GetSeededRand("random bounded positive int")
	v, err := sdkrand.RandPositiveInt(r, max)
	if err != nil {
		panic(err)
	}
	return v
}

// RandomAmount generates a random amount
// that is biased to return max and 0.
func (sim *SimCtx) RandomAmount(max sdk.Int) sdk.Int {
	r := sim.GetSeededRand("random bounded positive int")
	return sdkrand.RandomAmount(r, max)
}

// RandomDecAmount generates a random decimal amount
// Note: The range of RandomDecAmount includes max, and is, in fact, biased to return max as well as 0.
func (sim *SimCtx) RandomDecAmount(max sdk.Dec) sdk.Dec {
	r := sim.GetSeededRand("random bounded positive int")
	return sdkrand.RandomDecAmount(r, max)
}

// RandTimestamp generates a random timestamp
func (sim *SimCtx) RandTimestamp() time.Time {
	r := sim.GetSeededRand("random timestamp")
	return sdkrand.RandTimestamp(r)
}

// RandIntBetween returns a random int between two numbers, inclusive of min, exclusive of max.
func (sim *SimCtx) RandIntBetween(min, max int) int {
	r := sim.GetSeededRand("random int between")
	return sdkrand.RandIntBetween(r, min, max)
}

// returns random subset of the provided coins
// will return at least one coin unless coins argument is empty or malformed
// i.e. 0 amt in coins
func (sim *SimCtx) RandSubsetCoins(coins sdk.Coins) sdk.Coins {
	r := sim.GetSeededRand("random subset coins")
	return sdkrand.RandSubsetCoins(r, coins)
}
