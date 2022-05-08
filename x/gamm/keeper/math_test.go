package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func TestAbsDifferenceWithSign(t *testing.T) {
	decA, err := sdk.NewDecFromStr("3.2")
	require.NoError(t, err)
	decB, err := sdk.NewDecFromStr("4.3432389")
	require.NoError(t, err)

	s, b := absDifferenceWithSign(decA, decB)
	require.True(t, b)

	expectedDec, err := sdk.NewDecFromStr("1.1432389")
	require.NoError(t, err)
	require.Equal(t, expectedDec, s)
}

func TestPowApprox(t *testing.T) {
	base, err := sdk.NewDecFromStr("0.8")
	require.NoError(t, err)
	exp, err := sdk.NewDecFromStr("0.32")
	require.NoError(t, err)

	s := powApprox(base, exp, powPrecision)
	expectedDec, err := sdk.NewDecFromStr("0.93108385")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision),
		"expected value & actual value's difference should less than precision",
	)
}

func TestPow(t *testing.T) {
	base, err := sdk.NewDecFromStr("1.68")
	require.NoError(t, err)
	exp, err := sdk.NewDecFromStr("0.32")
	require.NoError(t, err)

	s := pow(base, exp)
	expectedDec, err := sdk.NewDecFromStr("1.18058965")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision),
		"expected value & actual value's difference should less than precision",
	)
}

func TestCalcSpotPrice(t *testing.T) {
	// TODO: Change test to be table driven
	tokenBalanceIn, err := sdk.NewDecFromStr("100")
	require.NoError(t, err)
	tokenWeightIn, err := sdk.NewDecFromStr("0.1")
	require.NoError(t, err)
	tokenBalanceOut, err := sdk.NewDecFromStr("200")
	require.NoError(t, err)
	tokenWeightOut, err := sdk.NewDecFromStr("0.3")
	require.NoError(t, err)

	actual_spot_price := calcSpotPrice(tokenBalanceIn, tokenWeightIn, tokenBalanceOut, tokenWeightOut)
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
	tokenBalanceIn, err := sdk.NewDecFromStr("100")
	require.NoError(t, err)
	tokenWeightIn, err := sdk.NewDecFromStr("0.1")
	require.NoError(t, err)
	tokenBalanceOut, err := sdk.NewDecFromStr("200")
	require.NoError(t, err)
	tokenWeightOut, err := sdk.NewDecFromStr("0.3")
	require.NoError(t, err)
	swapFee, err := sdk.NewDecFromStr("0.01")
	require.NoError(t, err)

	s := calcSpotPriceWithSwapFee(tokenBalanceIn, tokenWeightIn, tokenBalanceOut, tokenWeightOut, swapFee)

	expectedDec, err := sdk.NewDecFromStr("1.51515151")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision),
		"expected value & actual value's difference should less than precision",
	)
}

func TestCalcOutGivenIn(t *testing.T) {
	tokenBalanceIn, err := sdk.NewDecFromStr("100")
	require.NoError(t, err)
	tokenWeightIn, err := sdk.NewDecFromStr("0.1")
	require.NoError(t, err)
	tokenBalanceOut, err := sdk.NewDecFromStr("200")
	require.NoError(t, err)
	tokenWeightOut, err := sdk.NewDecFromStr("0.3")
	require.NoError(t, err)
	tokenAmountIn, err := sdk.NewDecFromStr("40")
	require.NoError(t, err)
	swapFee, err := sdk.NewDecFromStr("0.01")
	require.NoError(t, err)

	s := calcOutGivenIn(tokenBalanceIn, tokenWeightIn, tokenBalanceOut, tokenWeightOut, tokenAmountIn, swapFee)

	expectedDec, err := sdk.NewDecFromStr("21.0487006")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision.MulInt64(10000)),
		"expected value & actual value's difference should less than precision*10000",
	)
}

