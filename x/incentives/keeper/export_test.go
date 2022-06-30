package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"
)

// TODO: Why is this called _test?

// Appends the provided gauge ID into an array associated with the provided key.
func (k Keeper) AddGaugeRefByKey(ctx sdk.Context, key []byte, guageID uint64) error {
	return k.addGaugeRefByKey(ctx, key, guageID)
}

// Removes the provided gauge ID from an array associated with the provided key.
func (k Keeper) DeleteGaugeRefByKey(ctx sdk.Context, key []byte, guageID uint64) error {
	return k.deleteGaugeRefByKey(ctx, key, guageID)
}

// Returns the gauge IDs specified by the provided key.
func (k Keeper) GetGaugeRefs(ctx sdk.Context, key []byte) []uint64 {
	return k.getGaugeRefs(ctx, key)
}

// Returns all active gauge-IDs associated with lockups of the provided denom.
func (k Keeper) GetAllGaugeIDsByDenom(ctx sdk.Context, denom string) []uint64 {
	return k.getAllGaugeIDsByDenom(ctx, denom)
}

// Moves a gauge that has reached it's start time from an upcoming to an active status.
func (k Keeper) MoveUpcomingGaugeToActiveGauge(ctx sdk.Context, gauge types.Gauge) error {
	return k.moveUpcomingGaugeToActiveGauge(ctx, gauge)
}

// Moves a gauge that has completed its distribution from an active to a finished status.
func (k Keeper) MoveActiveGaugeToFinishedGauge(ctx sdk.Context, gauge types.Gauge) error {
	return k.moveActiveGaugeToFinishedGauge(ctx, gauge)
}
