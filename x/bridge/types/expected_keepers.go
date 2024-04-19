package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AccountKeeper interface {
	// GetModuleAddress is used to get the module account
	// to use it as the admin for denoms in x/tokenfactory.
	GetModuleAddress(name string) sdk.AccAddress
}
