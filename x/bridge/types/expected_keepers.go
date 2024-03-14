package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type AccountKeeper interface {
	// GetModuleAccount is used to create x/bridge module account
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	// GetModuleAddress is used to get the module account
	// to use it as the admin for denoms in x/tokenfactory.
	GetModuleAddress(name string) sdk.AccAddress
}

type TokenFactoryKeeper interface {
	CreateDenom(ctx sdk.Context, creatorAddr string, subdenom string) (newTokenDenom string, err error)
	Mint(ctx sdk.Context, sender string, amount sdk.Coin, mintTo string) error
	Burn(ctx sdk.Context, sender string, amount sdk.Coin, burnFrom string) error
}