func TestCalcInGivenOut(t *testing.T) {
	tokenBalanceIn, err := sdk.NewDecFromStr("100")
	require.NoError(t, err)
	tokenWeightIn, err := sdk.NewDecFromStr("0.1")
	require.NoError(t, err)
	tokenBalanceOut, err := sdk.NewDecFromStr("200")
	require.NoError(t, err)
	tokenWeightOut, err := sdk.NewDecFromStr("0.3")
	require.NoError(t, err)
	tokenAmountOut, err := sdk.NewDecFromStr("70")
	require.NoError(t, err)
	swapFee, err := sdk.NewDecFromStr("0.01")
	require.NoError(t, err)

	s := calcInGivenOut(tokenBalanceIn, tokenWeightIn, tokenBalanceOut, tokenWeightOut, tokenAmountOut, swapFee)

	expectedDec, err := sdk.NewDecFromStr("266.8009177")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision.MulInt64(10)),
		"expected value & actual value's difference should less than precision*10",
	)
}

func TestCalcPoolOutGivenSingleIn(t *testing.T) {
	tokenBalanceIn, err := sdk.NewDecFromStr("100")
	require.NoError(t, err)
	tokenWeightIn, err := sdk.NewDecFromStr("0.2")
	require.NoError(t, err)
	poolSupply, err := sdk.NewDecFromStr("300")
	require.NoError(t, err)
	totalWeight, err := sdk.NewDecFromStr("1")
	require.NoError(t, err)
	tokenAmountIn, err := sdk.NewDecFromStr("40")
	require.NoError(t, err)
	swapFee, err := sdk.NewDecFromStr("0.15")
	require.NoError(t, err)

	s := calcPoolOutGivenSingleIn(tokenBalanceIn, tokenWeightIn, poolSupply, totalWeight, tokenAmountIn, swapFee)

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

	s := calcSingleInGivenPoolOut(tokenBalanceIn, tokenWeightIn, poolSupply, totalWeight, poolAmountOut, swapFee)

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
	tokenBalanceOut, err := sdk.NewDecFromStr("200")
	require.NoError(t, err)
	tokenWeightOut, err := sdk.NewDecFromStr("0.8")
	require.NoError(t, err)
	poolSupply, err := sdk.NewDecFromStr("300")
	require.NoError(t, err)
	totalWeight, err := sdk.NewDecFromStr("1")
	require.NoError(t, err)
	poolAmountIn, err := sdk.NewDecFromStr("40")
	require.NoError(t, err)
	swapFee, err := sdk.NewDecFromStr("0.15")
	require.NoError(t, err)

	s := calcSingleOutGivenPoolIn(tokenBalanceOut, tokenWeightOut, poolSupply, totalWeight, poolAmountIn, swapFee, sdk.ZeroDec())

	expectedDec, err := sdk.NewDecFromStr("31.77534976")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision.MulInt64(10000)),
		"expected value & actual value's difference should less than precision*10000",
	)
}

func TestCalcPoolInGivenSingleOut(t *testing.T) {
	tokenBalanceOut, err := sdk.NewDecFromStr("200")
	require.NoError(t, err)
	tokenWeightOut, err := sdk.NewDecFromStr("0.8")
	require.NoError(t, err)
	poolSupply, err := sdk.NewDecFromStr("300")
	require.NoError(t, err)
	totalWeight, err := sdk.NewDecFromStr("1")
	require.NoError(t, err)
	tokenAmountOut, err := sdk.NewDecFromStr("70")
	require.NoError(t, err)
	swapFee, err := sdk.NewDecFromStr("0.15")
	require.NoError(t, err)

	s := calcPoolInGivenSingleOut(tokenBalanceOut, tokenWeightOut, poolSupply, totalWeight, tokenAmountOut, swapFee, sdk.ZeroDec())

	expectedDec, err := sdk.NewDecFromStr("90.29092777")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision.MulInt64(10000)),
		"expected value & actual value's difference should less than precision*10000",
	)
}
