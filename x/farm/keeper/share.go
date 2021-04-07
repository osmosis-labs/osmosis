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

	farmer, err := k.GetFarmer(ctx, farmId, address)
	if err != nil {
		if err == types.ErrNoFarmerExist {
			// Farm's period begins from 1 and only increases. Therefore farm.CurrentPeriod-1 will never run into underflow.
			// Also, an empty entry with 0 preiod is registered when a farm is created, so period 0 is valid.
			farmer = k.newFarmer(ctx, farmId, farm.CurrentPeriod-1, address, share)
		} else {
			return nil, err
		}
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
		return nil, err
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

	farmer, err := k.GetFarmer(ctx, farmId, address)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return rewards, nil
}

func (k Keeper) CalculatePendingRewards(ctx sdk.Context, farmId uint64, address sdk.AccAddress) (rewards sdk.DecCoins, err error) {
	farm, err := k.GetFarm(ctx, farmId)
	if err != nil {
		return nil, err
	}

	farmer, err := k.GetFarmer(ctx, farmId, address)
	if err != nil {
		return nil, err
	}

	lastRewardRatio := k.getHistoricalEntry(ctx, farm.FarmId, farm.CurrentPeriod-1).CumulativeRewardRatio
	farmerRewardRatio := k.getHistoricalEntry(ctx, farm.FarmId, farmer.LastWithdrawnPeriod).CumulativeRewardRatio

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

	farmer, err := k.GetFarmer(ctx, farmId, address)
	if err != nil {
		return nil, err
	}

	farmer.LastWithdrawnPeriod = farm.CurrentPeriod - 1
	k.setFarmer(ctx, farmer)

	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, rewards)
	if err != nil {
		return nil, err
	}
	return rewards, nil
}
