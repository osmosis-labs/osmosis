package keeper

import (
	"context"

	"github.com/cosmos/gogoproto/proto"

	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (k Keeper) GetAllIntermediaryAccounts(context context.Context) []types.SuperfluidIntermediaryAccount {
	ctx := sdk.UnwrapSDKContext(context)
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixIntermediaryAccount)

	accounts := []types.SuperfluidIntermediaryAccount{}

	iterator := storetypes.KVStorePrefixIterator(prefixStore, nil)
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
	stakingParams, err := k.sk.GetParams(ctx)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return types.SuperfluidIntermediaryAccount{}, err
	}
	gaugeID, err := k.ik.CreateGauge(ctx, true, accountAddr, sdk.Coins{}, lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		// move this synthetic denom creation to a dedicated function
		Denom:    stakingSyntheticDenom(denom, valAddr),
		Duration: stakingParams.UnbondingTime,
	}, ctx.BlockTime(), 1, 0)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return types.SuperfluidIntermediaryAccount{}, err
	}

	intermediaryAcct := types.NewSuperfluidIntermediaryAccount(denom, valAddr, gaugeID)
	k.SetIntermediaryAccount(ctx, intermediaryAcct)

	// If the intermediary account's address doesn't already have an auth account associated with it,
	// create a new account. We use base accounts, as this is what's done for cosmwasm smart contract accounts.
	// and in the off-chance someone manages to find a bug that forces the account's creation.
	if !k.ak.HasAccount(ctx, intermediaryAcct.GetAccAddress()) {
		k.ak.SetAccount(ctx, authtypes.NewBaseAccount(intermediaryAcct.GetAccAddress(), nil, k.ak.NextAccountNumber(ctx), 0))
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

func (k Keeper) GetAllLockIdIntermediaryAccountConnections(ctx sdk.Context) []types.LockIdIntermediaryAccountConnection {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixLockIntermediaryAccAddr)

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

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
