package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
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

	var (
		devProfit       sdk.Coins
		remainingProfit sdk.Coins
		profitSplit     int64
	)

	if daysSinceGenesis < types.Phase1Length {
		profitSplit = types.ProfitSplitPhase1
	} else if daysSinceGenesis < types.Phase2Length {
		profitSplit = types.ProfitSplitPhase2
	} else {
		profitSplit = types.ProfitSplitPhase3
	}

	// Calculate the developer fee from all arb profits
	for _, arbProfit := range arbProfits {
		devProfitAmount := arbProfit.Amount.MulRaw(profitSplit).QuoRaw(100)
		devProfit = append(devProfit, sdk.NewCoin(arbProfit.Denom, devProfitAmount))
	}

	// Send the developer profit to the developer account
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, developerAccount, devProfit); err != nil {
		return err
	}

	// Remove the developer profit from the remaining arb profits
	remainingProfit = arbProfits.Sub(devProfit...)

	// If the remaining arb profits has the OSMO denom for one of the coins, burn the OSMO by sending to the null address
	arbProfitsOsmoCoin := sdk.NewCoin(types.OsmosisDenomination, remainingProfit.AmountOf(types.OsmosisDenomination))
	if arbProfitsOsmoCoin.IsPositive() {
		err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			types.DefaultNullAddress,
			sdk.NewCoins(arbProfitsOsmoCoin),
		)
		if err != nil {
			return err
		}
	}

	// Remove the burned OSMO from the remaining arb profits
	remainingProfit = remainingProfit.Sub(arbProfitsOsmoCoin)

	// Send all remaining arb profits to the community pool
	return k.distributionKeeper.FundCommunityPool(
		ctx,
		remainingProfit,
		k.accountKeeper.GetModuleAddress(types.ModuleName),
	)
}
