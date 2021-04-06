package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	farmtypes "github.com/c-osmosis/osmosis/x/farm/types"
	gammtypes "github.com/c-osmosis/osmosis/x/gamm/types"
)

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) authtypes.ModuleAccountI
}

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin

	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule string, recipientModule string, amt sdk.Coins) error
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

type DistrKeeper interface {
	GetFeePool(ctx sdk.Context) (feePool distrtypes.FeePool)
	SetFeePool(ctx sdk.Context, feePool distrtypes.FeePool)
}
