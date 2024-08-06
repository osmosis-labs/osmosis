package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	oracletypes "github.com/osmosis-labs/osmosis/v23/x/oracle/types"
)

// AccountKeeper expected account keeper
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
}

// BankKeeper expected bank keeper
type BankKeeper interface {
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule string, recipientModule string, amt sdk.Coins) error
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
}

// MarketKeeper expected market keeper
type MarketKeeper interface {
	ComputeInternalSwap(ctx sdk.Context, offerCoin sdk.DecCoin, askDenom string) (sdk.DecCoin, error)
}

// StakingKeeper expected keeper for staking module
type StakingKeeper interface {
	TotalBondedTokens(sdk.Context) math.Int // total bonded tokens within the validator set
}

// DistributionKeeper expected keeper for distribution module
type DistributionKeeper interface {
	GetFeePool(ctx sdk.Context) (feePool distrtypes.FeePool)
	SetFeePool(ctx sdk.Context, feePool distrtypes.FeePool)
}

// OracleKeeper defines expected oracle keeper
type OracleKeeper interface {
	Whitelist(ctx sdk.Context) (res oracletypes.DenomList)

	// only used for test purpose
	SetLunaExchangeRate(ctx sdk.Context, denom string, exchangeRate sdk.Dec)
	SetWhitelist(ctx sdk.Context, whitelist oracletypes.DenomList)
}
