package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v23/x/tokenfactory/types"
)

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
	asset, ok := params.GetAsset(assetID)
	if !ok {
		return errorsmod.Wrapf(types.ErrInvalidAssetID, "Asset not found %s", assetID.Name())
	}
	if !asset.Status.InboundActive() {
		return errorsmod.Wrapf(types.ErrInvalidAssetStatus, "Inbound transfers are disabled for this asset")
	}

	// ---> Concurrency protection?

	// Get the transfer info from the store and update it properly
	transfer, found := k.GetInboundTransfer(ctx, externalID)
	if !found {
		// If the transfer is new, then create it
		transfer = types.NewInboundTransfer(externalID, destAddr, assetID, amount)
	}

	transfer.Voters = append(transfer.Voters, sender)
	transfer.Finalized = uint64(len(transfer.Voters)) >= params.VotesNeeded
	k.UpsertInboundTransfer(ctx, transfer)

	// <--- Concurrency protection?

	// Check if the transfer got a sufficient number of votes
	if !transfer.Finalized {
		// The transfer still needs more votes to be processed
		return nil
	}

	// The transfer is ready to be processed!

	// Perform tokenfactory mint
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	denom, err := tokenfactorytypes.GetTokenDenom(moduleAddr.String(), assetID.Name())
	if err != nil {
		return errorsmod.Wrapf(types.ErrTokenfactory, "Can't create a tokenfacroty denom for %s", assetID.Name())
	}

	msgMint := &tokenfactorytypes.MsgMint{
		Sender:        moduleAddr.String(),
		Amount:        sdk.NewCoin(denom, amount),
		MintToAddress: destAddr,
	}

	handler := k.router.Handler(msgMint)
	if handler == nil {
		return errorsmod.Wrapf(types.ErrTokenfactory, "Can't route a mint message")
	}

	// Ignore resp since it is empty in this method
	_, err = handler(ctx, msgMint)
	if err != nil {
		return errorsmod.Wrapf(types.ErrTokenfactory, "Can't execute a mint message: %s", err)
	}

	return nil
}

// GetInboundTransfer returns the transfer by externalID.
func (k Keeper) GetInboundTransfer(ctx sdk.Context, externalID string) (types.InboundTransfer, bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.InboundTransferKey(externalID))
	if b == nil {
		return types.InboundTransfer{}, false
	}

	var inboundTranfer types.InboundTransfer
	err := proto.Unmarshal(b, &inboundTranfer)
	if err != nil {
		// TODO: is it ok to panic there? does it make sense to log and return false?
		panic(err)
	}

	return inboundTranfer, true
}

// UpsertInboundTransfer updates or inserts the inbound transfer depending on
// whether it is already presented in the store or not.
func (k Keeper) UpsertInboundTransfer(ctx sdk.Context, t types.InboundTransfer) {
	store := ctx.KVStore(k.storeKey)
	key := types.InboundTransferKey(t.ExternalId)

	// Save the transfer
	value, err := proto.Marshal(&t)
	if err != nil {
		// TODO: is it ok to panic there? does it make sense to log and return error?
		panic(err)
	}
	store.Set(key, value)
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
