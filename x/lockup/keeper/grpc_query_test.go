package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

func (s *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) {
	s.FundAcc(addr, coins)
	_, err := s.querier.CreateLock(s.Ctx, addr, coins, duration)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) BeginUnlocking(addr sdk.AccAddress) {
	_, err := s.querier.BeginUnlockAllNotUnlockings(s.Ctx, addr)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) WithdrawAllMaturedLocks() {
	s.querier.WithdrawAllMaturedLocks(s.Ctx)
}

func (s *KeeperTestSuite) TestModuleBalance() {
	s.SetupTest()

	// initial check
	res, err := s.querier.ModuleBalance(s.Ctx, &types.ModuleBalanceRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// final check
	res, err = s.querier.ModuleBalance(s.Ctx, &types.ModuleBalanceRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, coins)
}

func (s *KeeperTestSuite) TestModuleLockedAmount() {
	// test for module locked balance check
	s.SetupTest()

	// initial check
	res, err := s.querier.ModuleLockedAmount(s.Ctx, &types.ModuleLockedAmountRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)
	s.BeginUnlocking(addr1)

	// current module locked balance check = unlockTime - 1s
	res, err = s.querier.ModuleLockedAmount(s.Ctx, &types.ModuleLockedAmountRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, coins)

	// module locked balance after 1 second = unlockTime
	now := s.Ctx.BlockTime()
	res, err = s.querier.ModuleLockedAmount(s.Ctx.WithBlockTime(now.Add(time.Second)), &types.ModuleLockedAmountRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	// module locked balance after 2 second = unlockTime + 1s
	res, err = s.querier.ModuleLockedAmount(s.Ctx.WithBlockTime(now.Add(2*time.Second)), &types.ModuleLockedAmountRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})
}

func (s *KeeperTestSuite) TestAccountUnlockableCoins() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))

	// empty address unlockable coins check
	_, err := s.querier.AccountUnlockableCoins(s.Ctx, &types.AccountUnlockableCoinsRequest{Owner: ""})
	s.Require().Error(err)

	// initial check
	res, err := s.querier.AccountUnlockableCoins(s.Ctx, &types.AccountUnlockableCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// check before start unlocking
	res, err = s.querier.AccountUnlockableCoins(s.Ctx, &types.AccountUnlockableCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	s.BeginUnlocking(addr1)

	// check = unlockTime - 1s
	res, err = s.querier.AccountUnlockableCoins(s.Ctx, &types.AccountUnlockableCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	// check after 1 second = unlockTime
	now := s.Ctx.BlockTime()
	res, err = s.querier.AccountUnlockableCoins(s.Ctx.WithBlockTime(now.Add(time.Second)), &types.AccountUnlockableCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, coins)

	// check after 2 second = unlockTime + 1s
	res, err = s.querier.AccountUnlockableCoins(s.Ctx.WithBlockTime(now.Add(2*time.Second)), &types.AccountUnlockableCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, coins)
}

func (s *KeeperTestSuite) TestAccountUnlockingCoins() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))

	// empty address unlockable coins check
	_, err := s.querier.AccountUnlockingCoins(s.Ctx, &types.AccountUnlockingCoinsRequest{Owner: ""})
	s.Require().Error(err)

	// initial check
	res, err := s.querier.AccountUnlockingCoins(s.Ctx, &types.AccountUnlockingCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// check before start unlocking
	res, err = s.querier.AccountUnlockingCoins(s.Ctx, &types.AccountUnlockingCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	s.BeginUnlocking(addr1)

	// check at unlockTime - 1s
	res, err = s.querier.AccountUnlockingCoins(s.Ctx, &types.AccountUnlockingCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// check after 1 second = unlockTime
	now := s.Ctx.BlockTime()
	res, err = s.querier.AccountUnlockingCoins(s.Ctx.WithBlockTime(now.Add(time.Second)), &types.AccountUnlockingCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	// check after 2 second = unlockTime + 1s
	res, err = s.querier.AccountUnlockingCoins(s.Ctx.WithBlockTime(now.Add(2*time.Second)), &types.AccountUnlockingCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})
}

func (s *KeeperTestSuite) TestAccountLockedCoins() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))

	// empty address locked coins check
	_, err := s.querier.AccountLockedCoins(s.Ctx, &types.AccountLockedCoinsRequest{})
	s.Require().Error(err)

	// initial check
	res, err := s.querier.AccountLockedCoins(s.Ctx, &types.AccountLockedCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)
	s.BeginUnlocking(addr1)

	// check = unlockTime - 1s
	res, err = s.querier.AccountLockedCoins(s.Ctx, &types.AccountLockedCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(coins, res.Coins)

	// check after 1 second = unlockTime
	now := s.Ctx.BlockTime()
	res, err = s.querier.AccountLockedCoins(s.Ctx.WithBlockTime(now.Add(time.Second)), &types.AccountLockedCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	// check after 2 second = unlockTime + 1s
	res, err = s.querier.AccountLockedCoins(s.Ctx.WithBlockTime(now.Add(2*time.Second)), &types.AccountLockedCoinsRequest{Owner: addr1.String()})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})
}

