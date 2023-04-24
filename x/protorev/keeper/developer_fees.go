package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// Deprecated: Used in v16 upgrade, can be removed in v17
// SendDeveloperFeesToDeveloperAccount sends the developer fees from the module account to the developer account
func (k Keeper) SendDeveloperFeesToDeveloperAccount(ctx sdk.Context) error {
	// Developer account must be set in order to be able to withdraw developer fees
	developerAccount, err := k.GetDeveloperAccount(ctx)
	if err != nil {
		return err
	}

	coins, err := k.GetAllDeveloperFees(ctx)
	if err != nil {
		return err
	}

	for _, coin := range coins {
		// Send the coins to the developer account
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, developerAccount, sdk.NewCoins(coin)); err != nil {
			return err
		}

		// Reset the developer fees for the coin
		k.DeleteDeveloperFees(ctx, coin.Denom)
	}

	return nil
}

// SendDeveloperFee sends the developer fee from the module account to the developer account
func (k Keeper) SendDeveloperFee(ctx sdk.Context, arbProfit sdk.Coin) error {
	// Developer account must be set in order to be able to withdraw developer fees
	developerAccount, err := k.GetDeveloperAccount(ctx)
	if err != nil {
		return err
	}

	// Get the days since genesis
	daysSinceGenesis, err := k.GetDaysSinceModuleGenesis(ctx)
	if err != nil {
		return err
	}

	// Initialize the developer profit to 0
	devProfit := sdk.NewCoin(arbProfit.Denom, sdk.ZeroInt())

	// Calculate the developer fee
	if daysSinceGenesis < types.Phase1Length {
		devProfit.Amount = arbProfit.Amount.MulRaw(types.ProfitSplitPhase1).QuoRaw(100)
	} else if daysSinceGenesis < types.Phase2Length {
		devProfit.Amount = arbProfit.Amount.MulRaw(types.ProfitSplitPhase2).QuoRaw(100)
	} else {
		devProfit.Amount = arbProfit.Amount.MulRaw(types.ProfitSplitPhase3).QuoRaw(100)
	}

	// Send the developer profit to the developer account
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, developerAccount, sdk.NewCoins(devProfit)); err != nil {
		return err
	}

	return nil
}
