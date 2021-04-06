package keeper

import (
	"github.com/c-osmosis/osmosis/x/farm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) DepositShareToFarm(ctx sdk.Context, farmId uint64, address sdk.AccAddress, share sdk.Int) (rewards sdk.Coins, err error) {
	farm, err := k.GetFarm(ctx, farmId)
	if err != nil {
		return nil, err
	}

	farmer := k.GetFarmer(ctx, farmId, address)
	if farmer == nil {
		farmer = k.NewFarmer(ctx, farmId, farm.CurrentPeriod-1, address, share)
	} else {
		rewards, err = k.WithdrawRewardsFromFarm(ctx, farmId, address)
		if err != nil {
			return nil, err
		}
		farmer.Share = farmer.Share.Add(share)
	}

	farm.TotalShare = farm.TotalShare.Add(share)

	err = k.setFarm(ctx, farm)
	if err != nil {
		return nil, err
	}
	k.setFarmer(ctx, farmer)

	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, rewards)
	if err != nil {
		return sdk.Coins{}, nil
	}
	return rewards, nil
}

func (k Keeper) WithdrawShareFromFarm(ctx sdk.Context, farmId uint64, address sdk.AccAddress, share sdk.Int) (rewards sdk.Coins, err error) {
	rewards, err = k.WithdrawRewardsFromFarm(ctx, farmId, address)
	if err != nil {
		return nil, err
	}

	farm, err := k.GetFarm(ctx, farmId)
	if err != nil {
		return nil, err
	}

	farmer := k.GetFarmer(ctx, farmId, address)
	if farmer == nil {
		return nil, types.ErrNoFarmerExist
	}

	if farmer.Share.LT(share) {
		return nil, types.ErrInsufficientShare
	}
	farmer.Share = farmer.Share.Sub(share)

	farm.TotalShare = farm.TotalShare.Sub(share)

	err = k.setFarm(ctx, farm)
	if err != nil {
		return nil, err
	}
	k.setFarmer(ctx, farmer)

	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, rewards)
	if err != nil {
		return sdk.Coins{}, nil
	}
	return rewards, nil
}

func (k Keeper) CalculatePendingRewards(ctx sdk.Context, farmId uint64, address sdk.AccAddress) (rewards sdk.DecCoins, err error) {
	farm, err := k.GetFarm(ctx, farmId)
	if err != nil {
		return nil, err
	}

	farmer := k.GetFarmer(ctx, farmId, address)
	if farmer == nil {
		return nil, types.ErrNoFarmerExist
	}

	if farm.CurrentPeriod-1 == 0 {
		return nil, nil
	}

	lastRewardRatio := k.GetHistoricalRecord(ctx, farm.FarmId, farm.CurrentPeriod-1).CumulativeRewardRatio
	farmerRewardRatio := sdk.DecCoins{}
	if farmer.LastWithdrawnPeriod > 0 {
		farmerRewardRatio = k.GetHistoricalRecord(ctx, farm.FarmId, farmer.LastWithdrawnPeriod).CumulativeRewardRatio
	}

	difference := lastRewardRatio.Sub(farmerRewardRatio)
	rewards = difference.MulDec(farmer.Share.ToDec())

	return rewards, nil
}

func (k Keeper) WithdrawRewardsFromFarm(ctx sdk.Context, farmId uint64, address sdk.AccAddress) (rewards sdk.Coins, err error) {
	decRewards, err := k.CalculatePendingRewards(ctx, farmId, address)
	if err != nil {
		return nil, err
	}

	rewards, _ = decRewards.TruncateDecimal()

	farm, err := k.GetFarm(ctx, farmId)
	if err != nil {
		return nil, err
	}

	farmer := k.GetFarmer(ctx, farmId, address)
	if farmer == nil {
		return nil, types.ErrNoFarmerExist
	}

	farmer.LastWithdrawnPeriod = farm.CurrentPeriod - 1
	k.setFarmer(ctx, farmer)

	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, rewards)
	if err != nil {
		return sdk.Coins{}, nil
	}
	return rewards, nil
}

func (k Keeper) NewFarmer(ctx sdk.Context, farmId uint64, currentPeriod uint64, address sdk.AccAddress, share sdk.Int) *types.Farmer {
	farmer := &types.Farmer{
		FarmId:              farmId,
		Address:             address.String(),
		Share:               share,
		LastWithdrawnPeriod: currentPeriod,
	}

	k.setFarmer(ctx, farmer)
	return farmer
}

func (k Keeper) GetFarmer(ctx sdk.Context, farmId uint64, address sdk.AccAddress) *types.Farmer {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetFarmerStoreKey(farmId, address))
	if len(bz) == 0 {
		return nil
	}

	farmer := &types.Farmer{}
	k.cdc.MustUnmarshalBinaryBare(bz, farmer)
	return farmer
}

func (k Keeper) setFarmer(ctx sdk.Context, farmer *types.Farmer) {
	store := ctx.KVStore(k.storeKey)
	accAddress, err := sdk.AccAddressFromBech32(farmer.Address)
	if err != nil {
		panic(err)
	}
	key := types.GetFarmerStoreKey(farmer.FarmId, accAddress)

	bz := k.cdc.MustMarshalBinaryBare(farmer)
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
