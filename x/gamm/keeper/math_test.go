package keeper

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func TestCalcSpotPrice(t *testing.T) {
	tc := tc(t, "100", "0.1", "200", "0.3", "", "0", "0", "0")

	actual_spot_price := calcSpotPrice(tc.tokenBalanceIn, tc.tokenWeightIn, tc.tokenBalanceOut, tc.tokenWeightOut)
	// s = (100/.1) / (200 / .3) = (1000) / (2000 / 3) = 1.5
	expected_spot_price, err := sdk.NewDecFromStr("1.5")
	require.NoError(t, err)

	// assert that the spot prices are within the error margin from one another.
	require.True(
		t,
		expected_spot_price.Sub(actual_spot_price).Abs().LTE(powPrecision),
		"expected value & actual value's difference should less than precision",
	)

}

// TODO: Create test vectors with balancer contract
func TestCalcSpotPriceWithSwapFee(t *testing.T) {
	tc := tc(t, "100", "0.1", "200", "0.3", "", "0", "0.01", "0")

	s := calcSpotPriceWithSwapFee(tc.tokenBalanceIn, tc.tokenWeightIn, tc.tokenBalanceOut, tc.tokenWeightOut, tc.swapFee)

	expectedDec, err := sdk.NewDecFromStr("1.51515151")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision),
		"expected value & actual value's difference should less than precision",
	)

}

func TestCalcOutGivenIn(t *testing.T) {
	tc := tc(t, "100", "0.1", "200", "0.3", "", "0", "0.01", "0")

	tokenAmountIn, err := sdk.NewDecFromStr("40")
	require.NoError(t, err)

	s := tc.calcOutGivenIn(tokenAmountIn)

	expectedDec, err := sdk.NewDecFromStr("21.0487006")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision.MulInt64(10000)),
		"expected value & actual value's difference should less than precision*10000",
	)

}

func TestCalcInGivenOut(t *testing.T) {
	tc := tc(t, "100", "0.1", "200", "0.3", "", "0", "0.01", "0")
	tokenAmountOut, err := sdk.NewDecFromStr("70")
	require.NoError(t, err)

	s := tc.calcInGivenOut(tokenAmountOut)

	expectedDec, err := sdk.NewDecFromStr("266.8009177")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision.MulInt64(10)),
		"expected value & actual value's difference should less than precision*10",
	)
}

func TestCalcPoolOutGivenSingleIn(t *testing.T) {
	tc := tc(t, "100", "0.2", "200", "0.8", "1", "300", "0.15", "0")

	tokenAmountIn, err := sdk.NewDecFromStr("40")
	require.NoError(t, err)

	s := tc.calcPoolOutGivenSingleIn(tokenAmountIn)

	expectedDec, err := sdk.NewDecFromStr("18.6519592")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision.MulInt64(10000)),
		"expected value & actual value's difference should less than precision*10000",
	)
}

/*
func TestCalcSingleInGivenPoolOut(t *testing.T) {

	tokenBalanceIn, err := sdk.NewDecFromStr("100")
	require.NoError(t, err)
	tokenWeightIn, err := sdk.NewDecFromStr("0.2")
	require.NoError(t, err)
	poolSupply, err := sdk.NewDecFromStr("300")
	require.NoError(t, err)
	totalWeight, err := sdk.NewDecFromStr("1")
	require.NoError(t, err)
	poolAmountOut, err := sdk.NewDecFromStr("70")
	require.NoError(t, err)
	swapFee, err := sdk.NewDecFromStr("0.15")
	require.NoError(t, err)

	normalizedWeight := tokenWeightIn.Quo(totalWeight)
	s := calcSingleInGivenPoolOut(tokenBalanceIn, normalizedWeight, poolSupply, poolAmountOut, swapFee)

	expectedDec, err := sdk.NewDecFromStr(".")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision.MulInt64(10000)),
		"expected value & actual value's difference should less than precision*10000",
	)
}
*/

func TestCalcSingleOutGivenPoolIn(t *testing.T) {
	tc := tc(t, "100", "0.2", "200", "0.8", "1", "300", "0.15", "0")
	poolAmountIn, err := sdk.NewDecFromStr("40")
	require.NoError(t, err)

	s := tc.calcSingleOutGivenPoolIn(poolAmountIn)

	expectedDec, err := sdk.NewDecFromStr("31.77534976")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision.MulInt64(10000)),
		"expected value & actual value's difference should less than precision*10000",
	)
}

func TestCalcPoolInGivenSingleOut(t *testing.T) {
	tc := tc(t, "100", "0.2", "200", "0.8", "1", "300", "0.15", "0")

	tokenAmountOut, err := sdk.NewDecFromStr("70")
	require.NoError(t, err)

	s := tc.calcPoolInGivenSingleOut(tokenAmountOut)

	expectedDec, err := sdk.NewDecFromStr("90.29092777")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision.MulInt64(10000)),
		"expected value & actual value's difference should less than precision*10000",
	)
}

type testCase struct {
	tokenBalanceIn, tokenWeightIn   sdk.Dec
	tokenBalanceOut, tokenWeightOut sdk.Dec
	totalWeight                     sdk.Dec
	poolSupply                      sdk.Dec
	swapFee, exitFee                sdk.Dec
}

func (tc testCase) reverse() testCase {
	return testCase{
		tc.tokenBalanceOut, tc.tokenWeightOut,
		tc.tokenBalanceIn, tc.tokenWeightIn,
		tc.totalWeight,
		tc.poolSupply,
		tc.swapFee, tc.exitFee,
	}
}

