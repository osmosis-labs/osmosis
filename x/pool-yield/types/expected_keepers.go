package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	farmtypes "github.com/c-osmosis/osmosis/x/farm/types"
	gammtypes "github.com/c-osmosis/osmosis/x/gamm/types"
)

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) types.ModuleAccountI
}

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

type GAMMKeeper interface {
	GetPool(ctx sdk.Context, poolId uint64) (gammtypes.PoolAccountI, error)
}

type FarmKeeper interface {
	NewFarm(ctx sdk.Context) (farmtypes.Farm, error)
	GetFarm(ctx sdk.Context, farmId uint64) (farmtypes.Farm, error)

	DepositShareToFarm(ctx sdk.Context, farmId uint64, address sdk.AccAddress, share sdk.Int) (rewards sdk.Coins, err error)
	WithdrawShareFromFarm(ctx sdk.Context, farmId uint64, address sdk.AccAddress, share sdk.Int) (rewards sdk.Coins, err error)

	AllocateAssetsFromModuleToFarm(ctx sdk.Context, farmId uint64, moduleName string, assets sdk.Coins) error
}
