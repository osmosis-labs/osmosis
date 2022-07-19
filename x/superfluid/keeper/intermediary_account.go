package keeper

import (
	"time"

	"github.com/gogo/protobuf/proto"

	lockuptypes "github.com/osmosis-labs/osmosis/v10/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v10/x/superfluid/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

func (k Keeper) GetIntermediaryAccountsForVal(ctx sdk.Context, valAddr sdk.ValAddress) []types.SuperfluidIntermediaryAccount {
	accs := k.GetAllIntermediaryAccounts(ctx)
	valAccs := []types.SuperfluidIntermediaryAccount{}
	for _, acc := range accs {
		if acc.ValAddr != valAddr.String() { // only apply for slashed validator
			continue
		}
		valAccs = append(valAccs, acc)
	}
	return valAccs
}

func (k Keeper) GetOrCreateIntermediaryAccount(ctx sdk.Context, denom, valAddr string) (types.SuperfluidIntermediaryAccount, error) {
	accountAddr := types.GetSuperfluidIntermediaryAccountAddr(denom, valAddr)
	storeAccount := k.GetIntermediaryAccount(ctx, accountAddr)
	// if storeAccount is in state, we return it.
	if !storeAccount.Empty() {
		return storeAccount, nil
	}
	// Otherwise we create the intermediary account.
	// first step, we create the gaugeID
	gaugeID, err := k.ik.CreateGauge(ctx, true, accountAddr, sdk.Coins{}, lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		// move this synthetic denom creation to a dedicated function
		Denom:    stakingSyntheticDenom(denom, valAddr),
		Duration: k.sk.GetParams(ctx).UnbondingTime,
	}, ctx.BlockTime(), 1)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return types.SuperfluidIntermediaryAccount{}, err
	}

	intermediaryAcct := types.NewSuperfluidIntermediaryAccount(denom, valAddr, gaugeID)
	k.SetIntermediaryAccount(ctx, intermediaryAcct)

	// If the intermediary account's address doesn't already have an auth account associated with it,
	// create a new account. We use base accounts, as this is whats done for cosmwasm smart contract accounts.
	// and in the off-chance someone manages to find a bug that forces the account's creation.
	if !k.ak.HasAccount(ctx, intermediaryAcct.GetAccAddress()) {
		k.ak.SetAccount(ctx, authtypes.NewBaseAccount(intermediaryAcct.GetAccAddress(), nil, 0, 0))
	}

	return intermediaryAcct, nil
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

func (k Keeper) DeleteAllEmptyIntermediaryAccounts(ctx sdk.Context) {
	itermedairyAccounts := k.GetAllIntermediaryAccounts(ctx)
	for _, intermediaryAccount := range itermedairyAccounts {
		k.DeleteIntermediaryAccountIfNoDelegation(ctx, intermediaryAccount)
	}
}

// DeleteIntermediaryAccount deletes given intermediary account from store
// Note that intermediary account is highly related to staking and delgation, and this
// method should be used with caution.
func (k Keeper) DeleteIntermediaryAccountIfNoDelegation(ctx sdk.Context, intermedairyAcc types.SuperfluidIntermediaryAccount) {
	// check if any delegations or intermediary connections exist
	if k.IntermediaryAccountDelegationsExists(ctx, intermedairyAcc) {
		return
	}

	store := ctx.KVStore(k.storeKey)

	// store for intermediary account
	prefixStore := prefix.NewStore(store, types.KeyPrefixIntermediaryAccount)
	prefixStore.Delete(intermedairyAcc.GetAccAddress())
}

// IntermediaryAccountDelegationsExists returns true if the gien intermediary account has any delegations.
// We check this by
// - using staking keeper to check if actual delegations exist,
// - checking if there are any intermediary account connection remaining to the intermediaryAcc
// - if there's no synthetic lock with the intermediaryAccount.Denom + intermediaryAccount.ValAddr combination
func (k Keeper) IntermediaryAccountDelegationsExists(ctx sdk.Context, intermedairyAcc types.SuperfluidIntermediaryAccount) (delegations bool) {
	store := ctx.KVStore(k.storeKey)

	// we first check if the intermediary account does not have any connections
	// store for intermediary account connection
	intermediaryAccConnectionPrefixStore := prefix.NewStore(store, types.KeyPrefixLockIntermediaryAccAddr)
	iterator := intermediaryAccConnectionPrefixStore.Iterator(nil, nil)

	intermediaryConnectionExists := false

	for ; iterator.Valid(); iterator.Next() {
		if sdk.AccAddress(iterator.Value()).Equals(intermedairyAcc.GetAccAddress()) {
			intermediaryConnectionExists = true
		}
	}

	if intermediaryConnectionExists {
		return true
	}

	// now check and verify that we don't have any delegations
	_, found := k.sk.GetDelegation(ctx, intermedairyAcc.GetAccAddress(), sdk.ValAddress(intermedairyAcc.ValAddr))
	if found {
		return found
	}

	// check that the there's no synth denom with the synthdenom from the intermediary account
	locks := k.lk.GetLocksLongerThanDurationDenom(ctx, intermedairyAcc.Denom, time.Second)
	for _, lock := range locks {
		// check if there are bonded synth locks with the synthdenom
		synthLock, _ := k.lk.GetSyntheticLockup(ctx, lock.ID, stakingSyntheticDenom(intermedairyAcc.Denom, intermedairyAcc.ValAddr))
		if synthLock != nil {
			return true
		}

		// check if there are unbonding synth locks with the synthdenom
		synthLock, _ = k.lk.GetSyntheticLockup(ctx, lock.ID, unstakingSyntheticDenom(intermedairyAcc.Denom, intermedairyAcc.ValAddr))
		if synthLock != nil {
			return true
		}
	}
	return false
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

func (k Keeper) GetAllLockIdIntermediaryAccountConnections(ctx sdk.Context) []types.LockIdIntermediaryAccountConnection {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixLockIntermediaryAccAddr)

	iterator := prefixStore.Iterator(nil, nil)

	connections := []types.LockIdIntermediaryAccountConnection{}
	for ; iterator.Valid(); iterator.Next() {
		connections = append(connections, types.LockIdIntermediaryAccountConnection{
			LockId:              sdk.BigEndianToUint64(iterator.Key()),
			IntermediaryAccount: sdk.AccAddress(iterator.Value()).String(),
		})
	}
	return connections
}

// Returns Superfluid Intermediate Account and a bool if found / not found.
func (k Keeper) GetIntermediaryAccountFromLockId(ctx sdk.Context, lockId uint64) (types.SuperfluidIntermediaryAccount, bool) {
	addr := k.GetLockIdIntermediaryAccountConnection(ctx, lockId)
	if addr.Empty() {
		return types.SuperfluidIntermediaryAccount{}, false
	}
	return k.GetIntermediaryAccount(ctx, addr), true
}

func (k Keeper) DeleteLockIdIntermediaryAccountConnection(ctx sdk.Context, lockId uint64) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixLockIntermediaryAccAddr)
	prefixStore.Delete(sdk.Uint64ToBigEndian(lockId))
}
