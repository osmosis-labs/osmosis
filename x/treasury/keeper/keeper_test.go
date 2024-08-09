package keeper

import (
	appparams "github.com/osmosis-labs/osmosis/v23/app/params"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/treasury/types"
)

func TestTaxRate(t *testing.T) {
	input := CreateTestInput(t)

	// See that we can get and set tax rate
	for i := int64(0); i < 10; i++ {
		input.TreasuryKeeper.SetTaxRate(input.Ctx, sdk.NewDecWithPrec(i, 2))
		require.Equal(t, sdk.NewDecWithPrec(i, 2), input.TreasuryKeeper.GetTaxRate(input.Ctx))
	}
}

func TestParams(t *testing.T) {
	input := CreateTestInput(t)

	defaultParams := types.DefaultParams()
	input.TreasuryKeeper.SetParams(input.Ctx, defaultParams)

	retrievedParams := input.TreasuryKeeper.GetParams(input.Ctx)
	require.Equal(t, defaultParams, retrievedParams)
}

// TestUpdateReserve tests updating of reserve fee. If the reserve is full, it has to be zero. If the reserve is empty, it has to set non-zero tax rate.
func TestUpdateReserve(t *testing.T) {

	t.Run("reserve is empty", func(t *testing.T) {
		input := CreateTestInput(t)

		// Update the reserve
		newTaxRate := input.TreasuryKeeper.UpdateReserveFee(input.Ctx)
		require.True(t, newTaxRate.GT(sdk.ZeroDec()), "reserve is empty so we should apply the tax rate")
	})
	t.Run("reserve is full", func(t *testing.T) {
		input := CreateTestInput(t)

		exchangeRequirement := input.MarketKeeper.GetExchangeRequirement(input.Ctx)
		require.True(t, exchangeRequirement.GT(sdk.ZeroDec()))

		err := input.BankKeeper.SendCoinsFromModuleToModule(input.Ctx, faucetAccountName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, exchangeRequirement.TruncateInt())))
		require.NoError(t, err)

		// Update the reserve
		newTaxRate := input.TreasuryKeeper.UpdateReserveFee(input.Ctx)
		require.True(t, newTaxRate.IsZero(), "has to be zero since reserve is full")
	})
}
