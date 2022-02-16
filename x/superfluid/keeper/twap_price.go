package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

// This function calculates the osmo equivalent worth of an LP share.
// It is intended to eventually use the TWAP of the worth of an LP share
// once that is exposed from the gamm module.
func (k Keeper) calculateOsmoBackingPerShare(pool gammtypes.PoolI, osmoInPool gammtypes.PoolAsset) sdk.Dec {
	twap := osmoInPool.Token.Amount.ToDec().Quo(pool.GetTotalShares().Amount.ToDec())
	return twap
}

func (k Keeper) SetEpochOsmoEquivalentTWAP(ctx sdk.Context, epoch int64, denom string, price sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPriceTwap)
	priceRecord := types.EpochOsmoEquivalentTWAP{
		EpochNumber:    epoch,
		Denom:          denom,
		EpochTwapPrice: price,
	}
	bz, err := proto.Marshal(&priceRecord)
	if err != nil {
		panic(err)
	}
	prefixStore.Set([]byte(denom), bz)
}

func (k Keeper) DeleteEpochOsmoEquivalentTWAP(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPriceTwap)
	prefixStore.Delete([]byte(denom))
}

func (k Keeper) GetEpochOsmoEquivalentTWAP(ctx sdk.Context, denom string) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPriceTwap)
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

func (k Keeper) GetAllEpochOsmoEquivalentTWAPs(ctx sdk.Context, epoch int64) []types.EpochOsmoEquivalentTWAP {
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