func tc(t *testing.T, tokenBalanceIn, tokenWeightIn, tokenBalanceOut, tokenWeightOut, totalWeight, poolSupply, swapFee, exitFee string) (res testCase) {
	var err error
	res.tokenBalanceIn, err = sdk.NewDecFromStr(tokenBalanceIn)
	require.NoError(t, err)
	res.tokenWeightIn, err = sdk.NewDecFromStr(tokenWeightIn)
	require.NoError(t, err)
	res.tokenBalanceOut, err = sdk.NewDecFromStr(tokenBalanceOut)
	require.NoError(t, err)
	res.tokenWeightOut, err = sdk.NewDecFromStr(tokenWeightOut)
	require.NoError(t, err)
	if totalWeight == "" {
		res.totalWeight = res.tokenWeightIn.Add(res.tokenWeightOut)
	} else {
		res.totalWeight, err = sdk.NewDecFromStr(totalWeight)
	}
	require.NoError(t, err)
	res.poolSupply, err = sdk.NewDecFromStr(poolSupply)
	require.NoError(t, err)
	res.swapFee, err = sdk.NewDecFromStr(swapFee)
	require.NoError(t, err)
	res.exitFee, err = sdk.NewDecFromStr(exitFee)
	require.NoError(t, err)

	return
}

func randtc(t *testing.T, swapFee, exitFee sdk.Dec) (res testCase) {
	res.tokenBalanceIn = sdk.NewInt(rand.Int63()).ToDec()
	res.tokenWeightIn = sdk.NewInt(rand.Int63n(90) + 10).ToDec()
	res.tokenBalanceOut = sdk.NewInt(rand.Int63()).ToDec()
	res.tokenWeightOut = sdk.NewInt(rand.Int63n(90) + 10).ToDec()
	res.totalWeight = res.tokenWeightIn.Add(res.tokenWeightOut)
	res.poolSupply = sdk.NewInt(rand.Int63()).ToDec()
	res.swapFee = swapFee
	res.exitFee = exitFee
	return
}

func (tc testCase) calcInGivenOut(amount sdk.Dec) sdk.Dec {
	return calcInGivenOut(tc.tokenBalanceIn, tc.tokenWeightIn, tc.tokenBalanceOut, tc.tokenWeightOut, amount, tc.swapFee)
}

func (tc testCase) calcOutGivenIn(amount sdk.Dec) sdk.Dec {
	return calcOutGivenIn(tc.tokenBalanceIn, tc.tokenWeightIn, tc.tokenBalanceOut, tc.tokenWeightOut, amount, tc.swapFee)
}

func (tc testCase) calcPoolOutGivenSingleIn(amount sdk.Dec) sdk.Dec {
	return calcPoolOutGivenSingleIn(tc.tokenBalanceIn, tc.tokenWeightIn.Quo(tc.totalWeight), tc.poolSupply, amount, tc.swapFee)
}

func (tc testCase) calcPoolInGivenSingleOut(amount sdk.Dec) sdk.Dec {
	return calcPoolInGivenSingleOut(tc.tokenBalanceOut, tc.tokenWeightOut.Quo(tc.totalWeight), tc.poolSupply, amount, tc.swapFee, tc.exitFee)
}

func (tc testCase) calcSingleInGivenPoolOut(amount sdk.Dec) sdk.Dec {
	return calcSingleInGivenPoolOut(tc.tokenBalanceIn, tc.tokenWeightIn.Quo(tc.totalWeight), tc.poolSupply, amount, tc.swapFee)
}

func (tc testCase) calcSingleOutGivenPoolIn(amount sdk.Dec) sdk.Dec {
	return calcSingleOutGivenPoolIn(tc.tokenBalanceOut, tc.tokenWeightOut.Quo(tc.totalWeight), tc.poolSupply, amount, tc.swapFee, tc.exitFee)
}

func equalWithError(t *testing.T, x, y sdk.Dec, precision int64) {
	require.True(t, x.Quo(y).Sub(sdk.OneDec()).Abs().LTE(sdk.OneDec().Quo(sdk.NewInt(precision).ToDec())),
		"Not equal within error margin with difference %s: %s, %s", x.Quo(y).Sub(sdk.OneDec()), x, y)
}

func TestCalcInverseInvariant(t *testing.T) {
	tcs := make([]testCase, 10000)
	for i := range tcs {
		tcs[i] = randtc(t, sdk.NewInt(rand.Int63n(100)).ToDec().Quo(sdk.NewInt(1000).ToDec()), sdk.NewInt(rand.Int63n(100)).ToDec().Quo(sdk.NewInt(500).ToDec()))
	}

	for _, tc := range tcs {
		for i := 0; i < 10; i++ {
			amount := sdk.NewInt(rand.Int63n(tc.tokenBalanceIn.TruncateInt().Int64() / 20)).ToDec()

			{
				amountOut := tc.calcOutGivenIn(amount)
				amount2 := tc.calcInGivenOut(amountOut)
				equalWithError(t, amount, amount2, 100000)
			}

			{
				shareOut := tc.calcPoolOutGivenSingleIn(amount)
				amount2 := tc.calcSingleInGivenPoolOut(shareOut)
				equalWithError(t, amount, amount2, 100000)
			}

			{
				amountOut := sdk.NewInt(rand.Int63n(tc.tokenBalanceOut.TruncateInt().Int64() / 20)).ToDec()
				shareIn := tc.calcPoolInGivenSingleOut(amountOut)
				amount2 := tc.calcSingleOutGivenPoolIn(shareIn)
				equalWithError(t, amountOut, amount2, 100000)
			}
		}
	}
}
