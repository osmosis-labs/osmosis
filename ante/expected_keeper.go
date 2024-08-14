package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TreasuryKeeper for tax charging & recording
type TreasuryKeeper interface {
	GetTaxRate(ctx sdk.Context) (taxRate sdk.Dec)
}

// OracleKeeper for feeder validation
type OracleKeeper interface {
	ValidateFeeder(ctx sdk.Context, feederAddr sdk.AccAddress, validatorAddr sdk.ValAddress) error
	// GetMelodyExchangeRate returns the exchange rate of the given denom to melody. Returned value is in melody.
	GetMelodyExchangeRate(ctx sdk.Context, denom string) (price sdk.Dec, err error)
}

// BankKeeper defines the contract needed for supply related APIs (noalias)
type BankKeeper interface {
	IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error
	SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule string, recipientModule string, amt sdk.Coins) error
}
