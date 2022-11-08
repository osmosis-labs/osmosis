package concentrated_liquidity

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) GetPoolTickKVStore(ctx sdk.Context, poolId uint64) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	prefixBz := types.KeyTickPrefix(poolId)
	return prefix.NewStore(store, prefixBz)
}
