package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (k Keeper) SetEpochTwapPrice(ctx sdk.Context, epoch int64, poolId uint64, price sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.TokenPriceTwapEpochPrefix(epoch))
	priceRecord := types.EpochTwapPrice{
		Epoch:  epoch,
		PoolId: poolId,
		Price:  price,
	}
	bz, err := proto.Marshal(&priceRecord)
	if err != nil {
		panic(err)
	}
	prefixStore.Set(sdk.Uint64ToBigEndian(poolId), bz)
}

func (k Keeper) DeleteEpochTwapPrice(ctx sdk.Context, epoch int64, poolId uint64) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.TokenPriceTwapEpochPrefix(epoch))
	prefixStore.Delete(sdk.Uint64ToBigEndian(poolId))
}

func (k Keeper) GetEpochTwapPrice(ctx sdk.Context, epoch int64, poolId uint64) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.TokenPriceTwapEpochPrefix(epoch))
	bz := prefixStore.Get(sdk.Uint64ToBigEndian(poolId))
	if bz == nil {
		return sdk.ZeroDec()
	}
	priceRecord := types.EpochTwapPrice{}
	err := proto.Unmarshal(bz, &priceRecord)
	if err != nil {
		panic(err)
	}
	return priceRecord.Price
}

func (k Keeper) GetAllEpochTwapPrices(ctx sdk.Context, epoch int64) []types.EpochTwapPrice {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.TokenPriceTwapEpochPrefix(epoch))
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	priceRecords := []types.EpochTwapPrice{}
	for ; iterator.Valid(); iterator.Next() {
		priceRecord := types.EpochTwapPrice{}

		err := proto.Unmarshal(iterator.Value(), &priceRecord)
		if err != nil {
			panic(err)
		}

		priceRecords = append(priceRecords, priceRecord)
	}
	return priceRecords
}

func (k Keeper) GetAllTwapPrices(ctx sdk.Context) []types.EpochTwapPrice {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPriceTwap)
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	priceRecords := []types.EpochTwapPrice{}
	for ; iterator.Valid(); iterator.Next() {
		priceRecord := types.EpochTwapPrice{}

		err := proto.Unmarshal(iterator.Value(), &priceRecord)
		if err != nil {
			panic(err)
		}

		priceRecords = append(priceRecords, priceRecord)
	}
	return priceRecords
}
