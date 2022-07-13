package simulation

import (
	"errors"
	"time"

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
	// TODO: Make random Sim Account with filter API, remove error here
	sender := sim.RandomSimAccount()
	lock, err := RandomAccountLock(k, sim, ctx, sender.Address)
	if err != nil {
		return nil, err
	}
	return &types.MsgBeginUnlocking{
		Owner: sender.Address.String(),
		ID:    lock.ID,
	}, nil
}

func RandomAccountLock(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context, addr sdk.AccAddress) (types.PeriodLock, error) {
	locks := k.GetAccountPeriodLocks(ctx, addr)
	if len(locks) == 0 {
		return types.PeriodLock{}, errors.New("no lock found for address")
	}
	return simulation.RandSelect(sim, locks...), nil
}
