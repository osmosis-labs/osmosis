package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func forks(ctx sdk.Context, app *OsmosisApp) {
	switch ctx.BlockHeight() {
	case 712000:
		fix_min_deposit_denom(ctx, app)
		fix_min_commission_rate(ctx, app)
	default:
		// do nothing
		return
	}
}

// Fixes an error where minimum deposit was set to "500 osmo"
// This denom does not exist, which makes it impossible for a proposal to go to a vote
func fix_min_deposit_denom(ctx sdk.Context, app *OsmosisApp) {
	params := app.GovKeeper.GetDepositParams(ctx)
	params.MinDeposit = sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(500000000)))
	app.GovKeeper.SetDepositParams(ctx, params)
}

// Fixes an error where validators can be created with a commission rate
// less than the network minimum rate.
func fix_min_commission_rate(ctx sdk.Context, app *OsmosisApp) {
	// Upgrade every validators min-commission rate
	validators := app.StakingKeeper.GetAllValidators(ctx)
	minCommissionRate := app.StakingKeeper.GetParams(ctx).MinCommissionRate
	for _, v := range validators {
		if v.Commission.Rate.LT(minCommissionRate) {
			comm, err := app.StakingKeeper.MustUpdateValidatorCommission(
				ctx, v, minCommissionRate)
			if err != nil {
				panic(err)
			}
			v.Commission = comm

			// call the before-modification hook since we're about to update the commission
			app.StakingKeeper.BeforeValidatorModified(ctx, v.GetOperator())

			app.StakingKeeper.SetValidator(ctx, v)
		}
	}
}
