package keeper

import (
	epochstypes "github.com/osmosis-labs/osmosis/v8/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v8/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v8/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	params := k.GetParams(ctx)
	if epochIdentifier == params.DistrEpochIdentifier {
		// begin distribution if it's start time
		gauges := k.GetUpcomingGauges(ctx)
		for _, gauge := range gauges {
			if !ctx.BlockTime().Before(gauge.StartTime) {
				if err := k.moveUpcomingGaugeToActiveGauge(ctx, gauge); err != nil {
					panic(err)
				}
			}
		}

		// distribute due to epoch event
		ctx.EventManager().IncreaseCapacity(2e6)
		gauges = k.GetActiveGauges(ctx)
		// only distribute to active gauges that are for native denoms
		// or non-perpetual and for synthetic denoms.
		// We distribute to perpetual synthetic denoms elsewhere in superfluid.
		// TODO: This method of doing is a bit of hack, should clean this up later.
		distrGauges := []types.Gauge{}
		for _, gauge := range gauges {
			isSynthetic := lockuptypes.IsSyntheticDenom(gauge.DistributeTo.Denom)
			if !(isSynthetic && gauge.IsPerpetual) {
				distrGauges = append(distrGauges, gauge)
			}
		}
		_, err := k.Distribute(ctx, distrGauges)
		if err != nil {
			panic(err)
		}
	}
}

// ___________________________________________________________________________________________________

// Hooks wrapper struct for incentives keeper.
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct.
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks.
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
