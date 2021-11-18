package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (k Keeper) GetAllIntermediaryAccounts() []types.SuperfluidIntermediaryAccount {
	// TODO: implement
	return []types.SuperfluidIntermediaryAccount{}
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

func (k Keeper) SetSyntheticLockupOwner(acc types.SuperfluidIntermediaryAccount, synthLock lockuptypes.SyntheticLock) {
	// TODO: this might be not be useful since synthetic lockup is already supporting this
	// addr := acc.GetAddress()
	// set on superfluid storage | `{address}{lockid}{suffix}` => synthLock
}

func (k Keeper) GetOwnedSyntheticLockups(acc types.SuperfluidIntermediaryAccount) []lockuptypes.SyntheticLock {
	// TODO: this might be not be useful since synthetic lockup is already supporting this
	// addr := acc.GetAddress()
	// TODO: read from iterator for `{address}` prefix
	return []lockuptypes.SyntheticLock{}
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
