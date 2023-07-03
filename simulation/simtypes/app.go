package simtypes

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

type AppCreator = func(homepath string, legacyInvariantPeriod uint, baseappOptions ...func(*baseapp.BaseApp)) App

type App interface {
	GetBaseApp() *baseapp.BaseApp
	AppCodec() codec.Codec
	GetAccountKeeper() AccountKeeper
	GetBankKeeper() BankKeeper
	GetStakingKeeper() stakingkeeper.Keeper
	ModuleManager() module.Manager
	GetPoolManagerKeeper() PoolManagerKeeper
}

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	// GetAllAccounts(ctx sdk.Context) []authtypes.AccountI
}

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// TODO: Revisit
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}

type PoolManagerKeeper interface {
	CreatePool(ctx sdk.Context, msg poolmanagertypes.CreatePoolMsg) (uint64, error)
}
