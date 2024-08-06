package treasury

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/osmosis-labs/osmosis/v23/app/params"
	"github.com/osmosis-labs/osmosis/v23/x/treasury/keeper"
	"github.com/osmosis-labs/osmosis/v23/x/treasury/types"
)

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// Check epoch last block
	if !appparams.IsPeriodLastBlock(ctx, appparams.BlocksPerWeek) {
		return
	}

	// Check probation period
	if ctx.BlockHeight() < int64(appparams.BlocksPerWeek*k.WindowProbation(ctx)) {
		return
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypePolicyUpdate,
			sdk.NewAttribute(types.AttributeKeyTaxRate, taxRate.String()),
			sdk.NewAttribute(types.AttributeKeyRewardWeight, rewardWeight.String()),
			sdk.NewAttribute(types.AttributeKeyTaxCap, taxCap.String()),
		),
	)
}
