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
			if farm.CurrentPeriod == 0 {
				farmer = k.NewFarmer(ctx, farmId, 0, address, share)
			} else {
				farmer = k.NewFarmer(ctx, farmId, farm.CurrentPeriod-1, address, share)
			}
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
		return sdk.Coins{}, nil
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

	farmer, err := k.GetFarmer(ctx, farmId, address)
	if err != nil {
		return nil, err
	}

	farmer.LastWithdrawnPeriod = farm.CurrentPeriod - 1
	k.setFarmer(ctx, farmer)

	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, rewards)
	if err != nil {
		return sdk.Coins{}, nil
	}
	return rewards, nil
}
