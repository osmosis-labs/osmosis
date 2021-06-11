package keeper

import (
	"math/rand"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/lockup/types"
)

type BehaviourAccount struct {
	Address sdk.AccAddress

	Locked []types.PeriodLock
	Unlocking []types.PeriodLock
	TotalLocked sdk.Coins
}

func NewBehaviourAccount(addr sdk.AccAddress) *BehaviourAccount {
	acc := &BehaviourAccount {
		Address: addr,
	}
	return acc
}

func (acc *BehaviourAccount) LockToken(keeper Keeper, coins sdk.Coins, duration time.Duration) func (sdk.Context, suite.Suite) {
	return func(ctx sdk.Context, suite suite.Suite) {

	beforeModuleCoins := keeper.GetModuleLockedCoins(ctx)

	// call LockTokens
	lock, err := keeper.LockTokens(ctx, acc.Address, coins, duration)
	suite.Require().NoError(err)

	// update behaviour state
	acc.Locked = append(acc.Locked, lock)
	acc.TotalLocked = acc.TotalLocked.Add(lock.Coins...)

	// assertion: GetLockByID() should return the lock
	clock, err := keeper.GetLockByID(ctx, lock.ID)
	suite.Require().NoError(err)
	suite.Require().True(lock.Equal(*clock))

	// assertion: GetModuleLockedCoins() should be increased for the amount of the
	// coins
	afterModuleCoins := keeper.GetModuleLockedCoins(ctx)
	suite.Require().True(beforeModuleCoins.Add(lock.Coins...).IsEqual(afterModuleCoins))

	// assertion: GetAccountLockedCoins() should be equal with the accountTotalLocked
	suite.Require().True(acc.TotalLocked.IsEqual(keeper.GetAccountLockedCoins(ctx, acc.Address)))

	// test all lockRefKeys are set correctly and getter works

	// assertion: GetAccountPeriodLocks() should include the lock
	suite.Require().True(LocksInclude(keeper.GetAccountPeriodLocks(ctx, acc.Address), lock))

	// assertion: GetAccountLockedPastTime() should include or exclude the lock
	t := ctx.BlockTime()
	suite.Require().False(LocksInclude(keeper.GetAccountLockedPastTime(ctx, acc.Address, t), lock))
	t = ctx.BlockTime().Add(lock.Duration)
	suite.Require().True(LocksInclude(keeper.GetAccountLockedPastTime(ctx, acc.Address, t), lock))

	// assertion: GetAccountUnlockedBeforeTime() should include or exclude the lock
	t = ctx.BlockTime()
	suite.Require().True(LocksInclude(keeper.GetAccountLockedPastTime(ctx, acc.Address, t), lock))
	t = ctx.BlockTime().Add(lock.Duration)
	suite.Require().False(LocksInclude(keeper.GetAccountLockedPastTime(ctx, acc.Address, t), lock))

	// assertion: GetAccountLockedLongerDuration() should include or exclude the
	// lock
	duration := time.Second * 0
	suite.Require().True(LocksInclude(keeper.GetAccountLockedLongerDuration(ctx, acc.Address, duration), lock))
	duration = lock.Duration
	suite.Require().False(LocksInclude(keeper.GetAccountLockedLongerDuration(ctx, acc.Address, duration), lock))
	duration = lock.Duration*2
	suite.Require().False(LocksInclude(keeper.GetAccountLockedLongerDuration(ctx, acc.Address, duration), lock))

	// assertion: GetLocksPastTimeDenom() should include or exclude the lock
	t = ctx.BlockTime()
	for _, coin := range lock.Coins {
		suite.Require().True(LocksInclude(keeper.GetLocksPastTimeDenom(ctx, coin.Denom, t), lock))
	}
	t = ctx.BlockTime().Add(lock.Duration)
	for _, coin := range lock.Coins {
		suite.Require().False(LocksInclude(keeper.GetLocksPastTimeDenom(ctx, coin.Denom, t), lock))
	}

	// assertion: GetLocksLongerThanDurationDenom() should include or exclude the
	// lock
	duration = time.Second * 0i
	for _, coin := range lock.Coins {
		suite.Require().True(LocksInclude(keeper.GetLocksLongerThanDurationDenom(ctx, coin.Denom, duration), lock))
	}
	duration = lock.Duration
	for _, coin := range lock.Coins {
		suite.Require().False(LocksInclude(keeper.GetLocksLongerThanDurationDenom(ctx, coin.Denom, duration), lock))
	}
	duration = lock.Duration*2
	for _, coin := range lock.Coins {
		suite.Require().False(LocksInclude(keeper.GetLocksLongerThanDurationDenom(ctx, coin.Denom, duration), lock))
	}

	// assertion: UnlockPeriodLockByID() should return error
	_, err = keeper.UnlockPeriodLockByID(ctx, lock.ID)
	suite.Require().Error(err)
}
}

