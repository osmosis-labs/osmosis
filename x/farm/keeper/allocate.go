package keeper

import (
	"github.com/c-osmosis/osmosis/x/farm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AllocateAssetToFarm(ctx sdk.Context, farmId uint64, newPeriod int64, allocator sdk.AccAddress, assets sdk.Coins) error {
	err := assets.Validate()
	if err != nil {
		return err
	}

	farm, err := k.GetFarm(ctx, farmId)
	if err != nil {
		return err
	}

	if farm.CurrentPeriod > newPeriod {
		panic("Period can't be decreased")
	}

	decCoins := sdk.NewDecCoinsFromCoins(assets...)
	farm.CurrentRewards = farm.CurrentRewards.Add(decCoins...)

	if farm.CurrentPeriod < newPeriod {
		prevRewardRatio := sdk.DecCoins{}
		if farm.LastPeriod != 0 {
			prevRewardRatio = k.GetHistoricalRecord(ctx, farm.FarmId, farm.LastPeriod).CumulativeRewardRatio
		}

		rewardRatio := sdk.DecCoins{}
		if farm.TotalShare.GT(sdk.NewInt(0)) {
			rewardRatio = farm.CurrentRewards.QuoDecTruncate(farm.TotalShare.ToDec())
		}

		k.SetHistoricalRecord(ctx, farm.FarmId, farm.CurrentPeriod, types.HistoricalRecord{
			CumulativeRewardRatio: prevRewardRatio.Add(rewardRatio...),
		})

		farm.LastPeriod = farm.CurrentPeriod
		farm.CurrentPeriod = newPeriod
		farm.CurrentRewards = sdk.DecCoins{}
	}

	err = k.bk.SendCoinsFromAccountToModule(ctx, allocator, types.ModuleName, assets)
	if err != nil {
		return err
	}

	return k.setFarm(ctx, farm)
}
