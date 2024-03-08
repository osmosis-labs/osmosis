package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
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

	return k.tokenFactoryKeeper.Mint(
		ctx,
		moduleAddr.String(),
		sdk.NewCoin(asset.Name(), amount),
		destAddr,
	)
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

	return k.tokenFactoryKeeper.Burn(
		ctx,
		moduleAddr.String(),
		sdk.NewCoin(asset.Name(), amount),
		sourceAddr,
	)
}
