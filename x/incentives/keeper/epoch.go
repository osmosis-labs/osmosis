package keeper

import (
	"github.com/c-osmosis/osmosis/x/incentives/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetCurrentEpochInfo(ctx sdk.Context, currentEpoch, epochBlockNum int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyCurrentEpoch, sdk.Uint64ToBigEndian(uint64(currentEpoch)))
	store.Set(types.KeyEpochBeginBlock, sdk.Uint64ToBigEndian(uint64(epochBlockNum)))
}

func (k Keeper) GetCurrentEpochInfo(ctx sdk.Context) (int64, int64) {
	store := ctx.KVStore(k.storeKey)
	epochKeyBz := store.Get(types.KeyCurrentEpoch)
	if epochKeyBz == nil {
		return 0, 0
	}

	beginBlockBz := store.Get(types.KeyEpochBeginBlock)
	if beginBlockBz == nil {
		return 0, 0
	}

	return int64(sdk.BigEndianToUint64(epochKeyBz)), int64(sdk.BigEndianToUint64(beginBlockBz))
}
