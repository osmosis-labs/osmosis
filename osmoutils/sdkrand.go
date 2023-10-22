package osmoutils

import (
	"errors"
	"math"
	"math/big"
	"math/rand"
	"time"

	sdk_math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RandPositiveInt get a rand positive sdk.Int
func RandPositiveInt(r *rand.Rand, max sdk_math.Int) (sdk_math.Int, error) {
	if !max.GTE(sdk.OneInt()) {
		return sdk_math.Int{}, errors.New("max too small")
	}

	max = max.Sub(sdk.OneInt())

	return sdk_math.NewIntFromBigInt(new(big.Int).Rand(r, max.BigInt())).Add(sdk_math.OneInt()), nil
}

// RandomAmount generates a random amount
// Note: The range of RandomAmount includes max, and is, in fact, biased to return max as well as 0.
func RandomAmount(r *rand.Rand, max sdk_math.Int) sdk_math.Int {
	randInt := big.NewInt(0)

	switch r.Intn(10) {
	case 0:
		// randInt = big.NewInt(0)
	case 1:
		randInt = max.BigInt()
	default: // NOTE: there are 10 total cases.
		randInt = big.NewInt(0).Rand(r, max.BigInt()) // up to max - 1
	}

	return sdk.NewIntFromBigInt(randInt)
}

// RandomDecAmount generates a random decimal amount
// Note: The range of RandomDecAmount includes max, and is, in fact, biased to return max as well as 0.
func RandomDecAmount(r *rand.Rand, max sdk.Dec) sdk.Dec {
	randInt := big.NewInt(0)

	switch r.Intn(10) {
	case 0:
		// randInt = big.NewInt(0)
	case 1:
		randInt = max.BigInt() // the underlying big int with all precision bits.
	default: // NOTE: there are 10 total cases.
		randInt = big.NewInt(0).Rand(r, max.BigInt())
	}

	return sdk.NewDecFromBigIntWithPrec(randInt, sdk.Precision)
}

// RandTimestamp generates a random timestamp
func RandTimestamp(r *rand.Rand) time.Time {
	// json.Marshal breaks for timestamps greater with year greater than 9999
	unixTime := r.Int63n(253373529600)
	return time.Unix(unixTime, 0)
}

// RandIntBetween returns a random int between two numbers inclusively.
func RandIntBetween(r *rand.Rand, min, max int) int {
	return r.Intn(max-min) + min
}

// returns random subset of the provided coins
// will return at least one coin unless coins argument is empty or malformed
// i.e. 0 amt in coins
func RandSubsetCoins(r *rand.Rand, coins sdk.Coins) sdk.Coins {
	if len(coins) == 0 {
		return sdk.Coins{}
	}
	// make sure at least one coin added
	denomIdx := r.Intn(len(coins))
	coin := coins[denomIdx]
	amt, err := RandPositiveInt(r, coin.Amount)
	// malformed coin. 0 amt in coins
	if err != nil {
		return sdk.Coins{}
	}

	subset := sdk.Coins{sdk.NewCoin(coin.Denom, amt)}

	for i, c := range coins {
		// skip denom that we already chose earlier
		if i == denomIdx {
			continue
		}
		// coin flip if multiple coins
		// if there is single coin then return random amount of it
		if r.Intn(2) == 0 && len(coins) != 1 {
			continue
		}

		amt, err := RandPositiveInt(r, c.Amount)
		// ignore errors and try another denom
		if err != nil {
			continue
		}

		subset = append(subset, sdk.NewCoin(c.Denom, amt))
	}

	return subset.Sort()
}

func RandCoin(r *rand.Rand, coins sdk.Coins) sdk.Coins {
	if len(coins) == 0 {
		return sdk.Coins{}
	}
	// make sure at least one coin added
	denomIdx := r.Intn(len(coins))
	coin := coins[denomIdx]
	amt, err := RandPositiveInt(r, coin.Amount)
	// malformed coin. 0 amt in coins
	if err != nil {
		return sdk.Coins{}
	}

	return sdk.Coins{sdk.NewCoin(coin.Denom, amt)}
}

// RandGeometricCoin uniformly samples a denom from the addr's balances.
// Then it samples an Exponentially distributed amount of the addr's coins, with rate = 10.
// (Meaning that on average it samples 10% of the chosen balance)
// Pre-condition: Addr must have a spendable balance
func RandExponentialCoin(r *rand.Rand, coin sdk.Coin) sdk.Coin {
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
