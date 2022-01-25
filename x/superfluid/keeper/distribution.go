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
		_, err = k.dk.WithdrawDelegationRewards(cacheCtx, addr, valAddr)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
		} else {
			write()
		}

		// Send delegation rewards to gauges
		cacheCtx, write = ctx.CacheContext()
		bondDenom := k.sk.BondDenom(cacheCtx)
		balance := k.bk.GetBalance(cacheCtx, addr, bondDenom)
		err = k.ik.AddToGaugeRewards(cacheCtx, addr, sdk.Coins{balance}, acc.GaugeId)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
		} else {
			write()
		}
	}
}
