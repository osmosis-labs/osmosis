package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AccountKeeper interface {
	GetAllAccounts(ctx context.Context) []sdk.AccountI
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins

	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error
}

type CommunityPoolKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}
