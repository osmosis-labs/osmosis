package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
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
		// TODO: at which point, gaugeId should be created? Probably when delegation start?
		// What if just put gaugeId as part of Intermediary account?
		// TODO: should the superfluid module keep the gaugeId mapping for account?
		gaugeId, err := k.ik.CreateGauge(ctx, true, addr, rewards, types.QueryCondition{}, ctx.BlockTime(), 1)
		k.ik.AddToGaugeRewards(ctx, addr, rewards, gaugeId)
	}
}
