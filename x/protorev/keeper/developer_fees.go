package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/protorev/types"
)

// DistributeProfit sends the developer fee from the module account to the developer account
// and burns the remaining profit if denominated in osmo.
func (k Keeper) DistributeProfit(ctx sdk.Context, arbProfits sdk.Coins) error {
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
	var remainingProfit sdk.Coins
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

	// Calculate the remaining profit
	remainingProfit = arbProfits.Sub(devProfit...)

	// Send the developer profit to the developer account
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, developerAccount, devProfit); err != nil {
		return err
	}

	// Burn the remaining osmo profit by sending to the null address iff the profit is denominated in osmo.
	arbProfitsOsmoCoin := sdk.NewCoin(types.OsmosisDenomination, remainingProfit.AmountOf(types.OsmosisDenomination))
	if arbProfitsOsmoCoin.IsPositive() {
		return k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			types.DefaultNullAddress,
			sdk.NewCoins(arbProfitsOsmoCoin),
		)
	}

	remainingProfit = remainingProfit.Sub(arbProfitsOsmoCoin)

	// Send the remaining profit to the community pool if the profit is not denominated in osmo.
	return k.distributionKeeper.FundCommunityPool(
		ctx,
		remainingProfit,
		k.accountKeeper.GetModuleAddress(types.ModuleName),
	)
}