func (s *KeeperTestSuite) TestAccountLockedPastTime() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	now := s.Ctx.BlockTime()

	// empty address locks check
	_, err := s.querier.AccountLockedPastTime(s.Ctx, &types.AccountLockedPastTimeRequest{Owner: "", Timestamp: now})
	s.Require().Error(err)

	// initial check
	res, err := s.querier.AccountLockedPastTime(s.Ctx, &types.AccountLockedPastTimeRequest{Owner: addr1.String(), Timestamp: now})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)
	s.BeginUnlocking(addr1)

	// check = unlockTime - 1s
	res, err = s.querier.AccountLockedPastTime(s.Ctx, &types.AccountLockedPastTimeRequest{Owner: addr1.String(), Timestamp: now})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 1)

	// check after 1 second = unlockTime
	res, err = s.querier.AccountLockedPastTime(s.Ctx, &types.AccountLockedPastTimeRequest{Owner: addr1.String(), Timestamp: now.Add(time.Second)})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// check after 2 second = unlockTime + 1s
	res, err = s.querier.AccountLockedPastTime(s.Ctx, &types.AccountLockedPastTimeRequest{Owner: addr1.String(), Timestamp: now.Add(2 * time.Second)})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)
}

func (s *KeeperTestSuite) TestAccountLockedPastTimeNotUnlockingOnly() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	now := s.Ctx.BlockTime()

	// empty address locks check
	_, err := s.querier.AccountLockedPastTimeNotUnlockingOnly(s.Ctx, &types.AccountLockedPastTimeNotUnlockingOnlyRequest{Owner: "", Timestamp: now})
	s.Require().Error(err)

	// initial check
	res, err := s.querier.AccountLockedPastTimeNotUnlockingOnly(s.Ctx, &types.AccountLockedPastTimeNotUnlockingOnlyRequest{Owner: addr1.String(), Timestamp: now})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// check when not start unlocking
	res, err = s.querier.AccountLockedPastTimeNotUnlockingOnly(s.Ctx, &types.AccountLockedPastTimeNotUnlockingOnlyRequest{Owner: addr1.String(), Timestamp: now})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 1)

	// begin unlocking
	s.BeginUnlocking(addr1)

	// check after start unlocking
	res, err = s.querier.AccountLockedPastTimeNotUnlockingOnly(s.Ctx, &types.AccountLockedPastTimeNotUnlockingOnlyRequest{Owner: addr1.String(), Timestamp: now})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)
}