func LocksInclude(locks []types.PeriodLock, lock types.PeriodLock) bool {
	for _, canonicalLock := range locks {
		if canonicalLock.Equal(lock) {
			return true
		}
	}
	return false
}

func LocksSum(locks []types.PeriodLock) (res sdk.Coins) {
	for _, lock := range locks {
		res = res.Add(lock.Coins...)
	}
	res.Sort()
	return
}

func (acc *BehaviourAccount) BeginUnlocking(keeper Keeper, unlockingStrategy string) func(sdk.Context, suite.Suite) {
	return func(ctx sdk.Context, suite suite.Suite) {
	if len(acc.Locked) == 0 {
		return
	}

	// call either BeginUnlockPeriodLockByID or BeginUnlockAllNotUnlockings
	var fn func()
	var target []types.PeriodLock
	switch unlockingStrategy {
	case "front":
		target = acc.Locked[:1]
		acc.Locked = acc.Locked[1:]
		fn = func() { keeper.BeginUnlockPeriodLockByID(ctx, target[0].ID) }
	case "all":
		target = acc.Locked
		acc.Locked = []types.PeriodLock{}
		fn = func() { keeper.BeginUnlockAllNotUnlockings(ctx, acc.Address) }
	default:
		panic("unknown strategy")
	}

	fn()

	// update behaviour state
	acc.Unlocking = append(acc.Unlocking, target...)

	for _, lock := range target {
		// assertion: GetLockByID() should return the lock
		clock, err := keeper.GetLockByID(ctx, lock.ID)
		suite.Require().NoError(err)
		suite.Require().Equal(lock, clock)

		// assertion: GetAccountLockedCoins() should be equal with the accountTotalLocked
		suite.Require().True(acc.TotalLocked.IsEqual(keeper.GetAccountLockedCoins(ctx, acc.Address)))

		// test all lockRefKeys are set correctly and getter works

		// assertion: GetAccountPeriodLocks() should include the lock
		suite.Require().True(LocksInclude(keeper.GetAccountPeriodLocks(ctx, acc.Address), lock))

		// assertion: GetAccountLockedPastTime() should include or exclude the lock
		t := ctx.BlockTime()
		suite.Require().False(LocksInclude(keeper.GetAccountLockedPastTime(ctx, acc.Address, t), lock))
		t = lock.EndTime
		suite.Require().True(LocksInclude(keeper.GetAccountLockedPastTime(ctx, acc.Address, t), lock))

		// assertion: GetAccountUnlockedBeforeTime() should include or exclude the lock
		t = ctx.BlockTime()
		suite.Require().True(LocksInclude(keeper.GetAccountLockedPastTime(ctx, acc.Address, t), lock))
		t = lock.EndTime
		suite.Require().False(LocksInclude(keeper.GetAccountLockedPastTime(ctx, acc.Address, t), lock))

		// assertion: GetLocksPastTimeDenom() should include or exclude the lock
		t = ctx.BlockTime()
		for _, coin := range lock.Coins {
			suite.Require().True(LocksInclude(keeper.GetLocksPastTimeDenom(ctx, coin.Denom, t), lock))
		}
		t = lock.EndTime
		for _, coin := range lock.Coins {
			suite.Require().False(LocksInclude(keeper.GetLocksPastTimeDenom(ctx, coin.Denom, t), lock))
		}

		// assertion: GetAccountUnlockingCoins() should be equal to the sum of
		// accountUnlocking coins
		suite.Require().True(LocksSum(acc.Unlocking).IsEqual(keeper.GetAccountUnlockingCoins(ctx, acc.Address)))
	}
}
}

