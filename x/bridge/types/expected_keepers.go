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
}

type TokenFactoryKeeper interface {
	// TODO: Expose TokenFactory methods?
}
