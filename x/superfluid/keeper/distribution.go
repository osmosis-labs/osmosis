package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmoutils"
)

func (k Keeper) MoveSuperfluidDelegationRewardToGauges(ctx sdk.Context) {
	accs := k.GetAllIntermediaryAccounts(ctx)
	for _, acc := range accs {
		addr := acc.GetAccAddress()
		valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
		if err != nil {
			panic(err)
		}

		// To avoid unexpected issues on WithdrawDelegationRewards and AddToGaugeRewards
		// we use cacheCtx and apply the changes later
		osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			_, err := k.dk.WithdrawDelegationRewards(cacheCtx, addr, valAddr)
			return err
		})

		// Send delegation rewards to gauges
		osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			// Note! We only send the bond denom (osmo), to avoid attack vectors where people
			// send many different denoms to the intermediary account, and make a resource exhaustion attack on end block.
			bondDenom := k.sk.BondDenom(cacheCtx)
			balance := k.bk.GetBalance(cacheCtx, addr, bondDenom)
			return k.ik.AddToGaugeRewards(cacheCtx, addr, sdk.Coins{balance}, acc.GaugeId)
		})
	}
}
