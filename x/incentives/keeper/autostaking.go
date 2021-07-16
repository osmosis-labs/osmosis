package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	proto "github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
)

// autostakingStoreKey returns storekey by address
func autostakingStoreKey(address string) []byte {
	return combineKeys(types.KeyPrefixAutostaking, []byte(address))
}

// IterateAutoStaking iterate through autostaking configurations
func (k Keeper) IterateAutoStaking(ctx sdk.Context, fn func(index int64, autostaking types.AutoStaking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixAutostaking)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		autostaking := types.AutoStaking{}
		err := proto.Unmarshal(iterator.Value(), &autostaking)
		if err != nil {
			panic(err)
		}
		stop := fn(i, autostaking)

		if stop {
			break
		}
		i++
	}
}

func (k Keeper) AllAutoStakings(ctx sdk.Context) []types.AutoStaking {
	autostakings := []types.AutoStaking{}
	k.IterateAutoStaking(ctx, func(index int64, autostaking types.AutoStaking) (stop bool) {
		autostakings = append(autostakings, autostaking)
		return false
	})
	return autostakings
}

// SetAutostaking set the autostaking configuration into the store
func (k Keeper) SetAutostaking(ctx sdk.Context, autostaking *types.AutoStaking) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(autostaking)
	if err != nil {
		return err
	}
	store.Set(autostakingStoreKey(autostaking.Address), bz)
	return nil
}

// GetAutostakingByAddress regurns autostaking config by address
func (k Keeper) GetAutostakingByAddress(ctx sdk.Context, address string) *types.AutoStaking {
	autostaking := types.AutoStaking{}
	store := ctx.KVStore(k.storeKey)
	autostakingKey := autostakingStoreKey(address)
	if !store.Has(autostakingKey) {
		return nil
	}
	bz := store.Get(autostakingKey)
	proto.Unmarshal(bz, &autostaking)
	return &autostaking
}

func (k Keeper) AutostakeRewards(ctx sdk.Context, owner sdk.AccAddress, distrCoins sdk.Coins) error {
	params := k.GetParams(ctx)
	bondDenom := k.sk.BondDenom(ctx)
	bondDenomAmt := distrCoins.AmountOf(bondDenom)
	// TODO: later should use user's manually configured delegation rate, for now, uses simple rate defined as param by governance
	autoDelegationAmt := bondDenomAmt.ToDec().Mul(params.MinAutostakingRate).RoundInt()
	if !autoDelegationAmt.IsPositive() {
		return nil
	}
	// auto delegate when can
	autostaking := k.GetAutostakingByAddress(ctx, owner.String())

	autostaked := false
	if autostaking != nil {
		valAddr, err := sdk.ValAddressFromBech32(autostaking.AutostakingValidator)
		if err != nil {
			return err
		}

		validator, found := k.sk.GetValidator(ctx, valAddr)
		if found {
			// NOTE: source funds are always unbonded and the param for `tokenSrc` is `stakingtypes.Unbonded`
			// Param for `subtractAccount` is `true` to substract delegating coins from owner's account.
			_, err = k.sk.Delegate(ctx, owner, autoDelegationAmt, stakingtypes.Unbonded, validator, true)
			if err != nil {
				return err
			}
			autostaked = true
		}
	}
	if !autostaked { // lock tokens forcefully - TODO: if lock tokens on every epoch, lots of locks will appear
		_, err := k.lk.LockTokens(ctx, owner, sdk.Coins{sdk.NewCoin(bondDenom, autoDelegationAmt)}, time.Hour*24*7*2)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: add few spec docs for auto-staking
