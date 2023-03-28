package simtypes

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

type AppCreator = func(homepath string, legacyInvariantPeriod uint, baseappOptions ...func(*baseapp.BaseApp)) App

type App interface {
	GetBaseApp() *baseapp.BaseApp
	AppCodec() codec.Codec
	GetAccountKeeper() AccountKeeper
	GetBankKeeper() bankkeeper.BaseKeeper
	GetStakingKeeper() stakingkeeper.Keeper
	ModuleManager() module.Manager
}

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	// GetAllAccounts(ctx sdk.Context) []authtypes.AccountI
}