func (acc *BehaviourAccount) Unlock(keeper Keeper, unlockingStrategy string) func(sdk.Context, suite.Suite) {
	return func(ctx sdk.Context, suite suite.Suite) {
	unlockable := keeper.GetAccountUnlockedBeforeTime(ctx, acc.Address, ctx.BlockTime())
	if len(unlockable) == 0 {
		return
	}

	// call either UnlockPeriodLockByID or UnlockTokens
	var fn func()
	var target []types.PeriodLock
	switch unlockingStrategy {
	case "front":
		target = unlockable[:1]
		acc.Unlocking = deletePeriodLock(acc.Unlocking, target)
		fn = func() { keeper.UnlockPeriodLockByID(ctx, target[0].ID) }
	case "all":
		target = unlockable
		acc.Unlocking = deletePeriodLock(acc.Unlocking, target)
		fn = func() { keeper.UnlockAllUnlockableCoins(ctx, acc.Address) }
	default:
		panic("unknown strategy")
	}

	fn()

	// update behaviour state
	acc.TotalLocked = acc.TotalLocked.Sub(keeper.getCoinsFromLocks(target))

	for _, lock := range target {
		// assertion: GetLockByID() should not return the lock
		_, err := keeper.GetLockByID(ctx, lock.ID)
		suite.Require().Error(err)

		// assertion: GetAccountLockedCoins() should be equal with the accountTotalLocked
		suite.Require().True(acc.TotalLocked.IsEqual(keeper.GetAccountLockedCoins(ctx, acc.Address)))

		// test all lockRefKeys are set correctly and getter works

		// assertion: GetAccountPeriodLocks() should not include the lock
		suite.Require().False(LocksInclude(keeper.GetAccountPeriodLocks(ctx, acc.Address), lock))

		// assertion: GetAccountUnlockingCoins() should be equal to the sum of
		// accountUnlocking coins
		suite.Require().True(LocksSum(acc.Unlocking).IsEqual(keeper.GetAccountUnlockingCoins(ctx, acc.Address)))
	}
}
}

func deletePeriodLock(locks []types.PeriodLock, targets []types.PeriodLock) (res []types.PeriodLock) {
	for _, lock := range locks {
		res = append(res, lock)
	}

	for _, lock := range targets {
		for i := range res {
			if res[i].ID == lock.ID {
				res = append(res[:i], res[i+1:]...)
				break
			}
		}
	}
	return
}

func (acc *BehaviourAccount) GenerateBehaviourLockToken(ctx sdk.Context, k Keeper, bk types.BankKeeper, durationLimit time.Duration) func(sdk.Context, suite.Suite) {
	var coins sdk.Coins
	balances := bk.GetAllBalances(ctx, acc.Address)
	for _, balance := range balances {
		if rand.Int()%2==0 {
			continue
		}
		coin := sdk.Coin {
			Denom: balance.Denom,
			Amount: balance.Amount.QuoRaw(10),
		}
		coins = coins.Add(coin)
	}


	duration := time.Duration(rand.Int63n(durationLimit.Milliseconds()))*time.Millisecond

	return acc.LockToken(k, coins, duration)
}

func (acc *BehaviourAccount) GenerateBehaviourBeginUnlocking(k Keeper) func(sdk.Context, suite.Suite) {
	strategies := []string{"front", "all"}
	return acc.BeginUnlocking(k, strategies[rand.Intn(len(strategies))])
}

func (acc *BehaviourAccount) GenerateBehaviourUnlock(k Keeper) func(sdk.Context, suite.Suite) {
	strategies := []string{"front", "all"}
	return acc.Unlock(k, strategies[rand.Intn(len(strategies))])
}
/*
func (acc *BehaviourAccount) GenerateBehaviour(suite suite.Suite, ctx sdk.Context, k Keeper, bk types.BankKeeper, blockTime time.Duration, blockLimit int64) Behaviour {
		flip := rand.Intn(3)
		switch flip {
		case 0:
			return acc.GenerateBehaviourLockToken(suite, cctx, k, blockTime*time.Duration(blockLimit/10))
		case 1:
			return acc.GenerateBehaviourBeginUnlocking(suite, cctx, k)
		case 2:
			return acc.GenerateBehaviourUnlock(suite, cctx, k)
		default:
			panic("aaa")
		}
		res = append(res, be)
		cctx = cctx.WithBlockTime(cctx.BlockTime().Add(blockTime))
	
}
*/
