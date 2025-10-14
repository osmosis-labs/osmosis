package types // noalias

import (
	context "context"

	"github.com/osmosis-labs/osmosis/osmomath"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool

	SetModuleAccount(context.Context, sdk.ModuleAccountI)
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI
}

// BankKeeper defines the contract needed to be fulfilled for banking and supply
// dependencies.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, name string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error
	AddSupplyOffset(ctx context.Context, denom string, offsetAmount osmomath.Int)
	GetSupply(ctx context.Context, denom string) sdk.Coin
	GetSupplyWithOffset(ctx context.Context, denom string) sdk.Coin
}

// CommunityPoolKeeper is just an alias to the cosmos SDK distribution keeper pointer.
type CommunityPoolKeeper = *distrkeeper.Keeper

// EpochKeeper defines the contract needed to be fulfilled for epochs keeper.
type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
}

// StakingKeeper defines the contract needed to query staking information.
type StakingKeeper interface {
	IterateDelegations(ctx context.Context, delegator sdk.AccAddress, fn func(int64, stakingtypes.DelegationI) bool) error
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)
}
