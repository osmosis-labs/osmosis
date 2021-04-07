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

	prevRewardRatio := k.getHistoricalEntry(ctx, farm.FarmId, farm.CurrentPeriod-1).CumulativeRewardRatio

	rewardRatio := sdk.DecCoins{}
	if farm.TotalShare.IsPositive() {
		rewardRatio = decCoins.QuoDecTruncate(farm.TotalShare.ToDec())
	}

	k.setHistoricalEntry(ctx, farm.FarmId, farm.CurrentPeriod, types.HistoricalEntry{
		CumulativeRewardRatio: prevRewardRatio.Add(rewardRatio...),
	})

	// Whenever rewards allocated to the farm, increase the currenct period.
	farm.CurrentPeriod = farm.CurrentPeriod + 1

	return k.setFarm(ctx, farm)
}
