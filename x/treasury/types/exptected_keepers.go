package types

import (
	"context"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/osmosis-labs/osmosis/osmomath"

	oracletypes "github.com/osmosis-labs/osmosis/v27/x/oracle/types"
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

	// only used for simulation
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	IsSendEnabledCoin(ctx context.Context, coin sdk.Coin) bool
}

// MarketKeeper expected market keeper
type MarketKeeper interface {
	// GetExchangePoolBalance returns the balance of the exchange pool. In Melody.
	GetExchangePoolBalance(ctx sdk.Context) sdk.Coin
	// GetExchangeRequirement calculates the total amount of Melody asset required to back the assets in the exchange pool.
	GetExchangeRequirement(ctx sdk.Context) osmomath.Dec
	ComputeInternalSwap(ctx sdk.Context, offerCoin sdk.DecCoin, askDenom string) (sdk.DecCoin, error)
}

// StakingKeeper expected keeper for staking module
type StakingKeeper interface {
	TotalBondedTokens(context.Context) math.Int // total bonded tokens within the validator set
}

// DistributionKeeper expected keeper for distribution module
type DistributionKeeper interface {
	GetFeePool(ctx context.Context) (feePool distrtypes.FeePool)
	SetFeePool(ctx context.Context, feePool distrtypes.FeePool)
}

// OracleKeeper defines expected oracle keeper
type OracleKeeper interface {
	Whitelist(ctx sdk.Context) (res oracletypes.DenomList)

	// only used for test purpose
	SetWhitelist(ctx sdk.Context, whitelist oracletypes.DenomList)
}
