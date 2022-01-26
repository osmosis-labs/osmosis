package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) MoveSuperfluidDelegationRewardToGauges(ctx sdk.Context) {
	accs := k.GetAllIntermediaryAccounts(ctx)
	for _, acc := range accs {
		addr := acc.GetAddress()
		valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
		if err != nil {
			panic(err)
		}

		// To avoid unexpected issues on WithdrawDelegationRewards and AddToGaugeRewards
		// we use cacheCtx and apply the changes later
		cacheCtx, write := ctx.CacheContext()

		// Withdraw delegation rewards into intermediary accounts
		rewards, err := k.dk.WithdrawDelegationRewards(cacheCtx, addr, valAddr)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
			continue
		}

		// Send delegation rewards to gauges
		err = k.ik.AddToGaugeRewards(cacheCtx, addr, rewards, acc.GaugeId)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
			continue
		}
		write()
	}
}
