package keeper

import (
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"

	"github.com/c-osmosis/osmosis/x/farm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryMarshaler

	ak types.AccountKeeper
	bk types.BankKeeper
}

func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, ak types.AccountKeeper, bk types.BankKeeper) Keeper {
	if ak.GetModuleAddress(types.ModuleName) == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		ak:       ak,
		bk:       bk,
	}
}

func (k Keeper) NewFarm(ctx sdk.Context) (types.Farm, error) {
	farmId := k.GetNextFarmId(ctx)
	farm := types.Farm{
		FarmId:         farmId,
		TotalShare:     sdk.NewInt(0),
		CurrentPeriod:  1,
		CurrentRewards: sdk.DecCoins{},
	}

	store := ctx.KVStore(k.storeKey)

	store.Set(types.GetFarmStoreKey(farmId), k.cdc.MustMarshalBinaryBare(&farm))
	return farm, nil
}

func (k Keeper) GetFarm(ctx sdk.Context, farmId uint64) (types.Farm, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetFarmStoreKey(farmId))
	if len(bz) == 0 {
		return types.Farm{}, types.ErrNoFarmExist
	}

	farm := types.Farm{}
	err := k.cdc.UnmarshalBinaryBare(bz, &farm)
	if err != nil {
		return types.Farm{}, err
	}

	return farm, nil
}

func (k Keeper) IterateFarms(ctx sdk.Context, handler func(farm types.Farm) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.FarmPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var farm types.Farm
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &farm)
		if handler(farm) {
			break
		}
	}
}

func (k Keeper) setFarm(ctx sdk.Context, farm types.Farm) error {
	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.GetFarmStoreKey(farm.FarmId)) {
		return types.ErrNoFarmExist
	}

	store.Set(types.GetFarmStoreKey(farm.FarmId), k.cdc.MustMarshalBinaryBare(&farm))
	return nil
}

func (k Keeper) GetNextFarmId(ctx sdk.Context) uint64 {
	var farmNumber uint64
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GlobalFarmNumber)
	if bz == nil {
		// initialize the account numbers
		farmNumber = 1
	} else {
		val := gogotypes.UInt64Value{}

		err := k.cdc.UnmarshalBinaryBare(bz, &val)
		if err != nil {
			panic(err)
		}

		farmNumber = val.GetValue()
	}

	bz = k.cdc.MustMarshalBinaryBare(&gogotypes.UInt64Value{Value: farmNumber + 1})
	store.Set(types.GlobalFarmNumber, bz)

	return farmNumber
}
