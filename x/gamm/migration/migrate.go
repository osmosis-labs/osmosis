package migration

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	oldbalancer "github.com/osmosis-labs/osmosis/v15/x/gamm/v2types/balancer"
	oldstableswap "github.com/osmosis-labs/osmosis/v15/x/gamm/v2types/stableswap"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

func RemoveExitFee(ctx sdk.Context, keeper gammkeeper.Keeper) error {
	store := ctx.KVStore(keeper.GetStoreKey(ctx))
	cdc := keeper.GetCodec(ctx)

	iterator := sdk.KVStorePrefixIterator(store, gammtypes.KeyPrefixPools)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		pool, err := keeper.UnmarshalPool(iterator.Value())
		if err != nil {
			return sdkerrors.Wrapf(err, "unable to unmarshal pool (%v)", iterator.Key())
		}
		poolKey := iterator.Key()

		switch pool.GetType() {
		case poolmanagertypes.Balancer:
			oldPool, ok := pool.(*oldbalancer.Pool)
			if !ok {
				return sdkerrors.Wrapf(err, "unable to unmarshal pool (%v) using old balancer data type", iterator.Key())
			}
			if !oldPool.PoolParams.ExitFee.Equal(sdk.ZeroDec()) {
				store.Delete(poolKey)
			} else {
				// Convert and serialize using the new type
				newPool := convertToNewBalancerPool(*oldPool)
				newPoolBz, err := cdc.MarshalInterface(&newPool)
				if err != nil {
					return sdkerrors.Wrapf(err, "unable to marshal pool (%v) using new balancer data type", iterator.Key())
				}
				store.Set(poolKey, newPoolBz)
			}
		case poolmanagertypes.Stableswap:
			oldPool, ok := pool.(*oldstableswap.Pool)
			if !ok {
				return sdkerrors.Wrapf(err, "unable to unmarshal pool (%v) using old stableswap data type", iterator.Key())
			}
			if !oldPool.PoolParams.ExitFee.Equal(sdk.ZeroDec()) {
				store.Delete(poolKey)
			} else {
				// Convert and serialize using the new type
				newPool := convertToNewStableSwapPool(*oldPool)
				newPoolBz, err := cdc.MarshalInterface(&newPool)
				if err != nil {
					return sdkerrors.Wrapf(err, "unable to marshal pool (%v) using new stableswap data type", iterator.Key())
				}
				store.Set(poolKey, newPoolBz)
			}
		}
	}
	return nil
}
