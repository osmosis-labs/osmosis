package simulation

import (
	"errors"
	"time"

	legacysimulationtype "github.com/cosmos/cosmos-sdk/types/simulation"

	simulation "github.com/osmosis-labs/osmosis/v7/simulation/types"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RandomMsgLockTokens(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) (*types.MsgLockTokens, error) {
	sender := sim.RandomSimAccount()
	lockCoins := sim.RandExponentialCoin(ctx, sender.Address)
	duration := simulation.RandSelect(sim, time.Minute, time.Hour, time.Hour*24)
	return &types.MsgLockTokens{
		Owner:    sender.Address.String(),
		Duration: duration,
		Coins:    sdk.Coins{lockCoins},
	}, nil
}

func RandomMsgBeginUnlockingAll(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) (*types.MsgBeginUnlockingAll, error) {
	sender := sim.RandomSimAccount()
	return &types.MsgBeginUnlockingAll{
		Owner: sender.Address.String(),
	}, nil
}

func RandomMsgBeginUnlocking(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) (*types.MsgBeginUnlocking, error) {
	sender, senderExists := sim.RandomSimAccountWithConstraint(accountHasLockConstraint(k, ctx))
	if !senderExists {
		return nil, errors.New("no addr has created a lock")
	}
	lock := randLock(k, sim, ctx, sender.Address)
	return &types.MsgBeginUnlocking{
		Owner: sender.Address.String(),
		ID:    lock.ID,
	}, nil
}

func accountHasLockConstraint(k keeper.Keeper, ctx sdk.Context) simulation.SimAccountConstraint {
	return func(acc legacysimulationtype.Account) bool {
		return len(k.GetAccountPeriodLocks(ctx, acc.Address)) != 0
	}
}

func randLock(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context, addr sdk.AccAddress) types.PeriodLock {
	locks := k.GetAccountPeriodLocks(ctx, addr)
	return simulation.RandSelect(sim, locks...)
}
