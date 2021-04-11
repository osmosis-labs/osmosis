package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/c-osmosis/osmosis/x/gamm/types"
)

func (k Keeper) GetPool(ctx sdk.Context, poolId uint64) (types.PoolAccountI, error) {
	acc := k.accountKeeper.GetAccount(ctx, types.NewPoolAddress(poolId))
	if acc == nil {
		return nil, sdkerrors.Wrapf(types.ErrPoolNotFound, "pool %d does not exist", poolId)
	}

	poolAcc, ok := acc.(types.PoolAccountI)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrPoolNotFound, "pool %d does not exist", poolId)
	}

	return poolAcc, nil
}

func (k Keeper) SetPool(ctx sdk.Context, poolAcc types.PoolAccountI) error {
	// Make sure that pool exists
	_, err := k.GetPool(ctx, poolAcc.GetId())
	if err != nil {
		return err
	}

	k.accountKeeper.SetAccount(ctx, poolAcc)
	return nil
}

func (k Keeper) NewPool(ctx sdk.Context, poolParams types.PoolParams) (types.PoolAccountI, error) {
	poolId := k.getNextPoolNumber(ctx)

	acc := k.accountKeeper.GetAccount(ctx, types.NewPoolAddress(poolId))
	if acc != nil {
		return nil, sdkerrors.Wrapf(types.ErrPoolAlreadyExist, "pool %d already exist", poolId)
	}

	poolAcc := types.NewPoolAccount(poolId, poolParams)
	poolAcc = k.accountKeeper.NewAccount(ctx, poolAcc).(types.PoolAccountI)

	k.accountKeeper.SetAccount(ctx, poolAcc)

	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetKeyPaginationPoolNumbers(poolId), sdk.Uint64ToBigEndian(poolId))

	return poolAcc, nil
}

// getNextPoolNumber returns the next pool number
func (k Keeper) getNextPoolNumber(ctx sdk.Context) uint64 {
	var poolNumber uint64
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GlobalPoolNumber)
	if bz == nil {
		// initialize the account numbers
		poolNumber = 1
	} else {
		val := gogotypes.UInt64Value{}

		err := k.cdc.UnmarshalBinaryBare(bz, &val)
		if err != nil {
			panic(err)
		}

		poolNumber = val.GetValue()
	}

	bz = k.cdc.MustMarshalBinaryBare(&gogotypes.UInt64Value{Value: poolNumber + 1})
	store.Set(types.GlobalPoolNumber, bz)

	return poolNumber
}
