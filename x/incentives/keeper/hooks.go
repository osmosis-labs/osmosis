package keeper

import (
	"fmt"

	epochstypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeforeEpochStart is the epoch start hook.
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// AfterEpochEnd is the epoch end hook.
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	params := k.GetParams(ctx)

	if epochIdentifier == params.DistrEpochIdentifier {
		groups, err := k.GetAllGroups(ctx)
		if err != nil {
			return err
		}

		ctx.Logger().Info(fmt.Sprintf("x/incentives AfterEpochEnd, num groups %d, %d", len(groups), ctx.BlockHeight()))
		err = k.AllocateAcrossGauges(ctx, groups)
		if err != nil {
			return err
		}

		// begin distribution if it's start time
		gauges := k.GetUpcomingGauges(ctx)
		ctx.Logger().Info(fmt.Sprintf("x/incentives AfterEpochEnd, num upcoming gauges %d, %d", len(gauges), ctx.BlockHeight()))
		for _, gauge := range gauges {
			if !ctx.BlockTime().Before(gauge.StartTime) {
				if err := k.moveUpcomingGaugeToActiveGauge(ctx, gauge); err != nil {
					return err
				}
			}
		}

		// UNFORKINGTODO OQ: do we upstream this method?
		// if len(gauges) > 10 {
		// 	ctx.EventManager().IncreaseCapacity(2e6)
		// }

		// distribute due to epoch event
		gauges = k.GetActiveGauges(ctx)
		// only distribute to active gauges that are for native denoms
		// or non-perpetual and for synthetic denoms.
		// We distribute to perpetual synthetic denoms elsewhere in superfluid.
		distrGauges := []types.Gauge{}
		for _, gauge := range gauges {
			isSynthetic := lockuptypes.IsSyntheticDenom(gauge.DistributeTo.Denom)
			if !(isSynthetic && gauge.IsPerpetual) {
				distrGauges = append(distrGauges, gauge)
			}
		}

		ctx.Logger().Info("x/incentives AfterEpochEnd: distributing to gauges", "module", types.ModuleName, "numGauges", len(distrGauges), "height", ctx.BlockHeight())
		_, err = k.Distribute(ctx, distrGauges)
		if err != nil {
			return err
		}
		ctx.Logger().Info("x/incentives AfterEpochEnd finished distribution")
	}
	return nil
}

// ___________________________________________________________________________________________________

// Hooks is the wrapper struct for the incentives keeper.
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Hooks returns the hook wrapper struct.
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// GetModuleName implements types.EpochHooks.
func (Hooks) GetModuleName() string {
	return types.ModuleName
}

// BeforeEpochStart is the epoch start hook.
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

// AfterEpochEnd is the epoch end hook.
func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
