package simtypes

import (
	"context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	ibctestingtypes "github.com/cosmos/ibc-go/v8/testing/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

type AppCreator = func(homepath string, legacyInvariantPeriod uint, baseappOptions ...func(*baseapp.BaseApp)) App

type App interface {
	GetBaseApp() *baseapp.BaseApp
	AppCodec() codec.Codec
	GetAccountKeeper() AccountKeeper
	GetBankKeeper() BankKeeper
	GetStakingKeeper() ibctestingtypes.StakingKeeper
	GetSDKStakingKeeper() stakingkeeper.Keeper
	ModuleManager() module.Manager
	GetPoolManagerKeeper() PoolManagerKeeper
	GetSubspace(moduleName string) paramtypes.Subspace
}

type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetAllAccounts(ctx context.Context) []sdk.AccountI
}

type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	// TODO: Revisit
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

type PoolManagerKeeper interface {
	CreatePool(ctx sdk.Context, msg poolmanagertypes.CreatePoolMsg) (uint64, error)
}
