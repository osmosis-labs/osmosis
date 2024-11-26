package types_test

import (
	"math/rand"
	"testing"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/mint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	defaultDeveloperVestingProportion = osmomath.NewDecWithPrec(3, 1)
	defaultProvisionsAmount           = osmomath.NewDec(10)
	defaultParams                     = types.Params{
		MintDenom: sdk.DefaultBondDenom,
		DistributionProportions: types.DistributionProportions{
			DeveloperRewards: defaultDeveloperVestingProportion,
		},
	}
)

// Benchmarking :)
// previously using osmomath.Int operations:
// BenchmarkEpochProvision-4 5000000 220 ns/op
//
// using osmomath.Dec operations: (current implementation)
// BenchmarkEpochProvision-4 3000000 429 ns/op
func BenchmarkEpochProvision(b *testing.B) {
	b.ReportAllocs()
	minter := types.InitialMinter()
	params := types.DefaultParams()

	s1 := rand.NewSource(100)
	r1 := rand.New(s1)
	minter.EpochProvisions = osmomath.NewDec(r1.Int63n(1000000))

	// run the EpochProvision function b.N times
	for n := 0; n < b.N; n++ {
		minter.EpochProvision(params)
	}
}

// Next epoch provisions benchmarking
// BenchmarkNextEpochProvisions-4 5000000 251 ns/op
func BenchmarkNextEpochProvisions(b *testing.B) {
	b.ReportAllocs()
	minter := types.InitialMinter()
	params := types.DefaultParams()

	// run the NextEpochProvisions function b.N times
	for n := 0; n < b.N; n++ {
		minter.NextEpochProvisions(params)
	}
}

func TestMinterValidate(t *testing.T) {
	testcases := []struct {
		name     string
		minter   types.Minter
		expected error
	}{
		{
			"valid - success",
			types.InitialMinter(),
			nil,
		},
		{
			"negative -errir",
			types.Minter{
				EpochProvisions: osmomath.NewDec(-1),
			},
			types.ErrNegativeEpochProvisions,
		},
		{
			"nil -error",
			types.Minter{},
			types.ErrNilEpochProvisions,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.minter.Validate()
			if tc.expected != nil {
				require.Error(t, actual)
				require.Equal(t, actual, tc.expected)
			} else {
				require.NoError(t, actual)
			}
		})
	}
}

// TestGetInflationProvisions sanity checks that inflation provisions are calculated correctly.
func TestGetInflationProvisions(t *testing.T) {
	// Setup
	var (
		minter = types.NewMinter(defaultProvisionsAmount)

		expectedDenom           = defaultParams.MintDenom
		expectedInflationAmount = defaultProvisionsAmount.Mul(osmomath.OneDec().Sub(defaultDeveloperVestingProportion))
	)

	// System under test
	actualInflationProvisions := minter.GetInflationProvisions(defaultParams)

	// Assertions
	require.Equal(t, expectedDenom, actualInflationProvisions.Denom)
	require.Equal(t, expectedInflationAmount, actualInflationProvisions.Amount)
}

// TestGetDeveloperVestingProvisions sanity checks that developer vesting provisions are calculated correctly.
func TestGetDeveloperVestingProvisions(t *testing.T) {
	// Setup
	var (
		minter = types.NewMinter(defaultProvisionsAmount)

		expectedDenom           = defaultParams.MintDenom
		expectedInflationAmount = defaultProvisionsAmount.Mul(defaultDeveloperVestingProportion)
	)

	// System under test
	actualInflationProvisions := minter.GetDeveloperVestingEpochProvisions(defaultParams)

	// Assertions
	require.Equal(t, expectedDenom, actualInflationProvisions.Denom)
	require.Equal(t, expectedInflationAmount, actualInflationProvisions.Amount)
}