func (s *KeeperTestSuite) TestAccountUnlockedBeforeTime() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	now := s.Ctx.BlockTime()

	// empty address unlockables check
	_, err := s.querier.AccountUnlockedBeforeTime(s.Ctx, &types.AccountUnlockedBeforeTimeRequest{Owner: "", Timestamp: now})
	s.Require().Error(err)

	// initial check
	res, err := s.querier.AccountUnlockedBeforeTime(s.Ctx, &types.AccountUnlockedBeforeTimeRequest{Owner: addr1.String(), Timestamp: now})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)
	s.BeginUnlocking(addr1)

	// check = unlockTime - 1s
	res, err = s.querier.AccountUnlockedBeforeTime(s.Ctx, &types.AccountUnlockedBeforeTimeRequest{Owner: addr1.String(), Timestamp: now})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// check after 1 second = unlockTime
	res, err = s.querier.AccountUnlockedBeforeTime(s.Ctx, &types.AccountUnlockedBeforeTimeRequest{Owner: addr1.String(), Timestamp: now.Add(time.Second)})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 1)

	// check after 2 second = unlockTime + 1s
	res, err = s.querier.AccountUnlockedBeforeTime(s.Ctx, &types.AccountUnlockedBeforeTimeRequest{Owner: addr1.String(), Timestamp: now.Add(2 * time.Second)})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 1)
}

func (s *KeeperTestSuite) TestAccountLockedPastTimeDenom() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	now := s.Ctx.BlockTime()

	// empty address locks by denom check
	_, err := s.querier.AccountLockedPastTimeDenom(s.Ctx, &types.AccountLockedPastTimeDenomRequest{Owner: "", Denom: "stake", Timestamp: now})
	s.Require().Error(err)

	// initial check
	res, err := s.querier.AccountLockedPastTimeDenom(s.Ctx, &types.AccountLockedPastTimeDenomRequest{Owner: addr1.String(), Denom: "stake", Timestamp: now})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)
	s.BeginUnlocking(addr1)

	// check = unlockTime - 1s
	res, err = s.querier.AccountLockedPastTimeDenom(s.Ctx, &types.AccountLockedPastTimeDenomRequest{Owner: addr1.String(), Denom: "stake", Timestamp: now})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 1)

	// account locks by not available denom
	res, err = s.querier.AccountLockedPastTimeDenom(s.Ctx, &types.AccountLockedPastTimeDenomRequest{Owner: addr1.String(), Denom: "stake2", Timestamp: now})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// account locks by denom after 1 second = unlockTime
	res, err = s.querier.AccountLockedPastTimeDenom(s.Ctx, &types.AccountLockedPastTimeDenomRequest{Owner: addr1.String(), Denom: "stake", Timestamp: now.Add(time.Second)})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// account locks by denom after 2 second = unlockTime + 1s
	res, err = s.querier.AccountLockedPastTimeDenom(s.Ctx, &types.AccountLockedPastTimeDenomRequest{Owner: addr1.String(), Denom: "stake", Timestamp: now.Add(2 * time.Second)})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// try querying with prefix coins like "stak" for potential attack
	res, err = s.querier.AccountLockedPastTimeDenom(s.Ctx, &types.AccountLockedPastTimeDenomRequest{Owner: addr1.String(), Denom: "stak", Timestamp: now})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)
}

func (s *KeeperTestSuite) TestLockedByID() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))

	// lock by not available id check
	_, err := s.querier.LockedByID(s.Ctx, &types.LockedRequest{LockId: 0})
	s.Require().Error(err)

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// lock by available available id check
	res, err := s.querier.LockedByID(s.Ctx, &types.LockedRequest{LockId: 1})
	s.Require().NoError(err)
	s.Require().Equal(res.Lock.ID, uint64(1))
	s.Require().Equal(res.Lock.Owner, addr1.String())
	s.Require().Equal(res.Lock.Coins, coins)
	s.Require().Equal(res.Lock.Duration, time.Second)
	s.Require().Equal(res.Lock.EndTime, time.Time{})
	s.Require().Equal(res.Lock.IsUnlocking(), false)
}

