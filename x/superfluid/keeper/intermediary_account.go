package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (k Keeper) GetAllIntermediaryAccounts(ctx sdk.Context) []types.SuperfluidIntermediaryAccount {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixIntermediaryAccount)

	accounts := []types.SuperfluidIntermediaryAccount{}

	iterator := sdk.KVStorePrefixIterator(prefixStore, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		account := types.SuperfluidIntermediaryAccount{}
		err := proto.Unmarshal(iterator.Value(), &account)
		if err != nil {
			panic(err)
		}

		accounts = append(accounts, account)
	}
	return accounts
}

func (k Keeper) GetIntermediaryAccount(ctx sdk.Context, address sdk.AccAddress) types.SuperfluidIntermediaryAccount {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixIntermediaryAccount)

	acc := types.SuperfluidIntermediaryAccount{}
	if address == nil {
		return acc
	}

	bz := prefixStore.Get(address)
	if bz == nil {
		return acc
	}
	err := proto.Unmarshal(bz, &acc)
	if err != nil {
		panic(err)
	}
	return acc
}

func (k Keeper) SetIntermediaryAccount(ctx sdk.Context, acc types.SuperfluidIntermediaryAccount) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixIntermediaryAccount)

	bz, err := proto.Marshal(&acc)
	if err != nil {
		panic(err)
	}
	prefixStore.Set(acc.GetAccAddress(), bz)
}

func (k Keeper) DeleteIntermediaryAccount(ctx sdk.Context, address sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixIntermediaryAccount)
	prefixStore.Delete(address)
}

func (k Keeper) SetLockIdIntermediaryAccountConnection(ctx sdk.Context, lockId uint64, acc types.SuperfluidIntermediaryAccount) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixLockIntermediaryAccAddr)

	prefixStore.Set(sdk.Uint64ToBigEndian(lockId), acc.GetAccAddress())
}

func (k Keeper) GetLockIdIntermediaryAccountConnection(ctx sdk.Context, lockId uint64) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixLockIntermediaryAccAddr)

	return prefixStore.Get(sdk.Uint64ToBigEndian(lockId))
}

func (k Keeper) DeleteLockIdIntermediaryAccountConnection(ctx sdk.Context, lockId uint64) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixLockIntermediaryAccAddr)
	prefixStore.Delete(sdk.Uint64ToBigEndian(lockId))
}
