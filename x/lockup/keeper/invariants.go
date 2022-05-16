package keeper

// DONTCOVER

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v8/x/gamm/types"
)

// RegisterInvariants registers all governance invariants
func RegisterInvariants(ir sdk.InvariantRegistry, keeper Keeper) {
	ir.RegisterRoute(types.ModuleName, "synthetic-lockup-invariant", SyntheticLockupInvariant(keeper))
}

func SyntheticLockupInvariant(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		synthlocks := keeper.GetAllSyntheticLockups(ctx)
		for _, synthlock := range synthlocks {
			baselock, err := keeper.GetLockByID(ctx, synthlock.UnderlyingLockId)
			if err != nil {
				panic(err)
			}
			if baselock.ID != synthlock.UnderlyingLockId {
				return sdk.FormatInvariant(types.ModuleName, "synthetic-lockup-invariant",
					fmt.Sprintf("\tSynthetic lock denom %s\n\tUnderlying lock ID: %d\n\tActual underying lock ID: %d\n",
						synthlock.SynthDenom, synthlock.UnderlyingLockId, baselock.ID,
					)), true
			}
		}

		return sdk.FormatInvariant(types.ModuleName, "synthetic-lockup-invariant", "All synthetic lockup invariant passed"), false
	}
}
