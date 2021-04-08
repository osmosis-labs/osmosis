package types

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestEpochProvision(t *testing.T) {
	minter := InitialMinter()
	params := DefaultParams()

	daysPerYear := int64(365)

	tests := []struct {
		annualProvisions int64
		expProvisions    int64
	}{
		{daysPerYear / 7, 1},
		{daysPerYear/7 + 1, 1},
		{(daysPerYear / 7) * 2, 2},
		{(daysPerYear / 7) / 2, 0},
	}
	for i, tc := range tests {
		minter.AnnualProvisions = sdk.NewDec(tc.annualProvisions)
		provisions := minter.EpochProvision(params)

		expProvisions := sdk.NewCoin(params.MintDenom,
			sdk.NewInt(tc.expProvisions))

		require.True(t, expProvisions.IsEqual(provisions),
			"test: %v\n\tExp: %v\n\tGot: %v\n",
			i, tc.expProvisions, provisions)
	}
}

// Benchmarking :)
// previously using sdk.Int operations:
// BenchmarkEpochProvision-4 5000000 220 ns/op
//
// using sdk.Dec operations: (current implementation)
// BenchmarkEpochProvision-4 3000000 429 ns/op
func BenchmarkEpochProvision(b *testing.B) {
	b.ReportAllocs()
	minter := InitialMinter()
	params := DefaultParams()

	s1 := rand.NewSource(100)
	r1 := rand.New(s1)
	minter.AnnualProvisions = sdk.NewDec(r1.Int63n(1000000))

	// run the EpochProvision function b.N times
	for n := 0; n < b.N; n++ {
		minter.EpochProvision(params)
	}
}

// Next annual provisions benchmarking
// BenchmarkNextAnnualProvisions-4 5000000 251 ns/op
func BenchmarkNextAnnualProvisions(b *testing.B) {
	b.ReportAllocs()
	minter := InitialMinter()
	params := DefaultParams()

	// run the NextAnnualProvisions function b.N times
	for n := 0; n < b.N; n++ {
		minter.NextAnnualProvisions(params)
	}
}