func (s *KeeperTestSuite) TestLockRewardReceiver() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	res, err := s.querier.LockRewardReceiver(s.Ctx, &types.LockRewardReceiverRequest{LockId: 1})
	s.Require().NoError(err)
	s.Require().Equal(res.RewardReceiver, addr1.String())

	// now change lock reward receiver and then query again
	s.App.LockupKeeper.SetLockRewardReceiverAddress(s.Ctx, 1, addr1, addr2.String())
	res, err = s.querier.LockRewardReceiver(s.Ctx, &types.LockRewardReceiverRequest{LockId: 1})
	s.Require().NoError(err)
	s.Require().Equal(res.RewardReceiver, addr2.String())

	// try getting lock reward receiver for invalid lock id, this should error
	res, err = s.querier.LockRewardReceiver(s.Ctx, &types.LockRewardReceiverRequest{LockId: 10})
	s.Require().Error(err)
	s.Require().Equal(res.RewardReceiver, "")
}

func (s *KeeperTestSuite) TestNextLockID() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// lock by available available id check
	res, err := s.querier.NextLockID(s.Ctx, &types.NextLockIDRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.LockId, uint64(2))

	// create 2 more locks
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)
	res, err = s.querier.NextLockID(s.Ctx, &types.NextLockIDRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.LockId, uint64(4))
}

func (s *KeeperTestSuite) TestAccountLockedLongerDuration() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))

	// empty address locks longer than duration check
	_, err := s.querier.AccountLockedLongerDuration(s.Ctx, &types.AccountLockedLongerDurationRequest{Owner: "", Duration: time.Second})
	s.Require().Error(err)

	// initial check
	res, err := s.querier.AccountLockedLongerDuration(s.Ctx, &types.AccountLockedLongerDurationRequest{Owner: addr1.String(), Duration: time.Second})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)
	s.BeginUnlocking(addr1)

	// account locks longer than duration check, duration = 0s
	res, err = s.querier.AccountLockedLongerDuration(s.Ctx, &types.AccountLockedLongerDurationRequest{Owner: addr1.String(), Duration: 0})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 1)

	// account locks longer than duration check, duration = 1s
	res, err = s.querier.AccountLockedLongerDuration(s.Ctx, &types.AccountLockedLongerDurationRequest{Owner: addr1.String(), Duration: time.Second})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 1)

	// account locks longer than duration check, duration = 2s
	res, err = s.querier.AccountLockedLongerDuration(s.Ctx, &types.AccountLockedLongerDurationRequest{Owner: addr1.String(), Duration: 2 * time.Second})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)
}

func (s *KeeperTestSuite) TestAccountLockedLongerDurationNotUnlockingOnly() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))

	// empty address locks longer than duration check
	_, err := s.querier.AccountLockedLongerDurationNotUnlockingOnly(s.Ctx, &types.AccountLockedLongerDurationNotUnlockingOnlyRequest{Owner: "", Duration: time.Second})
	s.Require().Error(err)

	// initial check
	res, err := s.querier.AccountLockedLongerDurationNotUnlockingOnly(s.Ctx, &types.AccountLockedLongerDurationNotUnlockingOnlyRequest{Owner: addr1.String(), Duration: time.Second})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// account locks longer than duration check before start unlocking, duration = 1s
	res, err = s.querier.AccountLockedLongerDurationNotUnlockingOnly(s.Ctx, &types.AccountLockedLongerDurationNotUnlockingOnlyRequest{Owner: addr1.String(), Duration: time.Second})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 1)

	s.BeginUnlocking(addr1)

	// account locks longer than duration check after start unlocking, duration = 1s
	res, err = s.querier.AccountLockedLongerDurationNotUnlockingOnly(s.Ctx, &types.AccountLockedLongerDurationNotUnlockingOnlyRequest{Owner: addr1.String(), Duration: time.Second})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)
}

