package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func forks(ctx sdk.Context, app *OsmosisApp) {
	switch ctx.BlockHeight() {
	case 650000:
		fix_min_deposit_denom(ctx, app)
	}
}

// Fixes an error where minimum deposit was set to "500 osmo"
// This denom does not exist, which makes it impossible for a proposal to go to a vote
func fix_min_deposit_denom(ctx sdk.Context, app *OsmosisApp) {
	var params = app.GovKeeper.GetDepositParams(ctx)
	params.MinDeposit = sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(500000000)))
	app.GovKeeper.SetDepositParams(ctx, params)
}
