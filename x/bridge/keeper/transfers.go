package keeper

import (
	"errors"
	"fmt"
	"slices"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v23/x/tokenfactory/types"
)

var ErrTransferNotFound = errors.New("transfer info not found")

func (k Keeper) InboundTransfer(
	ctx sdk.Context,
	externalID string,
	sender string,
	destAddr string,
	assetID types.AssetID,
	amount math.Int,
) error {
	params := k.GetParams(ctx)

	// Check if the asset accepts inbound transfers
	assetIdx := params.GetAssetIndex(assetID)
	if assetIdx == notFoundIdx {
		return errorsmod.Wrapf(types.ErrInvalidAssetID, "Asset not found %s", assetID.Name())
	}
	if !params.Assets[assetIdx].Status.InboundActive() {
		return errorsmod.Wrapf(types.ErrInvalidAssetStatus, "Inbound transfers are disabled for this asset")
	}

	// Try to finalize the transfer in store
	finalized, err := k.finalizeInboundTransfer(ctx, externalID, sender, destAddr, assetID, amount, params.VotesNeeded)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrLogic, "Can't finalize inbound trander: %s", err.Error())
	}
	if !finalized {
		// The transfer either doesn't have enough votes or has already been finalised before
		return nil
	}

	// Perform tokenfactory mint
	err = k.mint(ctx, destAddr, assetID, amount)
	if err != nil {
		return errorsmod.Wrap(types.ErrTokenfactory, err.Error())
	}

	// Update the last transfer height value
	params.Assets[assetIdx].LastTransferHeight = ctx.BlockHeight()
	k.SetParam(ctx, types.KeyAssets, params.Assets)

	return nil
}

// finalizeInboundTransfer returns true if the transfer is successfully finalized,
// e.i., the transfer was not finalized before adding the new voter to the voter list,
// but is finalized after the addition.
func (k Keeper) finalizeInboundTransfer(
	ctx sdk.Context,
	externalID string,
	sender string,
	destAddr string,
	assetID types.AssetID,
	amount math.Int,
	votesNeeded uint64,
) (bool, error) {
	// Get the transfer info from the store to update it properly
	transfer, err := k.GetInboundTransfer(ctx, externalID)
	switch {
	case err == nil:
	case errors.Is(err, ErrTransferNotFound):
		// If the transfer is new, then create it
		transfer = types.NewInboundTransfer(externalID, destAddr, assetID, amount)
	default:
		return false, fmt.Errorf("can't get the transfer info from store: %s", err.Error())
	}

	// Check if the sender has already signed this transfer
	if slices.Contains(transfer.Voters, sender) {
		return false, fmt.Errorf("the transfer has already been signed by this sender")
	}

	// This variable is used to detect the right moment to process the transfer.
	// It indicates if the transfer was already finalized before adding a new voter to the voter list.
	alreadyFinalized := transfer.Finalized

	// Add the new voter to the voter list and update the finalization flag
	transfer.Voters = append(transfer.Voters, sender)
	transfer.Finalized = uint64(len(transfer.Voters)) >= votesNeeded

	// Save the updated transfer info
	err = k.UpsertInboundTransfer(ctx, transfer)
	if err != nil {
		return false, fmt.Errorf("can't save the transfer to store: %s", err.Error())
	}

	// If the transfer is already finalized, then we only need to add the sender
	// to the voter list and return
	if alreadyFinalized {
		return false, nil
	}

	// If the transfer is not finalized after adding the new voter,
	// then it still needs more votes
	if !transfer.Finalized {
		return false, nil
	}

	// If the transfer is not finalized before adding the new voter to the voter list,
	// but is finalized after the addition, then it is time to process it

	return true, nil
}

func (k Keeper) mint(ctx sdk.Context, destAddr string, assetID types.AssetID, amount math.Int) error {
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	denom, err := tokenfactorytypes.GetTokenDenom(moduleAddr.String(), assetID.Name())
	if err != nil {
		return fmt.Errorf("can't create a tokenfacroty denom for %s: %w", assetID.Name(), err)
	}

	msgMint := &tokenfactorytypes.MsgMint{
		Sender:        moduleAddr.String(),
		Amount:        sdk.NewCoin(denom, amount),
		MintToAddress: destAddr,
	}

	handler := k.router.Handler(msgMint)
	if handler == nil {
		return fmt.Errorf("can't route a mint message")
	}

	// Ignore resp since it is empty in this method
	_, err = handler(ctx, msgMint)
	if err != nil {
		return fmt.Errorf("can't execute a mint message for %s: %w", assetID.Name(), err)
	}

	return nil
}

// GetInboundTransfer returns the transfer by the external id.
func (k Keeper) GetInboundTransfer(ctx sdk.Context, externalID string) (types.InboundTransfer, error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.InboundTransferKey(externalID))
	if b == nil {
		return types.InboundTransfer{}, ErrTransferNotFound
	}

	var inboundTransfer types.InboundTransfer
	err := k.cdc.Unmarshal(b, &inboundTransfer)
	if err != nil {
		return types.InboundTransfer{}, errors.New("can't unmarshal the inbound transfer")
	}

	return inboundTransfer, nil
}

// UpsertInboundTransfer updates or inserts the inbound transfer depending on
// whether it is already presented in the store or not.
func (k Keeper) UpsertInboundTransfer(ctx sdk.Context, t types.InboundTransfer) error {
	store := ctx.KVStore(k.storeKey)
	key := types.InboundTransferKey(t.ExternalId)

	value, err := k.cdc.Marshal(&t)
	if err != nil {
		return errors.New("can't marshal the inbound transfer")
	}
	store.Set(key, value)

	return nil
}

func (k Keeper) OutboundTransfer(
	ctx sdk.Context,
	sourceAddr string,
	assetID types.AssetID,
	amount math.Int,
) error {
	params := k.GetParams(ctx)

	asset, ok := params.GetAsset(assetID)
	if !ok {
		return errorsmod.Wrapf(types.ErrInvalidAssetID, "Asset not found %s", assetID.Name())
	}

	if !asset.Status.OutboundActive() {
		return errorsmod.Wrapf(types.ErrInvalidAssetStatus, "Outbound transfers are disabled for this asset")
	}

	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	denom, err := tokenfactorytypes.GetTokenDenom(moduleAddr.String(), assetID.Name())
	if err != nil {
		return errorsmod.Wrapf(types.ErrTokenfactory, "Can't create a tokenfacroty denom for %s", assetID.Name())
	}

	msgBurn := &tokenfactorytypes.MsgBurn{
		Sender:          moduleAddr.String(),
		Amount:          sdk.NewCoin(denom, amount),
		BurnFromAddress: sourceAddr,
	}

	handler := k.router.Handler(msgBurn)
	if handler == nil {
		return errorsmod.Wrapf(types.ErrTokenfactory, "Can't route a burn message")
	}

	// ignore resp since it is empty in this method
	_, err = handler(ctx, msgBurn)
	if err != nil {
		return errorsmod.Wrapf(types.ErrTokenfactory, "Can't execute a burn message: %s", err)
	}

	return nil
}
