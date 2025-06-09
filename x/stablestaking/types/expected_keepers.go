package types

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	epochstypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
)

// AccountKeeper is expected keeper for auth module
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
}

// BankKeeper defines expected supply keeper
type BankKeeper interface {
	SendCoinsFromModuleToModule(ctx context.Context, senderModule string, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	GetSupply(ctx context.Context, denom string) sdk.Coin

	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error

	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// OracleKeeper defines expected oracle keeper
type OracleKeeper interface {
	// GetMelodyExchangeRate returns the exchange rate of the given denom to melody. Returned value is in melody.
	GetMelodyExchangeRate(ctx sdk.Context, denom string) (price osmomath.Dec, err error)
	GetTobinTax(ctx sdk.Context, denom string) (tobinTax osmomath.Dec, err error)
}

// EpochKeeper defines the contract needed to be fulfilled for epochs keeper.
type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
}
