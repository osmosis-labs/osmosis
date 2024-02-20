package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v23/x/protorev/types"
)

// DistributeProfit sends the developer fee from the module account to the developer account
// and burns the remaining profit if denominated in osmo.
func (k Keeper) DistributeProfit(ctx sdk.Context, arbProfit sdk.Coin) error {
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
	devProfit := sdk.NewCoin(arbProfit.Denom, osmomath.ZeroInt())

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

	// Burn the remaining profit by sending to the null address iff the profit is denominated in osmo.
	remainingProfit := sdk.NewCoin(arbProfit.Denom, arbProfit.Amount.Sub(devProfit.Amount))
	if arbProfit.Denom == types.OsmosisDenomination {
		return k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			types.DefaultNullAddress,
			sdk.NewCoins(remainingProfit),
		)
	}

	// Otherwise distribute the remaining profit to the community pool.
	return k.distributionKeeper.FundCommunityPool(
		ctx,
		sdk.NewCoins(remainingProfit),
		k.accountKeeper.GetModuleAddress(types.ModuleName),
	)
}
