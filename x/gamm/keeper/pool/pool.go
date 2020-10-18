package pool

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type Pool interface {
	Viewer

	CreatePool(sdk.Context, sdk.AccAddress, sdk.Dec, []types.TokenInfo) error
	JoinPool(sdk.Context, sdk.AccAddress, sdk.AccAddress, sdk.Int, []types.MaxAmountIn) error
	ExitPool(sdk.Context, sdk.AccAddress, sdk.AccAddress, sdk.Int, []types.MinAmountOut) error
}

type pool struct {
	viewer

	cdc           codec.BinaryMarshaler
	storeKey      sdk.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    bankkeeper.Keeper
}

func NewPool(
	cdc codec.BinaryMarshaler,
	storeKey sdk.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
) Pool {
	return pool{
		viewer: viewer{
			cdc:      cdc,
			storeKey: storeKey,
		},
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (p pool) CreatePool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	swapFee sdk.Dec,
	tokenInfo []types.TokenInfo,
) error {
	return nil
}

func (p pool) JoinPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	targetPool sdk.AccAddress,
	poolAmountOut sdk.Int,
	maxAmountsIn []types.MaxAmountIn,
) error {
	return nil
}

func (p pool) ExitPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	targetPool sdk.AccAddress,
	poolAmountIn sdk.Int,
	minAmountsOut []types.MinAmountOut,
) error {
	return nil
}
