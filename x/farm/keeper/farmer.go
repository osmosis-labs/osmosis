package keeper

import (
	"github.com/c-osmosis/osmosis/x/farm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) NewFarmer(ctx sdk.Context, farmId uint64, currentPeriod uint64, address sdk.AccAddress, share sdk.Int) types.Farmer {
	farmer := types.Farmer{
		FarmId:              farmId,
		Address:             address.String(),
		Share:               share,
		LastWithdrawnPeriod: currentPeriod,
	}

	k.setFarmer(ctx, farmer)
	return farmer
}

func (k Keeper) GetFarmer(ctx sdk.Context, farmId uint64, address sdk.AccAddress) (types.Farmer, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetFarmerStoreKey(farmId, address))
	if len(bz) == 0 {
		return types.Farmer{}, types.ErrNoFarmerExist
	}

	farmer := types.Farmer{}
	k.cdc.MustUnmarshalBinaryBare(bz, &farmer)
	return farmer, nil
}

func (k Keeper) setFarmer(ctx sdk.Context, farmer types.Farmer) {
	store := ctx.KVStore(k.storeKey)
	accAddress, err := sdk.AccAddressFromBech32(farmer.Address)
	if err != nil {
		panic(err)
	}
	key := types.GetFarmerStoreKey(farmer.FarmId, accAddress)

	bz := k.cdc.MustMarshalBinaryBare(&farmer)
	store.Set(key, bz)
}

func (k Keeper) IterateFarmers(ctx sdk.Context, handler func(farm types.Farmer) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.FarmerPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var farmer types.Farmer
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &farmer)
		if handler(farmer) {
			break
		}
	}
}

func (k Keeper) IterateFarmersInFarm(ctx sdk.Context, farmId uint64, handler func(farm types.Farmer) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, append(types.FarmerPrefix, sdk.Uint64ToBigEndian(farmId)...))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var farmer types.Farmer
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &farmer)
		if handler(farmer) {
			break
		}
	}
}
