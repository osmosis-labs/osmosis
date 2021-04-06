package keeper

import (
	"github.com/c-osmosis/osmosis/x/farm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AllocateAssetsFromAccountToFarm(ctx sdk.Context, farmId uint64, allocator sdk.AccAddress, assets sdk.Coins) error {
	err := k.allocateAssetsToFarm(ctx, farmId, assets)
	if err != nil {
		return err
	}

	return k.bk.SendCoinsFromAccountToModule(ctx, allocator, types.ModuleName, assets)
}

func (k Keeper) AllocateAssetsFromModuleToFarm(ctx sdk.Context, farmId uint64, moduleName string, assets sdk.Coins) error {
	err := k.allocateAssetsToFarm(ctx, farmId, assets)
	if err != nil {
		return err
	}

	return k.bk.SendCoinsFromModuleToModule(ctx, moduleName, types.ModuleName, assets)
}

func (k Keeper) allocateAssetsToFarm(ctx sdk.Context, farmId uint64, assets sdk.Coins) error {
	err := assets.Validate()
	if err != nil {
		return err
	}

	farm, err := k.GetFarm(ctx, farmId)
	if err != nil {
		return err
	}

	decCoins := sdk.NewDecCoinsFromCoins(assets...)

	prevRewardRatio := sdk.DecCoins{}
	if farm.CurrentPeriod-1 != 0 {
		prevRewardRatio = k.GetHistoricalRecord(ctx, farm.FarmId, farm.CurrentPeriod-1).CumulativeRewardRatio
	}

	rewardRatio := sdk.DecCoins{}
	if farm.TotalShare.GT(sdk.NewInt(0)) {
		rewardRatio = decCoins.QuoDecTruncate(farm.TotalShare.ToDec())
	}

	k.SetHistoricalRecord(ctx, farm.FarmId, farm.CurrentPeriod, types.HistoricalRecord{
		CumulativeRewardRatio: prevRewardRatio.Add(rewardRatio...),
	})

	farm.CurrentPeriod = farm.CurrentPeriod + 1

	return k.setFarm(ctx, farm)
}
