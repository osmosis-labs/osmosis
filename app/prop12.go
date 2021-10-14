package app

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func prop12(ctx sdk.Context, app *OsmosisApp) {

	payments := GetProp12Payments()

	var total = int64(0)

	for _, payment := range payments {
		addr, err := sdk.AccAddressFromBech32(payment[0])
		if err != nil {
			panic(err)
		}
		amount, err := strconv.ParseInt(strings.TrimSpace(payment[1]), 10, 64)
		if err != nil {
			panic(err)
		}
		coins := sdk.NewCoins(sdk.NewInt64Coin("uosmo", amount))
		if err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, "distribution", addr, coins); err != nil {
			panic(err)
		}
		total += amount
	}

	//deduct from the feePool tracker
	feePool := app.DistrKeeper.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Sub(sdk.NewDecCoins(sdk.NewInt64DecCoin("uosmo", total)))
	app.DistrKeeper.SetFeePool(ctx, feePool)

}
