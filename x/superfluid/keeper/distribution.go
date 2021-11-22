package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) moveDelegationRewardToGauges(ctx sdk.Context) {
	accs := k.GetAllIntermediaryAccounts(ctx)
	for _, acc := range accs {
		addr := acc.GetAddress()
		valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
		if err != nil {
			panic(err)
		}
		// delegation, found := k.sk.GetDelegation(ctx, addr, valAddr)
		// if !found {
		// 	continue
		// }
		rewards, err := k.dk.WithdrawDelegationRewards(ctx, addr, valAddr)
		if err != nil {
			panic(err)
		}

		// Send delegation rewards to
		err = k.ik.AddToGaugeRewards(ctx, addr, rewards, acc.GaugeId)
		if err != nil {
			panic(err)
		}
	}
}
