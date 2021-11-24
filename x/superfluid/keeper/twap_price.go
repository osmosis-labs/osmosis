package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (k Keeper) SetEpochOsmoEquivalentTWAP(ctx sdk.Context, epoch int64, denom string, price sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.TokenPriceTwapEpochPrefix(epoch))
	priceRecord := types.EpochOsmoEquivalentTWAP{
		Epoch:          epoch,
		Denom:          denom,
		EpochTwapPrice: price,
	}
	bz, err := proto.Marshal(&priceRecord)
	if err != nil {
		panic(err)
	}
	prefixStore.Set([]byte(denom), bz)
}

func (k Keeper) DeleteEpochOsmoEquivalentTWAP(ctx sdk.Context, epoch int64, denom string) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.TokenPriceTwapEpochPrefix(epoch))
	prefixStore.Delete([]byte(denom))
}

func (k Keeper) GetEpochOsmoEquivalentTWAP(ctx sdk.Context, epoch int64, denom string) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.TokenPriceTwapEpochPrefix(epoch))
	bz := prefixStore.Get([]byte(denom))
	if bz == nil {
		return sdk.ZeroDec()
	}
	priceRecord := types.EpochOsmoEquivalentTWAP{}
	err := proto.Unmarshal(bz, &priceRecord)
	if err != nil {
		panic(err)
	}
	return priceRecord.EpochTwapPrice
}

func (k Keeper) GetLastEpochOsmoEquivalentTWAP(ctx sdk.Context, denom string) types.EpochOsmoEquivalentTWAP {
	params := k.GetParams(ctx)
	epochInfo := k.ek.GetEpochInfo(ctx, params.RefreshEpochIdentifier)

	return types.EpochOsmoEquivalentTWAP{
		Epoch:          epochInfo.CurrentEpoch - 1,
		Denom:          denom,
		EpochTwapPrice: k.GetEpochOsmoEquivalentTWAP(ctx, epochInfo.CurrentEpoch-1, denom),
	}
}

func (k Keeper) GetAllEpochOsmoEquivalentTWAPs(ctx sdk.Context, epoch int64) []types.EpochOsmoEquivalentTWAP {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.TokenPriceTwapEpochPrefix(epoch))
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	priceRecords := []types.EpochOsmoEquivalentTWAP{}
	for ; iterator.Valid(); iterator.Next() {
		priceRecord := types.EpochOsmoEquivalentTWAP{}

		err := proto.Unmarshal(iterator.Value(), &priceRecord)
		if err != nil {
			panic(err)
		}

		priceRecords = append(priceRecords, priceRecord)
	}
	return priceRecords
}

func (k Keeper) GetAllOsmoEquivalentTWAPs(ctx sdk.Context) []types.EpochOsmoEquivalentTWAP {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPriceTwap)
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	priceRecords := []types.EpochOsmoEquivalentTWAP{}
	for ; iterator.Valid(); iterator.Next() {
		priceRecord := types.EpochOsmoEquivalentTWAP{}

		err := proto.Unmarshal(iterator.Value(), &priceRecord)
		if err != nil {
			panic(err)
		}

		priceRecords = append(priceRecords, priceRecord)
	}
	return priceRecords
}
