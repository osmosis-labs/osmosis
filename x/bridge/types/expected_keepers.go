package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type BankKeeper interface {
	// HasBalance is used to check whether the sender has
	// sufficient balance.
	//
	// TODO: maybe this check is redundant since it's already
	// 	a part of x/tokenfactory logics?
	HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool
}

type AccountKeeper interface {
	// GetModuleAccount is used to get x/bridge module account
	// to use it as the admin for denoms in x/tokenfactory.
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	GetModuleAddress(name string) sdk.AccAddress
}

type TokenFactoryKeeper interface {
	CreateDenom(ctx sdk.Context, creatorAddr string, subdenom string) (newTokenDenom string, err error)
	//Mint(ctx sdk.Context, sender string, amount sdk.Coin, mintTo string) error
	//Burn(ctx sdk.Context, sender string, amount sdk.Coin, burnFrom string) error
}
