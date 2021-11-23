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

	bz := prefixStore.Get(address)
	acc := types.SuperfluidIntermediaryAccount{}
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
	prefixStore.Set(acc.GetAddress(), bz)
}

func (k Keeper) DeleteIntermediaryAccount(ctx sdk.Context, address sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixIntermediaryAccount)
	prefixStore.Delete(address)
}

func (k Keeper) GetIntermediaryAccountOSMODelegation(acc types.SuperfluidIntermediaryAccount) sdk.Int {
	// addr := acc.GetAddress()
	// k.sk.GetDelegation(addr, acc.ValAddr) * ...
	return sdk.OneInt()
}

func (k Keeper) GetLPTokenAmount(acc types.SuperfluidIntermediaryAccount) sdk.Int {
	// addr := acc.GetAddress()
	// k.bk.GetBalance(addr)
	return sdk.OneInt()
}

func (k Keeper) GetMintedOSMOAmount(acc types.SuperfluidIntermediaryAccount) sdk.Int {
	// TODO: read from own superfluid storage
	return sdk.OneInt()
}

func (k Keeper) SetMintedOSMOAmount(acc types.SuperfluidIntermediaryAccount, amount sdk.Int) {
	// TODO: write on own superfluid storage
}

func (k Keeper) GetSlashedOSMOAmount(acc types.SuperfluidIntermediaryAccount) sdk.Int {
	// Slashed amount = Minted OSMO amount - Delegation amount
	return k.GetMintedOSMOAmount(acc).Sub(k.GetIntermediaryAccountOSMODelegation(acc))
}

func (k Keeper) SetLockIdIntermediaryAccountConnection(ctx sdk.Context, lockId uint64, acc types.SuperfluidIntermediaryAccount) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixLockIntermediaryAccAddr)

	prefixStore.Set(sdk.Uint64ToBigEndian(lockId), acc.GetAddress())
}

func (k Keeper) GetLockIdIntermediaryAccountConnection(ctx sdk.Context, lockId uint64) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixLockIntermediaryAccAddr)

	return prefixStore.Get(sdk.Uint64ToBigEndian(lockId))
}

// TODO: On TWAP change, set target OSMO mint amount to TWAP * LP
