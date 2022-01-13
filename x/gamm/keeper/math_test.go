package keeper

import (
	"testing"
	"time"

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

func testPool(t *testing.T, tokenBalaceInRaw, tokenWeightInRaw, tokenBalanceOutRaw, tokenWeightOutRaw int64, swapFeeStr string) types.PoolI {
	// TODO: Change test to be table driven
	tokenBalanceIn := sdk.NewInt(tokenBalaceInRaw)
	tokenWeightIn := sdk.NewInt(tokenWeightInRaw)
	tokenBalanceOut := sdk.NewInt(tokenBalanceOutRaw)
	tokenWeightOut := sdk.NewInt(tokenWeightOutRaw)
	swapFee, err := sdk.NewDecFromStr(swapFeeStr)
	require.NoError(t, err)

	pool, err := balancer.NewBalancerPool(
		0,
		balancer.BalancerPoolParams{
			SwapFee: swapFee,
			ExitFee: sdk.ZeroDec(),
		},
		[]types.PoolAsset{
			{Token: sdk.NewCoin(denomIn, tokenBalanceIn), Weight: tokenWeightIn},
			{Token: sdk.NewCoin(denomOut, tokenBalanceOut), Weight: tokenWeightOut},
		},
		"", time.Time{},
	)
	require.NoError(t, err)

	return &pool
}

func TestCalcSpotPrice(t *testing.T) {
	pool := testPool(t, 100, 1, 200, 3, "0")

	actual_spot_price, err := types.CalcSpotPrice(pool, denomIn, denomOut)
	require.NoError(t, err)
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
	pool := testPool(t, 100, 1, 200, 3, "0.01")

	s, err := types.CalcSpotPriceWithSwapFee(pool, denomIn, denomOut)
	require.NoError(t, err)

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
	pool := testPool(t, 100, 1, 200, 3, "0.01")

	s, err := types.CalcOutGivenIn(pool, sdk.Coin{denomIn, sdk.NewInt(40)}, denomOut)
	require.NoError(t, err)

	expectedDec, err := sdk.NewDecFromStr("21.0487006")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(osmomath.PowPrecision().MulInt64(100)),
		"expected value & actual value's difference should less than precision*10: %s, %s",
		expectedDec.String(),
		s.String(),
	)

}

func TestCalcInGivenOut(t *testing.T) {
	pool := testPool(t, 100, 1, 200, 3, "0.01")

	s, err := types.CalcInGivenOut(pool, sdk.Coin{denomOut, sdk.NewInt(70)}, denomIn)
	require.NoError(t, err)

	expectedDec, err := sdk.NewDecFromStr("266.8009177")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(osmomath.PowPrecision().MulInt64(100)),
		"expected value & actual value's difference should less than precision*10: %s, %s",
		expectedDec.String(),
		s.String(),
	)
}

func TestCalcPoolOutGivenSingleIn(t *testing.T) {
	pool := testPool(t, 100, 2, 9999, 8, "0.15")
	pool.AddTotalShares(sdk.NewInt(300))

	s, err := types.CalcPoolOutGivenSingleIn(pool, sdk.Coin{denomIn, sdk.NewInt(40)})
	require.NoError(t, err)

	expectedDec, err := sdk.NewDecFromStr("18.6519592")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(osmomath.PowPrecision().MulInt64(100)),
		"expected value & actual value's difference should less than precision*10: %s, %s",
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
	pool := testPool(t, 9999, 2, 200, 8, "0.15")
	pool.AddTotalShares(sdk.NewInt(300))

	s, err := types.CalcSingleOutGivenPoolIn(pool, sdk.NewInt(40), denomOut)
	require.NoError(t, err)

	expectedDec, err := sdk.NewDecFromStr("31.77534976")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(osmomath.PowPrecision().MulInt64(100)),
		"expected value & actual value's difference should less than precision*10: %s, %s",
		expectedDec.String(),
		s.String(),
	)
}

func TestCalcPoolInGivenSingleOut(t *testing.T) {
	pool := testPool(t, 9999, 2, 200, 8, "0.15")
	pool.AddTotalShares(sdk.NewInt(300))

	s, err := types.CalcPoolInGivenSingleOut(pool, sdk.Coin{denomOut, sdk.NewInt(70)})
	require.NoError(t, err)

	expectedDec, err := sdk.NewDecFromStr("90.29092777")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(osmomath.PowPrecision().MulInt64(100)),
		"expected value & actual value's difference should less than precision*10: %s, %s",
		expectedDec.String(),
		s.String(),
	)
}
