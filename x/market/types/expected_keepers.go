package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper is expected keeper for auth module
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI // only used for simulation
}

// BankKeeper defines expected supply keeper
type BankKeeper interface {
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule string, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error

	// only used for simulation
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	IsSendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool
}

// OracleKeeper defines expected oracle keeper
type OracleKeeper interface {
	// GetMelodyExchangeRate returns the exchange rate of the given denom to melody. Returned value is in melody.
	GetMelodyExchangeRate(ctx sdk.Context, denom string) (price sdk.Dec, err error)
	GetTobinTax(ctx sdk.Context, denom string) (tobinTax sdk.Dec, err error)

	// only used for simulation
	IterateNoteExchangeRates(ctx sdk.Context, handler func(denom string, exchangeRate sdk.Dec) (stop bool))
	// SetMelodyExchangeRate sets the exchange rate of the given denom to melody. The exchange rate is in melody.
	SetMelodyExchangeRate(ctx sdk.Context, denom string, exchangeRate sdk.Dec)
	SetTobinTax(ctx sdk.Context, denom string, tobinTax sdk.Dec)
}
