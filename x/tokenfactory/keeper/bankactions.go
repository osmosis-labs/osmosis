package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/osmosis-labs/osmosis/v23/x/tokenfactory/types"
)

func (k Keeper) Mint(ctx sdk.Context, sender string, amount sdk.Coin, mintTo string) error {
	// pay some extra gas cost to give a better error here.
	_, denomExists := k.bankKeeper.GetDenomMetaData(ctx, amount.GetDenom())
	if !denomExists {
		return types.ErrDenomDoesNotExist.Wrapf("denom: %s", amount.GetDenom())
	}

	authorityMetadata, err := k.GetAuthorityMetadata(ctx, amount.GetDenom())
	if err != nil {
		return err
	}

	if sender != authorityMetadata.GetAdmin() {
		return types.ErrUnauthorized
	}

	if mintTo == "" {
		mintTo = sender
	}

	err = k.mintTo(ctx, amount, mintTo)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) mintTo(ctx sdk.Context, amount sdk.Coin, mintTo string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(amount.Denom)
	if err != nil {
		return err
	}

	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(mintTo)
	if err != nil {
		return err
	}

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName,
		addr,
		sdk.NewCoins(amount))
}

func (k Keeper) Burn(ctx sdk.Context, sender string, amount sdk.Coin, burnFrom string) error {
	authorityMetadata, err := k.GetAuthorityMetadata(ctx, amount.GetDenom())
	if err != nil {
		return err
	}

	if sender != authorityMetadata.GetAdmin() {
		return types.ErrUnauthorized
	}

	if burnFrom == "" {
		burnFrom = sender
	}

	accountI := k.accountKeeper.GetAccount(ctx, sdk.AccAddress(burnFrom))
	_, ok := accountI.(authtypes.ModuleAccountI)
	if ok {
		return types.ErrBurnFromModuleAccount
	}

	err = k.burnFrom(ctx, amount, burnFrom)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) burnFrom(ctx sdk.Context, amount sdk.Coin, burnFrom string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(amount.Denom)
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(burnFrom)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx,
		addr,
		types.ModuleName,
		sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	return k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
}

func (k Keeper) forceTransfer(ctx sdk.Context, amount sdk.Coin, fromAddr string, toAddr string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(amount.Denom)
	if err != nil {
		return err
	}

	fromSdkAddr, err := sdk.AccAddressFromBech32(fromAddr)
	if err != nil {
		return err
	}

	toSdkAddr, err := sdk.AccAddressFromBech32(toAddr)
	if err != nil {
		return err
	}

	return k.bankKeeper.SendCoins(ctx, fromSdkAddr, toSdkAddr, sdk.NewCoins(amount))
}
