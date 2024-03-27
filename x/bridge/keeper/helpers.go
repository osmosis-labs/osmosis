package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
<<<<<<< HEAD
	"golang.org/x/exp/slices"
=======
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v24/x/bridge/types"
>>>>>>> 63ccc366 (auto: update Go import paths to v24 on branch main (#7864))
)

// Difference returns the slice of elements that are elements of a but not elements of b.
// TODO: Placed here temporarily. Delete after releasing the new osmoutils version.
func Difference[T comparable](a, b []T) []T {
	mb := make(map[T]struct{}, len(a))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	diff := make([]T, 0)
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// validateSenderIsSigner ensures that the sender is a part of the signers set.
func (k Keeper) validateSenderIsSigner(ctx sdk.Context, sender string) bool {
	return slices.Contains(k.GetParams(ctx).Signers, sender)
}
