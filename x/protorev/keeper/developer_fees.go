package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/protorev/types"
)

// SendDeveloperFee sends the developer fee from the module account to the developer account
func (k Keeper) SendDeveloperFee(ctx sdk.Context, arbProfits sdk.Coins) error {
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

	var devProfit sdk.Coins
	var profitSplit int64

	if daysSinceGenesis < types.Phase1Length {
		profitSplit = types.ProfitSplitPhase1
	} else if daysSinceGenesis < types.Phase2Length {
		profitSplit = types.ProfitSplitPhase2
	} else {
		profitSplit = types.ProfitSplitPhase3
	}

	for _, arbProfit := range arbProfits {
		// Calculate the developer fee
		devProfitAmount := arbProfit.Amount.MulRaw(profitSplit).QuoRaw(100)
		devProfit = append(devProfit, sdk.NewCoin(arbProfit.Denom, devProfitAmount))
	}

	// Send the developer profit to the developer account
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, developerAccount, devProfit); err != nil {
		return err
	}

	return nil
}
