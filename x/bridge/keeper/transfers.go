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

func (k Keeper) InboundTransfer(
	ctx sdk.Context,
	externalID string,
	externalHeight uint64,
	sender string,
	destAddr string,
	assetID types.AssetID,
	amount math.Int,
) error {
	params := k.GetParams(ctx)

	// Check if the asset accepts inbound transfers
	asset, found := params.GetAsset(assetID)
	if !found {
		return errorsmod.Wrapf(types.ErrInvalidAssetID, "Asset not found %s", assetID.Name())
	}
	if !asset.Status.InboundActive() {
		return errorsmod.Wrapf(types.ErrInvalidAssetStatus, "Inbound transfers are disabled for this asset")
	}

	// Try to finalize the transfer
	finalized, err := k.finalizeInboundTransfer(
		ctx,
		externalID,
		externalHeight,
		sender,
		destAddr,
		assetID,
		amount,
		params.VotesNeeded,
	)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrLogic, "Can't finalize inbound trander: %s", err.Error())
	}
	if !finalized {
		// The transfer either doesn't have enough votes or has already been finalised
		return nil
	}

	// Perform tokenfactory mint
	err = k.mint(ctx, destAddr, assetID, amount)
	if err != nil {
		return errorsmod.Wrap(types.ErrTokenfactory, err.Error())
	}

	// Now when the transfer is finalized, update the state
	k.SaveFinalizedTransfer(ctx, externalID)
	k.UpdateLastAssetHeight(ctx, assetID, externalHeight)

	return nil
}

// finalizeInboundTransfer returns true if the transfer is successfully finalized,
// i.e., the transfer was not finalized before adding the new voter to the voter list,
// but is finalized after the addition.
func (k Keeper) finalizeInboundTransfer(
	ctx sdk.Context,
	externalID string,
	externalHeight uint64,
	sender string,
	destAddr string,
	assetID types.AssetID,
	amount math.Int,
	votesNeeded uint64,
) (bool, error) {
	if k.IsTransferFinalized(ctx, externalID) {
		// Can't finalize the transfer since it's already finalized
		return false, nil
	}

	// Get the transfer info from the store to update it properly
	transfer, err := k.GetInboundTransfer(ctx, externalID, externalHeight)
	switch {
	case err == nil:
	case errors.Is(err, sdkerrors.ErrNotFound):
		// If the transfer is new, then create it
		transfer = types.NewInboundTransfer(externalID, externalHeight, destAddr, assetID, amount)
	default:
		return false, fmt.Errorf("can't get the transfer info from store: %s", err.Error())
	}

	// Check if the sender has already signed this transfer
	if slices.Contains(transfer.Voters, sender) {
		return false, fmt.Errorf("the transfer has already been signed by this sender")
	}

	// Add the new voter to the voter list and update the finalization flag
	transfer.Voters = append(transfer.Voters, sender)
	transfer.Finalized = uint64(len(transfer.Voters)) >= votesNeeded

	err = k.UpsertInboundTransfer(ctx, transfer)
	if err != nil {
		return false, fmt.Errorf("can't save the transfer to store: %s", err.Error())
	}

	// If the transfer is not finalized after adding the new voter,
	// then it still needs more votes
	if !transfer.Finalized {
		return false, nil
	}

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

	// Perform tokenfactory burn
	err := k.burn(ctx, sourceAddr, assetID, amount)
	if err != nil {
		return errorsmod.Wrap(types.ErrTokenfactory, err.Error())
	}

	return nil
}

func (k Keeper) burn(ctx sdk.Context, sourceAddr string, assetID types.AssetID, amount math.Int) error {
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	denom, err := tokenfactorytypes.GetTokenDenom(moduleAddr.String(), assetID.Name())
	if err != nil {
		return fmt.Errorf("can't create a tokenfacroty denom for %s", assetID.Name())
	}

	msgBurn := &tokenfactorytypes.MsgBurn{
		Sender:          moduleAddr.String(),
		Amount:          sdk.NewCoin(denom, amount),
		BurnFromAddress: sourceAddr,
	}

	handler := k.router.Handler(msgBurn)
	if handler == nil {
		return fmt.Errorf("can't route a burn message")
	}

	// Ignore resp since it is empty in this method
	_, err = handler(ctx, msgBurn)
	if err != nil {
		return fmt.Errorf("can't execute a burn message: %s", err)
	}

	return nil
}
