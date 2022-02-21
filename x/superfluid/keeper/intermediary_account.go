package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/gogo/protobuf/proto"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
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
	// if storeAccount isn't set in state, we get the default storeAccount.
	// if it set, then the denom is non-blank
	if storeAccount.Denom != "" {
		return storeAccount, nil
	}
	// Otherwise we create the intermediary account.
	// first step, we create the gaugeID
	gaugeID, err := k.ik.CreateGauge(ctx, true, accountAddr, sdk.Coins{}, lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		// move this synthetic denom creation to a dedicated function
		Denom:    SyntheticDenom(denom, valAddr),
		Duration: k.sk.GetParams(ctx).UnbondingTime,
	}, ctx.BlockTime(), 1)

	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return types.SuperfluidIntermediaryAccount{}, err
	}

	intermediaryAcct := types.NewSuperfluidIntermediaryAccount(denom, valAddr, gaugeID)
	k.SetIntermediaryAccount(ctx, intermediaryAcct)

	// TODO: @Dev added this hasAccount gating, think through if theres an edge case that makes it not right
	if !k.ak.HasAccount(ctx, intermediaryAcct.GetAccAddress()) {
		// TODO: Why is this a base account, not a module account?
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
