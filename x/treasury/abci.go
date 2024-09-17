package treasury

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/osmosis-labs/osmosis/v26/app/params"
	"github.com/osmosis-labs/osmosis/v26/x/treasury/keeper"
	"github.com/osmosis-labs/osmosis/v26/x/treasury/types"
)

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// Check epoch last block
	if !appparams.IsPeriodLastBlock(ctx, 3*appparams.BlocksPerMinute) {
		return
	}

	refillAmount := k.RefillExchangePool(ctx)
	oldTaxRate := k.GetTaxRate(ctx)
	newTaxRate := k.UpdateReserveFee(ctx)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeTaxRateUpdate,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyExchangePoolRefillAmount, refillAmount.String()),
			sdk.NewAttribute(types.AttributeKeyOldTaxRate, oldTaxRate.String()),
			sdk.NewAttribute(types.AttributeKeyNewTaxRate, newTaxRate.String()),
		),
	)
}
