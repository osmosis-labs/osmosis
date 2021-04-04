package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	farmtypes "github.com/c-osmosis/osmosis/x/farm/types"
	gammtypes "github.com/c-osmosis/osmosis/x/gamm/types"
)

type GAMMKeeper interface {
	GetPool(ctx sdk.Context, poolId uint64) (gammtypes.PoolAccountI, error)
}

type FarmKeeper interface {
	NewFarm(ctx sdk.Context) (farmtypes.Farm, error)
	GetFarm(ctx sdk.Context, farmId uint64) (farmtypes.Farm, error)

	DepositShareToFarm(ctx sdk.Context, farmId uint64, address sdk.AccAddress, share sdk.Int) (rewards sdk.Coins, err error)
	WithdrawShareFromFarm(ctx sdk.Context, farmId uint64, address sdk.AccAddress, share sdk.Int) (rewards sdk.Coins, err error)
}
