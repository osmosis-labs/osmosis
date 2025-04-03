package keeper

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	markettypes "github.com/osmosis-labs/osmosis/v27/x/market/types"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/treasury/types"
)

func TestTaxRate(t *testing.T) {
	input := CreateTestInput(t)

	// See that we can get and set tax rate
	for i := int64(0); i < 10; i++ {
		input.TreasuryKeeper.SetTaxRate(input.Ctx, osmomath.NewDecWithPrec(i, 2))
		require.Equal(t, osmomath.NewDecWithPrec(i, 2), input.TreasuryKeeper.GetTaxRate(input.Ctx))
	}
}

func TestParams(t *testing.T) {
	input := CreateTestInput(t)

	defaultParams := types.DefaultParams()
	input.TreasuryKeeper.SetParams(input.Ctx, defaultParams)

	retrievedParams := input.TreasuryKeeper.GetParams(input.Ctx)
	require.Equal(t, defaultParams, retrievedParams)
}

// TestKeeper_UpdateReserveFee tests updating of reserve fee. If the reserve is full, it has to be zero. If the reserve is empty, it has to set non-zero tax rate.
func TestKeeper_UpdateReserveFee(t *testing.T) {
	t.Run("reserve is empty", func(t *testing.T) {
		input := CreateTestInput(t)

		// Update the reserve
		newTaxRate := input.TreasuryKeeper.UpdateReserveFee(input.Ctx)
		require.True(t, newTaxRate.GT(osmomath.ZeroDec()), "reserve is empty so we should apply the tax rate")
		require.True(t, newTaxRate.LTE(input.TreasuryKeeper.GetParams(input.Ctx).MaxFeeMultiplier))
	})
	t.Run("reserve is variable", func(t *testing.T) {
		testCases := []struct {
			currentReserveMelody int64
			expectedReserveFee   float64
		}{
			{0, 0.99},
			{37, 0.64},
			{69, 0.32},
			{85, 0.16},
			{93, 0.08},
			{97, 0.02},
		}

		input := CreateTestInput(t)

		exchangeRequirement := input.MarketKeeper.GetExchangeRequirement(input.Ctx)
		t.Logf("exchangeRequirement: %s", exchangeRequirement)

		treasuryModuleAddress := input.AccountKeeper.GetModuleAddress(types.ModuleName)
		for _, testCase := range testCases {
			currentBalance := input.BankKeeper.GetBalance(input.Ctx, treasuryModuleAddress, appparams.BaseCoinUnit)
			reservePercentage := testCase.currentReserveMelody
			mult := osmomath.NewDecWithPrec(reservePercentage, 2)
			expectedReserveValue := exchangeRequirement.Mul(mult).TruncateInt()
			diff := expectedReserveValue.Sub(currentBalance.Amount)
			t.Logf("current balance: %s, expected balance: %s, mult: %s, diff: %s", currentBalance, expectedReserveValue, mult, diff)

			err := input.BankKeeper.SendCoinsFromModuleToModule(input.Ctx, faucetAccountName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, diff)))
			require.NoError(t, err)

			newTaxRate := input.TreasuryKeeper.UpdateReserveFee(input.Ctx)
			t.Logf("newTaxRate: %s", newTaxRate)

			require.InEpsilon(t, testCase.expectedReserveFee, newTaxRate.MustFloat64(), 0.1)
		}
	})
	t.Run("reserve is full", func(t *testing.T) {
		input := CreateTestInput(t)

		exchangeRequirement := input.MarketKeeper.GetExchangeRequirement(input.Ctx)
		require.True(t, exchangeRequirement.GT(osmomath.ZeroDec()))

		err := input.BankKeeper.SendCoinsFromModuleToModule(input.Ctx, faucetAccountName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, exchangeRequirement.TruncateInt())))
		require.NoError(t, err)

		// Update the reserve
		newTaxRate := input.TreasuryKeeper.UpdateReserveFee(input.Ctx)
		require.True(t, newTaxRate.IsZero(), "has to be zero since reserve is full")
	})
}

// TestKeeper_RefillExchangePool tests that reserve pool correctly refills the exchange pool in case of insufficient funds.
func TestKeeper_RefillExchangePool(t *testing.T) {
	t.Run("exchange pool is full", func(t *testing.T) {
		input := CreateTestInput(t)

		exchangeRequirement := input.MarketKeeper.GetExchangeRequirement(input.Ctx)
		err := input.BankKeeper.SendCoinsFromModuleToModule(input.Ctx, faucetAccountName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, exchangeRequirement.TruncateInt())))
		require.NoError(t, err)

		err = input.BankKeeper.SendCoinsFromModuleToModule(input.Ctx, faucetAccountName, markettypes.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, exchangeRequirement.TruncateInt())))
		require.NoError(t, err)

		refillAmount := input.TreasuryKeeper.RefillExchangePool(input.Ctx)
		require.True(t, refillAmount.IsZero())
	})
	t.Run("exchange is pool is under threshold", func(t *testing.T) {
		input := CreateTestInput(t)

		exchangeRequirement := input.MarketKeeper.GetExchangeRequirement(input.Ctx)
		err := input.BankKeeper.SendCoinsFromModuleToModule(input.Ctx, faucetAccountName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, exchangeRequirement.TruncateInt())))
		require.NoError(t, err)

		allowedOffsetPercent := input.TreasuryKeeper.GetParams(input.Ctx).ReserveAllowableOffset
		fillValue := exchangeRequirement.Mul(osmomath.NewDec(100).Sub(allowedOffsetPercent).QuoInt64(100)).TruncateInt()

		err = input.BankKeeper.SendCoinsFromModuleToModule(input.Ctx, faucetAccountName, markettypes.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, fillValue)))
		require.NoError(t, err)

		refillAmount := input.TreasuryKeeper.RefillExchangePool(input.Ctx)
		require.True(t, refillAmount.IsZero())
	})
	t.Run("exchange pool needs a refill", func(t *testing.T) {
		input := CreateTestInput(t)

		exchangeRequirement := input.MarketKeeper.GetExchangeRequirement(input.Ctx)
		err := input.BankKeeper.SendCoinsFromModuleToModule(input.Ctx, faucetAccountName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, exchangeRequirement.TruncateInt())))
		require.NoError(t, err)

		// since exchange pool is empty we will refill for full amount of reserve.
		refillAmount := input.TreasuryKeeper.RefillExchangePool(input.Ctx)
		require.Equal(t, exchangeRequirement, refillAmount, "exchange pool should be refilled for full amount of reserve")
	})

}
