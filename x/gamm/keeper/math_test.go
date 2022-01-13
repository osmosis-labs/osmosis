package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/x/gamm/types"

	"github.com/stretchr/testify/require"
)

const (
	denomIn  = "denomin"
	denomOut = "denomout"
)

func poolAsset(denom string, balance int64, weight int64) types.PoolAsset {
	return types.PoolAsset{
		Token:  sdk.NewInt64Coin(denom, balance),
		Weight: sdk.NewInt(weight),
	}
}

func normalizedPoolAsset(denom string, balance int64, normalizedWeightStr string) types.NormalizedPoolAsset {
	normalizedWeight, err := sdk.NewDecFromStr(normalizedWeightStr)
	if err != nil {
		panic(err)
	}
	return types.NormalizedPoolAsset{
		Token:  sdk.NewInt64Coin(denom, balance),
		Weight: normalizedWeight,
	}
}

func swapFee(swapFeeStr string) sdk.Dec {
	swapFee, err := sdk.NewDecFromStr(swapFeeStr)
	if err != nil {
		panic(err)
	}
	return swapFee
}

func TestCalcSpotPrice(t *testing.T) {
	actual_spot_price := types.CalcSpotPrice(
		poolAsset(denomIn, 100, 1),
		poolAsset(denomOut, 200, 3),
	)
	// s = (100/.1) / (200 / .3) = (1000) / (2000 / 3) = 1.5
	expected_spot_price, err := sdk.NewDecFromStr("1.5")
	require.NoError(t, err)

	// assert that the spot prices are within the error margin from one another.
	require.True(
		t,
		expected_spot_price.Sub(actual_spot_price).Abs().LTE(osmomath.PowPrecision()),
		"expected value & actual value's difference should less than precision: %s, %s",
		expected_spot_price.String(),
		actual_spot_price.String(),
	)

}

// TODO: Create test vectors with balancer contract
func TestCalcSpotPriceWithSwapFee(t *testing.T) {
	swapFee, err := sdk.NewDecFromStr("0.01")
	require.NoError(t, err)

	s := types.CalcSpotPriceWithSwapFee(
		poolAsset(denomIn, 100, 1),
		poolAsset(denomOut, 200, 3),
		swapFee,
	)

	expectedDec, err := sdk.NewDecFromStr("1.51515151")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(osmomath.PowPrecision()),
		"expected value & actual value's difference should less than precision: %s, %s",
		expectedDec.String(),
		s.String(),
	)

}

func TestCalcOutGivenIn(t *testing.T) {
	s := types.CalcOutGivenIn(
		balancer.BalancerSwap{},
		normalizedPoolAsset(denomIn, 100, "0.25"),
		normalizedPoolAsset(denomOut, 200, "0.75"),
		sdk.NewInt(40),
		swapFee("0.01"),
	)

	expectedDec, err := sdk.NewDecFromStr("21.0487006")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(osmomath.PowPrecision().MulInt64(10000)),
		"expected value & actual value's difference should less than precision*10000: %s, %s",
		expectedDec.String(),
		s.String(),
	)

}

func TestCalcInGivenOut(t *testing.T) {
	s := types.CalcInGivenOut(
		balancer.BalancerSwap{},
		normalizedPoolAsset(denomIn, 100, "0.25"),
		normalizedPoolAsset(denomOut, 200, "0.75"),
		sdk.NewInt(70),
		swapFee("0.01"),
	)

	expectedDec, err := sdk.NewDecFromStr("266.8009177")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(osmomath.PowPrecision().MulInt64(10)),
		"expected value & actual value's difference should less than precision*10: %s, %s",
		expectedDec.String(),
		s.String(),
	)
}

func TestCalcPoolOutGivenSingleIn(t *testing.T) {
	s := types.CalcPoolOutGivenSingleIn(
		balancer.BalancerSwap{},
		normalizedPoolAsset(denomIn, 100, "0.2"),
		sdk.NewInt(300),
		sdk.NewInt(40),
		swapFee("0.15"),
	)

	expectedDec, err := sdk.NewDecFromStr("18.6519592")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(osmomath.PowPrecision().MulInt64(10000)),
		"expected value & actual value's difference should less than precision*10000: %s, %s",
		expectedDec.String(),
		s.String(),
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
		expectedDec.Sub(s).Abs().LTE(osmomath.PowPrecision().MulInt64(10000)),
		"expected value & actual value's difference should less than precision*10000",
	)
}
*/

func TestCalcSingleOutGivenPoolIn(t *testing.T) {
	s := types.CalcSingleOutGivenPoolIn(
		balancer.BalancerSwap{},
		normalizedPoolAsset(denomOut, 200, "0.8"),
		sdk.NewInt(300),
		sdk.NewInt(40),
		swapFee("0.15"),
		sdk.ZeroDec(),
	)

	expectedDec, err := sdk.NewDecFromStr("31.77534976")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(osmomath.PowPrecision().MulInt64(10000)),
		"expected value & actual value's difference should less than precision*10000: %s, %s",
		expectedDec.String(),
		s.String(),
	)
}

func TestCalcPoolInGivenSingleOut(t *testing.T) {
	s := types.CalcPoolInGivenSingleOut(
		balancer.BalancerSwap{},
		normalizedPoolAsset(denomOut, 200, "0.8"),
		sdk.NewInt(300),
		sdk.NewInt(70),
		swapFee("0.15"),
		sdk.ZeroDec(),
	)

	expectedDec, err := sdk.NewDecFromStr("90.29092777")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(osmomath.PowPrecision().MulInt64(10000)),
		"expected value & actual value's difference should less than precision*10000: %s, %s",
		expectedDec.String(),
		s.String(),
	)
}
