package keeper

import (
	"encoding/binary"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"golang.org/x/exp/slices"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

const notFoundIdx = -1

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

// GetInboundTransfer returns the transfer by the external id and height.
func (k Keeper) GetInboundTransfer(ctx sdk.Context, externalID string, externalHeight uint64) (types.InboundTransfer, error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.InboundTransferKey(externalID, externalHeight))
	if b == nil {
		return types.InboundTransfer{}, sdkerrors.ErrNotFound
	}

	var inboundTransfer types.InboundTransfer
	err := k.cdc.Unmarshal(b, &inboundTransfer)
	if err != nil {
		return types.InboundTransfer{}, errors.New("can't unmarshal the inbound transfer")
	}

	return inboundTransfer, nil
}

// UpsertInboundTransfer updates or inserts the value depending on whether it is
// already presented in the store or not.
func (k Keeper) UpsertInboundTransfer(ctx sdk.Context, t types.InboundTransfer) error {
	store := ctx.KVStore(k.storeKey)
	key := types.InboundTransferKey(t.ExternalId, t.ExternalHeight)

	value, err := k.cdc.Marshal(&t)
	if err != nil {
		return errors.New("can't marshal the inbound transfer")
	}
	store.Set(key, value)

	return nil
}

// IsTransferFinalized returns true if the transfer was found in the finalized transfers set.
func (k Keeper) IsTransferFinalized(ctx sdk.Context, externalID string) bool {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.FinalizedTransferKey(externalID))
	return b != nil
}

// SaveFinalizedTransfer creates a new finalized transfer with the given external id.
func (k Keeper) SaveFinalizedTransfer(ctx sdk.Context, externalID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.FinalizedTransferKey(externalID), []byte{})
}

// GetLastTransferHeight returns the last transfer height for the given asset.
func (k Keeper) GetLastTransferHeight(ctx sdk.Context, assetID types.AssetID) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.LastHeightKey(assetID))
	if b == nil {
		return 0, sdkerrors.ErrNotFound
	}

	return binary.BigEndian.Uint64(b), nil
}

// UpsertLastAssetHeight updates or inserts the value depending on whether it is
// already presented in the store or not.
func (k Keeper) UpsertLastAssetHeight(ctx sdk.Context, assetID types.AssetID, height uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(
		types.LastHeightKey(assetID),
		binary.BigEndian.AppendUint64([]byte{}, height),
	)
}
