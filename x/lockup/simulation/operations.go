package simulation

import (
	"errors"
	"time"

	legacysimulationtype "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/simulation/simtypes"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RandomMsgLockTokens(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgLockTokens, error) {
	sender, err := sim.RandomSimAccountWithBalance(ctx)
	if err != nil {
		return nil, err
	}
	lockCoin := sim.RandExponentialCoin(ctx, sender.Address)
	duration := simtypes.RandSelect(sim, time.Minute, time.Hour, time.Hour*24)
	return &types.MsgLockTokens{
		Owner:    sender.Address.String(),
		Duration: duration,
		Coins:    sdk.Coins{lockCoin},
	}, nil
}

func RandomMsgBeginUnlockingAll(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgBeginUnlockingAll, error) {
	sender := sim.RandomSimAccount()
	return &types.MsgBeginUnlockingAll{
		Owner: sender.Address.String(),
	}, nil
}

func RandomMsgBeginUnlocking(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgBeginUnlocking, error) {
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

var notUnlockingFilter = func(l types.PeriodLock) bool { return !l.IsUnlocking() }

func accountHasLockConstraint(k keeper.Keeper, ctx sdk.Context) simtypes.SimAccountConstraint {
	return func(acc legacysimulationtype.Account) bool {
		return len(osmoutils.Filter(notUnlockingFilter, k.GetAccountPeriodLocks(ctx, acc.Address))) != 0
	}
}

func randLock(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context, addr sdk.AccAddress) types.PeriodLock {
	locks := k.GetAccountPeriodLocks(ctx, addr)
	notUnlockingLocks := osmoutils.Filter(notUnlockingFilter, locks)
	return simtypes.RandSelect(sim, notUnlockingLocks...)
}
