package keeper

// DONTCOVER

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/gamm/types"
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
			if !baselock.Coins.IsAllGTE(synthlock.Coins) {
				return sdk.FormatInvariant(types.ModuleName, "synthetic-lockup-invariant",
					fmt.Sprintf("\tSynthetic lock token amount %s\n\tUnderlying lock ID: %d, token amount %s\n",
						synthlock.Coins, baselock.ID, baselock.Coins)), true
			}
		}
		return sdk.FormatInvariant(types.ModuleName, "synthetic-lockup-invariant", "All synthetic lockup invariant passed"), false
	}
}
