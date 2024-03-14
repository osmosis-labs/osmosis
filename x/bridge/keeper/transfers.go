package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v23/x/tokenfactory/types"
)

func (k Keeper) InboundTransfer(
	ctx sdk.Context,
	destAddr string,
	asset types.Asset,
	amount math.Int,
) error {
	params := k.GetParams(ctx)

	assetWithStatus, ok := params.GetAsset(asset)
	if !ok {
		return errorsmod.Wrapf(types.ErrInvalidAsset, "Asset not found %s", asset.Name())
	}

	if !assetWithStatus.AssetStatus.InboundActive() {
		return errorsmod.Wrapf(types.ErrInvalidAssetStatus, "Inbound transfers are disabled for this asset")
	}

	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	denom, err := tokenfactorytypes.GetTokenDenom(moduleAddr.String(), asset.Name())
	if err != nil {
		return errorsmod.Wrapf(types.ErrTokenfactory, "Can't create a tokenfacroty denom for %s", asset.Name())
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

	// ignore resp since it is empty in this method
	// TODO: double-check if we need to handle the response
	_, err = handler(ctx, msgMint)
	if err != nil {
		return errorsmod.Wrapf(types.ErrTokenfactory, "Can't execute a mint message: %s", err)
	}

	return nil
}

func (k Keeper) OutboundTransfer(
	ctx sdk.Context,
	sourceAddr string,
	asset types.Asset,
	amount math.Int,
) error {
	params := k.GetParams(ctx)

	assetWithStatus, ok := params.GetAsset(asset)
	if !ok {
		return errorsmod.Wrapf(types.ErrInvalidAsset, "Asset not found %s", asset.Name())
	}

	if !assetWithStatus.AssetStatus.OutboundActive() {
		return errorsmod.Wrapf(types.ErrInvalidAssetStatus, "Outbound transfers are disabled for this asset")
	}

	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	denom, err := tokenfactorytypes.GetTokenDenom(moduleAddr.String(), asset.Name())
	if err != nil {
		return errorsmod.Wrapf(types.ErrTokenfactory, "Can't create a tokenfacroty denom for %s", asset.Name())
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
	// TODO: double-check if we need to handle the response
	_, err = handler(ctx, msgBurn)
	if err != nil {
		return errorsmod.Wrapf(types.ErrTokenfactory, "Can't execute a burn message: %s", err)
	}

	return nil
}