func (s *KeeperTestSuite) TestAccountLockedLongerDurationDenom() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))

	// empty address locks longer than duration by denom check
	_, err := s.querier.AccountLockedLongerDurationDenom(s.Ctx, &types.AccountLockedLongerDurationDenomRequest{Owner: "", Duration: time.Second, Denom: "stake"})
	s.Require().Error(err)

	// initial check
	res, err := s.querier.AccountLockedLongerDurationDenom(s.Ctx, &types.AccountLockedLongerDurationDenomRequest{Owner: addr1.String(), Duration: time.Second, Denom: "stake"})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)
	s.BeginUnlocking(addr1)

	// account locks longer than duration check by denom, duration = 0s
	res, err = s.querier.AccountLockedLongerDurationDenom(s.Ctx, &types.AccountLockedLongerDurationDenomRequest{Owner: addr1.String(), Duration: 0, Denom: "stake"})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 1)

	// account locks longer than duration check by denom, duration = 1s
	res, err = s.querier.AccountLockedLongerDurationDenom(s.Ctx, &types.AccountLockedLongerDurationDenomRequest{Owner: addr1.String(), Duration: time.Second, Denom: "stake"})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 1)

	// account locks longer than duration check by not available denom, duration = 1s
	res, err = s.querier.AccountLockedLongerDurationDenom(s.Ctx, &types.AccountLockedLongerDurationDenomRequest{Owner: addr1.String(), Duration: time.Second, Denom: "stake2"})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// account locks longer than duration check by denom, duration = 2s
	res, err = s.querier.AccountLockedLongerDurationDenom(s.Ctx, &types.AccountLockedLongerDurationDenomRequest{Owner: addr1.String(), Duration: 2 * time.Second, Denom: "stake"})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)

	// try querying with prefix coins like "stak" for potential attack
	res, err = s.querier.AccountLockedLongerDurationDenom(s.Ctx, &types.AccountLockedLongerDurationDenomRequest{Owner: addr1.String(), Duration: 0, Denom: "sta"})
	s.Require().NoError(err)
	s.Require().Len(res.Locks, 0)
}

func (s *KeeperTestSuite) TestLockedDenom() {
	s.SetupTest()
	addr1 := sdk.AccAddress([]byte("addr1---------------"))

	testTotalLockedDuration := func(durationStr string, expectedAmount int64) {
		duration, _ := time.ParseDuration(durationStr)
		res, err := s.querier.LockedDenom(
			s.Ctx,
			&types.LockedDenomRequest{Denom: "stake", Duration: duration})
		s.Require().NoError(err)
		s.Require().Equal(res.Amount, osmomath.NewInt(expectedAmount))
	}

	// lock coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Hour)

	// test with single lockup
	testTotalLockedDuration("0s", 10)
	testTotalLockedDuration("30m", 10)
	testTotalLockedDuration("1h", 10)
	testTotalLockedDuration("2h", 0)

	// adds different account and lockup for testing
	addr2 := sdk.AccAddress([]byte("addr2---------------"))

	coins = sdk.Coins{sdk.NewInt64Coin("stake", 20)}
	s.LockTokens(addr2, coins, time.Hour*2)

	testTotalLockedDuration("30m", 30)
	testTotalLockedDuration("1h", 30)
	testTotalLockedDuration("2h", 20)

	// test unlocking
	s.BeginUnlocking(addr2)
	testTotalLockedDuration("2h", 20)

	// finish unlocking
	duration, _ := time.ParseDuration("2h")
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(duration))
	s.WithdrawAllMaturedLocks()
	testTotalLockedDuration("2h", 0)
	testTotalLockedDuration("1h", 10)
}

func (s *KeeperTestSuite) TestParams() {
	s.SetupTest()

	// Query default params
	res, err := s.querier.Params(s.Ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal([]string(nil), res.Params.ForceUnlockAllowedAddresses)

	// Set new params & query
	s.App.LockupKeeper.SetParams(s.Ctx, types.NewParams([]string{s.TestAccs[0].String()}))
	res, err = s.querier.Params(s.Ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal([]string{s.TestAccs[0].String()}, res.Params.ForceUnlockAllowedAddresses)
}
