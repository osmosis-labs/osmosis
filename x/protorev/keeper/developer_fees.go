package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"
)

// SendDeveloperFeesToDeveloperAccount sends the developer fees from the module account to the developer account
func (k Keeper) SendDeveloperFeesToDeveloperAccount(ctx sdk.Context) error {
	// Developer account must be set in order to be able to withdraw developer fees
	developerAccount, err := k.GetDeveloperAccount(ctx)
	if err != nil {
		return err
	}

	coins := k.GetAllDeveloperFees(ctx)

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

// UpdateDeveloperFees updates the fees that developers can withdraw from the module account
func (k Keeper) UpdateDeveloperFees(ctx sdk.Context, denom string, profit sdk.Int) error {
	daysSinceGenesis, err := k.GetDaysSinceModuleGenesis(ctx)
	if err != nil {
		return err
	}

	// Calculate the developer fee
	if daysSinceGenesis < 365 {
		// 20% of profit in the first year
		profit = profit.MulRaw(20).QuoRaw(100)
	} else if daysSinceGenesis < 730 {
		// 10% of profit in the second year
		profit = profit.MulRaw(10).QuoRaw(100)
	} else {
		// 5% of profit in subsequent years
		profit = profit.MulRaw(5).QuoRaw(100)
	}

	// Get the developer fees for the denom, if not there then set it to 0 and initialize it
	currentDeveloperFee, err := k.GetDeveloperFees(ctx, denom)
	if err != nil {
		currentDeveloperFee = sdk.NewCoin(denom, sdk.ZeroInt())
	}
	currentDeveloperFee.Amount = currentDeveloperFee.Amount.Add(profit)

	// Set the developer fees for the denom
	if err = k.SetDeveloperFees(ctx, currentDeveloperFee); err != nil {
		return err
	}

	return nil
}
