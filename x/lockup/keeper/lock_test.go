package keeper_test

import (
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestBeginUnlocking() { // test for all unlockable coins
	s.SetupTest()

	// initial check
	locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err = s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(locks, 1)
	s.Require().Equal(locks[0].EndTime, time.Time{})
	s.Require().Equal(locks[0].IsUnlocking(), false)

	// begin unlock
	locks, err = s.App.LockupKeeper.BeginUnlockAllNotUnlockings(s.Ctx, addr1)
	unlockedCoins := s.App.LockupKeeper.GetCoinsFromLocks(locks)
	s.Require().NoError(err)
	s.Require().Len(locks, 1)
	s.Require().Equal(unlockedCoins, coins)
	s.Require().Equal(locks[0].ID, uint64(1))

	// check locks
	locks, err = s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(locks, 1)
	s.Require().NotEqual(locks[0].EndTime, time.Time{})
	s.Require().NotEqual(locks[0].IsUnlocking(), false)
}

func (s *KeeperTestSuite) TestBeginForceUnlock() {
	// coins to lock
	coins := sdk.NewCoins(sdk.NewInt64Coin("stake", 10))

	testCases := []struct {
		name             string
		coins            sdk.Coins
		unlockCoins      sdk.Coins
		expectSameLockID bool
	}{
		{
			name:             "new lock id is returned if the lock was split",
			coins:            coins,
			unlockCoins:      sdk.NewCoins(sdk.NewInt64Coin("stake", 1)),
			expectSameLockID: false,
		},
		{
			name:             "same lock id is returned if the lock was not split",
			coins:            coins,
			unlockCoins:      sdk.Coins{},
			expectSameLockID: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// initial check
			locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
			s.Require().NoError(err)
			s.Require().Len(locks, 0)

			// lock coins
			addr1 := sdk.AccAddress([]byte("addr1---------------"))
			s.LockTokens(addr1, tc.coins, time.Second)

			// check locks
			locks, err = s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
			s.Require().NoError(err)
			s.Require().True(len(locks) > 0)

			for _, lock := range locks {
				s.Require().Equal(lock.EndTime, time.Time{})
				s.Require().Equal(lock.IsUnlocking(), false)

				lockID, err := s.App.LockupKeeper.BeginForceUnlock(s.Ctx, lock.ID, tc.unlockCoins)
				s.Require().NoError(err)

				if tc.expectSameLockID {
					s.Require().Equal(lockID, lock.ID)
				} else {
					s.Require().Equal(lockID, lock.ID+1)
				}

				// new or updated lock
				newLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockID)
				s.Require().NoError(err)
				s.Require().True(newLock.IsUnlocking())
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetPeriodLocks() {
	s.SetupTest()

	// initial check
	locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err = s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(locks, 1)
}

func (s *KeeperTestSuite) TestUnlock() {
	s.SetupTest()
	initialLockCoins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	concentratedShareCoins := sdk.NewCoins(sdk.NewCoin(cltypes.GetConcentratedLockupDenomFromPoolId(1), osmomath.NewInt(10)))

	testCases := []struct {
		name                          string
		unlockingCoins                sdk.Coins
		fundAcc                       sdk.Coins
		expectedBeginUnlockPass       bool
		passedTime                    time.Duration
		expectedUnlockMaturedLockPass bool
		balanceAfterUnlock            sdk.Coins
		isPartial                     bool
	}{
		{
			name:                          "normal unlocking case",
			unlockingCoins:                initialLockCoins,
			fundAcc:                       initialLockCoins,
			expectedBeginUnlockPass:       true,
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: true,
			balanceAfterUnlock:            initialLockCoins,
		},
		{
			name:                          "unlocking case with cl shares",
			unlockingCoins:                concentratedShareCoins,
			fundAcc:                       concentratedShareCoins,
			expectedBeginUnlockPass:       true,
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: true,
			balanceAfterUnlock:            sdk.Coins{}, // cl shares get burned after unlock
		},
		{
			name:                          "begin unlocking with nil as unlocking coins",
			unlockingCoins:                nil,
			fundAcc:                       initialLockCoins,
			expectedBeginUnlockPass:       true,
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: true,
			balanceAfterUnlock:            initialLockCoins,
		},
		{
			name:                          "unlocking coins exceed what's in lock",
			unlockingCoins:                sdk.Coins{sdk.NewInt64Coin("stake", 20)},
			fundAcc:                       initialLockCoins,
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: false,
			balanceAfterUnlock:            sdk.Coins{},
		},
		{
			name:                          "unlocking unknown tokens",
			unlockingCoins:                sdk.Coins{sdk.NewInt64Coin("unknown", 10)},
			fundAcc:                       initialLockCoins,
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: false,
			balanceAfterUnlock:            sdk.Coins{},
		},
		{
			name:                          "partial unlocking",
			unlockingCoins:                sdk.Coins{sdk.NewInt64Coin("stake", 5)},
			fundAcc:                       initialLockCoins,
			expectedBeginUnlockPass:       true,
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: true,
			balanceAfterUnlock:            sdk.Coins{sdk.NewInt64Coin("stake", 5)},
			isPartial:                     true,
		},
		{
			name:                          "partial unlocking cl shares",
			unlockingCoins:                sdk.Coins{sdk.NewInt64Coin(cltypes.GetConcentratedLockupDenomFromPoolId(1), 5)},
			fundAcc:                       concentratedShareCoins,
			expectedBeginUnlockPass:       true,
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: true,
			balanceAfterUnlock:            sdk.Coins{}, // cl shares get burned after unlock
			isPartial:                     true,
		},
		{
			name:                          "partial unlocking unknown tokens",
			unlockingCoins:                sdk.Coins{sdk.NewInt64Coin("unknown", 5)},
			fundAcc:                       initialLockCoins,
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: false,
			balanceAfterUnlock:            sdk.Coins{},
		},
		{
			name:                          "unlocking should not finish yet",
			unlockingCoins:                initialLockCoins,
			fundAcc:                       initialLockCoins,
			expectedBeginUnlockPass:       true,
			passedTime:                    time.Millisecond,
			expectedUnlockMaturedLockPass: false,
			balanceAfterUnlock:            sdk.Coins{},
		},
	}

	for _, tc := range testCases {
		s.SetupTest()
		lockupKeeper := s.App.LockupKeeper
		bankKeeper := s.App.BankKeeper
		ctx := s.Ctx

		addr1 := sdk.AccAddress([]byte("addr1---------------"))
		lock := types.NewPeriodLock(1, addr1, addr1.String(), time.Second, time.Time{}, tc.fundAcc)

		// lock with balance
		s.FundAcc(addr1, tc.fundAcc)
		lock, err := lockupKeeper.CreateLock(ctx, addr1, tc.fundAcc, time.Second)
		s.Require().NoError(err)

		// store in variable if we're testing partial unlocking for future use
		partialUnlocking := tc.unlockingCoins.IsAllLT(tc.fundAcc) && tc.unlockingCoins != nil

		// begin unlocking
		unlockingLock, err := lockupKeeper.BeginUnlock(ctx, lock.ID, tc.unlockingCoins)

		if tc.expectedBeginUnlockPass {
			s.Require().NoError(err)

			if tc.isPartial {
				s.Require().Equal(unlockingLock, lock.ID+1)
			}

			// check unlocking coins. When a lock is a partial lock
			// (i.e. tc.unlockingCoins is not nit and less than tc.fundAcc),
			// we only unlock the partial amount of tc.unlockingCoins
			expectedUnlockingCoins := tc.unlockingCoins
			if expectedUnlockingCoins == nil {
				expectedUnlockingCoins = tc.fundAcc
			}
			actualUnlockingCoins := s.App.LockupKeeper.GetAccountUnlockingCoins(s.Ctx, addr1)
			s.Require().Equal(len(actualUnlockingCoins), 1)
			s.Require().Equal(expectedUnlockingCoins[0].Amount, actualUnlockingCoins[0].Amount)

			lock = lockupKeeper.GetAccountPeriodLocks(ctx, addr1)[0]

			// if it is partial unlocking, get the new partial lock id
			if partialUnlocking {
				lock = lockupKeeper.GetAccountPeriodLocks(ctx, addr1)[1]
			}

			// check lock state
			s.Require().Equal(ctx.BlockTime().Add(lock.Duration), lock.EndTime)
			s.Require().Equal(true, lock.IsUnlocking())
		} else {
			s.Require().Error(err)

			// check unlocking coins, should not be unlocking any coins
			unlockingCoins := s.App.LockupKeeper.GetAccountUnlockingCoins(s.Ctx, addr1)
			s.Require().Equal(len(unlockingCoins), 0)

			lockedCoins := s.App.LockupKeeper.GetAccountLockedCoins(s.Ctx, addr1)
			s.Require().Equal(len(lockedCoins), 1)
			s.Require().Equal(tc.fundAcc[0], lockedCoins[0])
		}

		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(tc.passedTime))

		err = lockupKeeper.UnlockMaturedLock(ctx, lock.ID)
		if tc.expectedUnlockMaturedLockPass {
			s.Require().NoError(err)

			unlockings := lockupKeeper.GetAccountUnlockingCoins(ctx, addr1)
			s.Require().Equal(len(unlockings), 0)
		} else {
			s.Require().Error(err)
			// things to test if unlocking has started
			if tc.expectedBeginUnlockPass {
				// should still be unlocking if `UnlockMaturedLock` failed
				actualUnlockingCoins := s.App.LockupKeeper.GetAccountUnlockingCoins(s.Ctx, addr1)
				s.Require().Equal(len(actualUnlockingCoins), 1)

				expectedUnlockingCoins := tc.unlockingCoins
				if tc.unlockingCoins == nil {
					actualUnlockingCoins = tc.fundAcc
				}
				s.Require().Equal(expectedUnlockingCoins, actualUnlockingCoins)
			}
		}

		balance := bankKeeper.GetAllBalances(ctx, addr1)
		s.Require().Equal(tc.balanceAfterUnlock, balance)
	}
}

func (s *KeeperTestSuite) TestUnlockMaturedLockInternalLogic() {
	testCases := []struct {
		name                       string
		coinsLocked, coinsBurned   sdk.Coins
		expectedFinalCoinsSentBack sdk.Coins

		expectedError bool
	}{
		{
			name:                       "unlock lock with gamm shares",
			coinsLocked:                sdk.NewCoins(sdk.NewCoin("gamm/pool/1", osmomath.NewInt(100))),
			coinsBurned:                sdk.NewCoins(),
			expectedFinalCoinsSentBack: sdk.NewCoins(sdk.NewCoin("gamm/pool/1", osmomath.NewInt(100))),
			expectedError:              false,
		},
		{
			name:                       "unlock lock with cl shares",
			coinsLocked:                sdk.NewCoins(sdk.NewCoin(cltypes.GetConcentratedLockupDenomFromPoolId(1), osmomath.NewInt(100))),
			coinsBurned:                sdk.NewCoins(sdk.NewCoin(cltypes.GetConcentratedLockupDenomFromPoolId(1), osmomath.NewInt(100))),
			expectedFinalCoinsSentBack: sdk.NewCoins(),
			expectedError:              false,
		},
		{
			name:                       "unlock lock with gamm and cl shares (should not be possible)",
			coinsLocked:                sdk.NewCoins(sdk.NewCoin("gamm/pool/1", osmomath.NewInt(100)), sdk.NewCoin("cl/pool/1/1", osmomath.NewInt(100))),
			coinsBurned:                sdk.NewCoins(sdk.NewCoin("cl/pool/1/1", osmomath.NewInt(100))),
			expectedFinalCoinsSentBack: sdk.NewCoins(sdk.NewCoin("gamm/pool/1", osmomath.NewInt(100))),
			expectedError:              false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx := s.Ctx
			lockupKeeper := s.App.LockupKeeper
			bankKeeper := s.App.BankKeeper
			owner := s.TestAccs[0]

			// Fund the account with lp shares we intend to lock
			s.FundAcc(owner, tc.coinsLocked)

			// Note the supply of the coins being locked
			assetsSupplyAtLockStart := sdk.Coins{}
			for _, coin := range tc.coinsLocked {
				assetSupplyAtLockStart := s.App.BankKeeper.GetSupply(s.Ctx, coin.Denom)
				assetsSupplyAtLockStart = assetsSupplyAtLockStart.Add(assetSupplyAtLockStart)
			}

			// Lock the shares
			lockCreated, err := lockupKeeper.CreateLock(ctx, owner, tc.coinsLocked, time.Hour)
			s.Require().NoError(err)

			// Begin unlocking the lock
			_, err = lockupKeeper.BeginUnlock(ctx, lockCreated.ID, lockCreated.Coins)
			s.Require().NoError(err)

			// Note the balance of the lockup module before the unlock
			lockupModuleBalancePre := lockupKeeper.GetModuleBalance(ctx)

			// System under test
			err = lockupKeeper.UnlockMaturedLockInternalLogic(ctx, lockCreated)
			s.Require().NoError(err)

			// Check that the correct coins were sent back to the owner
			actualFinalCoinsSentBack := bankKeeper.GetAllBalances(ctx, owner)
			s.Require().Equal(tc.expectedFinalCoinsSentBack.String(), actualFinalCoinsSentBack.String())

			// Ensure that the lock was deleted
			_, err = lockupKeeper.GetLockByID(ctx, lockCreated.ID)
			s.Require().ErrorIs(err, types.ErrLockupNotFound)

			// Ensure that the lock refs were deleted from the unlocking queue
			allLocks, err := lockupKeeper.GetPeriodLocks(ctx)
			s.Require().NoError(err)
			s.Require().Empty(allLocks)

			// Ensure that the correct coins left the module account
			lockupModuleBalancePost := lockupKeeper.GetModuleBalance(ctx)
			coinsRemovedFromModuleAccount := lockupModuleBalancePre.Sub(lockupModuleBalancePost...)
			s.Require().Equal(tc.coinsLocked, coinsRemovedFromModuleAccount)

			// Note the supply of the coins after the lock has matured
			assetsSupplyAtLockEnd := sdk.Coins{}
			for _, coin := range tc.coinsLocked {
				assetSupplyAtLockEnd := s.App.BankKeeper.GetSupply(s.Ctx, coin.Denom)
				assetsSupplyAtLockEnd = assetsSupplyAtLockEnd.Add(assetSupplyAtLockEnd)
			}

			for _, coin := range tc.coinsLocked {
				if coin.Denom == "gamm/pool/1" {
					// The supply should be the same as before the lock matured
					s.Require().Equal(assetsSupplyAtLockStart.AmountOf(coin.Denom).String(), assetsSupplyAtLockEnd.AmountOf(coin.Denom).String())
				} else if coin.Denom == "cl/pool/1/1" {
					// The supply should be zero
					s.Require().Equal(osmomath.ZeroInt().String(), assetsSupplyAtLockEnd.AmountOf(coin.Denom).String())
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestModuleLockedCoins() {
	s.SetupTest()

	// initial check
	lockedCoins := s.App.LockupKeeper.GetModuleLockedCoins(s.Ctx)
	s.Require().Equal(lockedCoins, sdk.Coins{})

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// final check
	lockedCoins = s.App.LockupKeeper.GetModuleLockedCoins(s.Ctx)
	s.Require().Equal(lockedCoins, coins)
}

func (s *KeeperTestSuite) TestLocksPastTimeDenom() {
	s.SetupTest()

	now := time.Now()
	s.Ctx = s.Ctx.WithBlockTime(now)

	// initial check
	locks := s.App.LockupKeeper.GetLocksPastTimeDenom(s.Ctx, "stake", now)
	s.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// final check
	locks = s.App.LockupKeeper.GetLocksPastTimeDenom(s.Ctx, "stake", now)
	s.Require().Len(locks, 1)
}

func (s *KeeperTestSuite) TestLocksLongerThanDurationDenom() {
	s.SetupTest()

	// initial check
	duration := time.Second
	locks := s.App.LockupKeeper.GetLocksLongerThanDurationDenom(s.Ctx, "stake", duration)
	s.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// final check
	locks = s.App.LockupKeeper.GetLocksLongerThanDurationDenom(s.Ctx, "stake", duration)
	s.Require().Len(locks, 1)
}

func (s *KeeperTestSuite) TestCreateLock() {
	s.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

	// test locking without balance
	_, err := s.App.LockupKeeper.CreateLock(s.Ctx, addr1, coins, time.Second)
	s.Require().Error(err)

	s.FundAcc(addr1, coins)

	lock, err := s.App.LockupKeeper.CreateLock(s.Ctx, addr1, coins, time.Second)
	s.Require().NoError(err)

	// check new lock
	s.Require().Equal(coins, lock.Coins)
	s.Require().Equal(time.Second, lock.Duration)
	s.Require().Equal(time.Time{}, lock.EndTime)
	s.Require().Equal(uint64(1), lock.ID)
	s.Require().Equal("", lock.RewardReceiverAddress)

	lockID := s.App.LockupKeeper.GetLastLockID(s.Ctx)
	s.Require().Equal(uint64(1), lockID)

	// check accumulation store
	accum := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	s.Require().Equal(accum.String(), "10")

	// create new lock
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 20)}
	s.FundAcc(addr1, coins)

	lock, err = s.App.LockupKeeper.CreateLock(s.Ctx, addr1, coins, time.Second)
	s.Require().NoError(err)

	lockID = s.App.LockupKeeper.GetLastLockID(s.Ctx)
	s.Require().Equal(uint64(2), lockID)

	// check accumulation store
	accum = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	s.Require().Equal(accum.String(), "30")

	// check balance
	balance := s.App.BankKeeper.GetBalance(s.Ctx, addr1, "stake")
	s.Require().Equal(osmomath.ZeroInt(), balance.Amount)

	acc := s.App.AccountKeeper.GetModuleAccount(s.Ctx, types.ModuleName)
	balance = s.App.BankKeeper.GetBalance(s.Ctx, acc.GetAddress(), "stake")
	s.Require().Equal(osmomath.NewInt(30), balance.Amount)
}

func (s *KeeperTestSuite) TestSetLockRewardReceiverAddress() {
	testCases := []struct {
		name                  string
		isnotOwner            bool
		lockID                uint64
		useNewReceiverAddress bool
		exepctedErrorType     error
	}{
		{
			name:                  "happy case",
			isnotOwner:            false,
			lockID:                1,
			useNewReceiverAddress: true,
		},
		{
			name:                  "error: caller of the function is not the owner",
			isnotOwner:            true,
			lockID:                1,
			useNewReceiverAddress: true,
			exepctedErrorType:     types.ErrNotLockOwner,
		},
		{
			name:                  "error: lock id is invalid",
			isnotOwner:            false,
			lockID:                5,
			useNewReceiverAddress: true,
			exepctedErrorType:     errorsmod.Wrap(types.ErrLockupNotFound, fmt.Sprintf("lock with ID %d does not exist", 5)),
		},
		{
			name:                  "error: new receiver address is same as old",
			isnotOwner:            false,
			lockID:                1,
			useNewReceiverAddress: false,
			exepctedErrorType:     types.ErrRewardReceiverIsSame,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			addr1 := s.TestAccs[0]
			coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

			s.FundAcc(addr1, coins)

			lock, err := s.App.LockupKeeper.CreateLock(s.Ctx, addr1, coins, time.Second)
			s.Require().NoError(err)

			// check that the reward receiver is the lock owner by default
			s.Require().Equal(lock.RewardReceiverAddress, "")

			owner := addr1
			if tc.isnotOwner {
				owner = s.TestAccs[1]
			}

			newReceiver := addr1
			// if this field is set to true, use different account as input
			if tc.useNewReceiverAddress {
				newReceiver = s.TestAccs[1]
			}

			// System under test
			// now change the reward receiver state
			err = s.App.LockupKeeper.SetLockRewardReceiverAddress(s.Ctx, tc.lockID, owner, newReceiver.String())
			if tc.exepctedErrorType != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(tc.exepctedErrorType, err.Error())
			} else {
				s.Require().NoError(err)
				lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
				s.Require().NoError(err)
				s.Require().Equal(lock.RewardReceiverAddress, newReceiver.String())
			}

		})
	}

}

func (s *KeeperTestSuite) TestCreateLockNoSend() {
	s.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

	// test locking without balance
	lock, err := s.App.LockupKeeper.CreateLockNoSend(s.Ctx, addr1, coins, time.Second)
	s.Require().NoError(err)

	// check new lock
	s.Require().Equal(coins, lock.Coins)
	s.Require().Equal(time.Second, lock.Duration)
	s.Require().Equal(time.Time{}, lock.EndTime)
	s.Require().Equal(uint64(1), lock.ID)

	lockID := s.App.LockupKeeper.GetLastLockID(s.Ctx)
	s.Require().Equal(uint64(1), lockID)

	// check accumulation store
	accum := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	s.Require().Equal(accum.String(), "10")

	// create new lock (this time with a balance)
	originalLockBalance := int64(20)
	coins = sdk.Coins{sdk.NewInt64Coin("stake", originalLockBalance)}
	s.FundAcc(addr1, coins)

	lock, err = s.App.LockupKeeper.CreateLockNoSend(s.Ctx, addr1, coins, time.Second)
	s.Require().NoError(err)

	lockID = s.App.LockupKeeper.GetLastLockID(s.Ctx)
	s.Require().Equal(uint64(2), lockID)

	// check accumulation store
	accum = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	s.Require().Equal(accum.String(), "30")

	// check that send did not occur and balances are unchanged
	balance := s.App.BankKeeper.GetBalance(s.Ctx, addr1, "stake")
	s.Require().Equal(osmomath.NewInt(originalLockBalance).String(), balance.Amount.String())

	acc := s.App.AccountKeeper.GetModuleAccount(s.Ctx, types.ModuleName)
	balance = s.App.BankKeeper.GetBalance(s.Ctx, acc.GetAddress(), "stake")
	s.Require().Equal(osmomath.ZeroInt().String(), balance.Amount.String())
}

func (s *KeeperTestSuite) TestAddTokensToLock() {
	initialLockCoin := sdk.NewInt64Coin("stake", 10)
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))

	testCases := []struct {
		name                         string
		tokenToAdd                   sdk.Coin
		duration                     time.Duration
		lockingAddress               sdk.AccAddress
		expectAddTokensToLockSuccess bool
	}{
		{
			name:                         "normal add tokens to lock case",
			tokenToAdd:                   initialLockCoin,
			duration:                     time.Second,
			lockingAddress:               addr1,
			expectAddTokensToLockSuccess: true,
		},
		{
			name:           "not the owner of the lock",
			tokenToAdd:     initialLockCoin,
			duration:       time.Second,
			lockingAddress: addr2,
		},
		{
			name:           "lock with matching duration not existing",
			tokenToAdd:     initialLockCoin,
			duration:       time.Second * 2,
			lockingAddress: addr1,
		},
		{
			name:           "lock invalid tokens",
			tokenToAdd:     sdk.NewCoin("unknown", osmomath.NewInt(10)),
			duration:       time.Second,
			lockingAddress: addr1,
		},
		{
			name:           "token to add exceeds balance",
			tokenToAdd:     sdk.NewCoin("stake", osmomath.NewInt(20)),
			duration:       time.Second,
			lockingAddress: addr1,
		},
	}

	for _, tc := range testCases {
		s.SetupTest()
		// lock with balance
		s.FundAcc(addr1, sdk.Coins{initialLockCoin})
		originalLock, err := s.App.LockupKeeper.CreateLock(s.Ctx, addr1, sdk.Coins{initialLockCoin}, time.Second)
		s.Require().NoError(err)

		s.FundAcc(addr1, sdk.Coins{initialLockCoin})
		balanceBeforeLock := s.App.BankKeeper.GetAllBalances(s.Ctx, tc.lockingAddress)

		lockID, err := s.App.LockupKeeper.AddToExistingLock(s.Ctx, tc.lockingAddress, tc.tokenToAdd, tc.duration)

		if tc.expectAddTokensToLockSuccess {
			s.Require().NoError(err)

			// get updated lock
			lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockID)
			s.Require().NoError(err)

			// check that tokens have been added successfully to the lock
			s.Require().True(sdk.Coins{initialLockCoin.Add(tc.tokenToAdd)}.Equal(lock.Coins))

			// check balance has decreased
			balanceAfterLock := s.App.BankKeeper.GetAllBalances(s.Ctx, tc.lockingAddress)
			s.Require().True(balanceBeforeLock.Equal(balanceAfterLock.Add(tc.tokenToAdd)))

			// check accumulation store
			accum := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
				LockQueryType: types.ByDuration,
				Denom:         "stake",
				Duration:      time.Second,
			})
			s.Require().Equal(initialLockCoin.Amount.Add(tc.tokenToAdd.Amount), accum)
		} else {
			s.Require().Error(err)
			s.Require().Equal(uint64(0), lockID)

			lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, originalLock.ID)
			s.Require().NoError(err)

			// check that locked coins haven't changed
			s.Require().True(lock.Coins.Equal(sdk.Coins{initialLockCoin}))

			// check accumulation store didn't change
			accum := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
				LockQueryType: types.ByDuration,
				Denom:         "stake",
				Duration:      time.Second,
			})
			s.Require().Equal(initialLockCoin.Amount, accum)
		}
	}
}

func (s *KeeperTestSuite) TestHasLock() {
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))

	testCases := []struct {
		name            string
		tokenLocked     sdk.Coin
		durationLocked  time.Duration
		lockAddr        sdk.AccAddress
		denomToQuery    string
		durationToQuery time.Duration
		expectedHas     bool
	}{
		{
			name:            "same token, same duration",
			tokenLocked:     sdk.NewInt64Coin("stake", 10),
			durationLocked:  time.Minute,
			lockAddr:        addr1,
			denomToQuery:    "stake",
			durationToQuery: time.Minute,
			expectedHas:     true,
		},
		{
			name:            "same token, shorter duration",
			tokenLocked:     sdk.NewInt64Coin("stake", 10),
			durationLocked:  time.Minute,
			lockAddr:        addr1,
			denomToQuery:    "stake",
			durationToQuery: time.Second,
			expectedHas:     false,
		},
		{
			name:            "same token, longer duration",
			tokenLocked:     sdk.NewInt64Coin("stake", 10),
			durationLocked:  time.Minute,
			lockAddr:        addr1,
			denomToQuery:    "stake",
			durationToQuery: time.Minute * 2,
			expectedHas:     false,
		},
		{
			name:            "different token, same duration",
			tokenLocked:     sdk.NewInt64Coin("stake", 10),
			durationLocked:  time.Minute,
			lockAddr:        addr1,
			denomToQuery:    appparams.BaseCoinUnit,
			durationToQuery: time.Minute,
			expectedHas:     false,
		},
		{
			name:            "same token, same duration, different address",
			tokenLocked:     sdk.NewInt64Coin("stake", 10),
			durationLocked:  time.Minute,
			lockAddr:        addr2,
			denomToQuery:    appparams.BaseCoinUnit,
			durationToQuery: time.Minute,
			expectedHas:     false,
		},
	}
	for _, tc := range testCases {
		s.SetupTest()

		s.FundAcc(tc.lockAddr, sdk.Coins{tc.tokenLocked})
		_, err := s.App.LockupKeeper.CreateLock(s.Ctx, tc.lockAddr, sdk.Coins{tc.tokenLocked}, tc.durationLocked)
		s.Require().NoError(err)

		hasLock := s.App.LockupKeeper.HasLock(s.Ctx, addr1, tc.denomToQuery, tc.durationToQuery)
		s.Require().Equal(tc.expectedHas, hasLock)
	}
}

func (s *KeeperTestSuite) TestLock() {
	s.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

	lock := types.PeriodLock{
		ID:       1,
		Owner:    addr1.String(),
		Duration: time.Second,
		EndTime:  time.Time{},
		Coins:    coins,
	}

	// test locking without balance (should work since we don't send the underlying balance)
	err := s.App.LockupKeeper.Lock(s.Ctx, lock, coins)
	s.Require().NoError(err)

	// check accumulation store
	accum := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	s.Require().Equal(accum.String(), "10")

	s.FundAcc(addr1, coins)
	err = s.App.LockupKeeper.Lock(s.Ctx, lock, coins)
	s.Require().NoError(err)

	// check accumulation store
	accum = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	s.Require().Equal(accum.String(), "20")

	// Since lock method no longer sends the underlying coins, the account balance should be unchanged
	balance := s.App.BankKeeper.GetBalance(s.Ctx, addr1, "stake")
	s.Require().Equal(osmomath.NewInt(10).String(), balance.Amount.String())

	acc := s.App.AccountKeeper.GetModuleAccount(s.Ctx, types.ModuleName)
	balance = s.App.BankKeeper.GetBalance(s.Ctx, acc.GetAddress(), "stake")
	s.Require().Equal(osmomath.NewInt(0).String(), balance.Amount.String())
}
func (s *KeeperTestSuite) TestSplitLock() {
	defaultAmount := sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(100)), sdk.NewCoin("bar", osmomath.NewInt(200)))
	defaultHalfAmount := sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(40)), sdk.NewCoin("bar", osmomath.NewInt(110)))
	testCases := []struct {
		name                      string
		amountToSplit             sdk.Coins
		isUnlocking               bool
		isForceUnlock             bool
		useDifferentRewardAddress bool
		expectedErr               bool
	}{
		{
			"happy path: split half amount",
			defaultHalfAmount,
			false,
			false,
			false,
			false,
		},
		{
			"happy path: split full amount",
			defaultAmount,
			false,
			false,
			false,
			false,
		},
		{
			"happy path: try using reward address with different reward receiver",
			defaultAmount,
			false,
			false,
			true,
			false,
		},
		{
			"error: unlocking lock",
			defaultAmount,
			true,
			false,
			false,
			true,
		},
		{
			"error: force unlock",
			defaultAmount,
			true,
			false,
			false,
			true,
		},
	}
	for _, tc := range testCases {
		s.SetupTest()
		defaultDuration := time.Minute
		defaultEndTime := time.Time{}
		lock := types.NewPeriodLock(
			1,
			s.TestAccs[0],
			s.TestAccs[0].String(),
			defaultDuration,
			defaultEndTime,
			defaultAmount,
		)
		if tc.isUnlocking {
			lock.EndTime = s.Ctx.BlockTime()
		}
		if tc.useDifferentRewardAddress {
			lock.RewardReceiverAddress = s.TestAccs[1].String()
		}

		// manually set last lock id to 1
		s.App.LockupKeeper.SetLastLockID(s.Ctx, 1)
		// System under test
		newLock, err := s.App.LockupKeeper.SplitLock(s.Ctx, lock, tc.amountToSplit, tc.isForceUnlock)
		if tc.expectedErr {
			s.Require().Error(err)
			return
		}
		s.Require().NoError(err)

		// check if the new lock has correct states
		s.Require().Equal(newLock.ID, lock.ID+1)
		s.Require().Equal(newLock.Owner, lock.Owner)
		s.Require().Equal(newLock.Duration, lock.Duration)
		s.Require().Equal(newLock.EndTime, lock.EndTime)
		s.Require().Equal(newLock.RewardReceiverAddress, lock.RewardReceiverAddress)
		s.Require().True(newLock.Coins.Equal(tc.amountToSplit))

		// now check if the old lock has correctly updated state
		updatedOriginalLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
		s.Require().Equal(updatedOriginalLock.ID, lock.ID)
		s.Require().Equal(updatedOriginalLock.Owner, lock.Owner)
		s.Require().Equal(updatedOriginalLock.Duration, lock.Duration)
		s.Require().Equal(updatedOriginalLock.EndTime, lock.EndTime)
		s.Require().Equal(updatedOriginalLock.RewardReceiverAddress, lock.RewardReceiverAddress)
		s.Require().True(updatedOriginalLock.Coins.Equal(lock.Coins.Sub(tc.amountToSplit...)))

		// check that last lock id has incremented properly
		lastLockId := s.App.LockupKeeper.GetLastLockID(s.Ctx)
		s.Require().Equal(lastLockId, newLock.ID)
	}
}

func (s *KeeperTestSuite) AddTokensToLockForSynth() {
	s.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// lock coins on other durations
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 20)}
	s.LockTokens(addr1, coins, time.Second*2)
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 30)}
	s.LockTokens(addr1, coins, time.Second*3)

	synthlocks := []types.SyntheticLock{}
	// make three synthetic locks on each locks
	for i := uint64(1); i <= 3; i++ {
		// testing not unlocking synthlock, with same duration with underlying
		synthlock := types.SyntheticLock{
			UnderlyingLockId: i,
			SynthDenom:       fmt.Sprintf("synth1/%d", i),
			Duration:         time.Second * time.Duration(i),
		}
		err := s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, i, synthlock.SynthDenom, synthlock.Duration, false)
		s.Require().NoError(err)
		synthlocks = append(synthlocks, synthlock)

		// testing not unlocking synthlock, different duration with underlying
		synthlock.SynthDenom = fmt.Sprintf("synth2/%d", i)
		synthlock.Duration = time.Second * time.Duration(i) / 2
		err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, i, synthlock.SynthDenom, synthlock.Duration, false)
		s.Require().NoError(err)
		synthlocks = append(synthlocks, synthlock)

		// testing unlocking synthlock, different duration with underlying
		synthlock.SynthDenom = fmt.Sprintf("synth3/%d", i)
		err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, i, synthlock.SynthDenom, synthlock.Duration, true)
		s.Require().NoError(err)
		synthlocks = append(synthlocks, synthlock)
	}

	// check synthlocks are all set
	checkSynthlocks := func(amounts []uint64) {
		// by GetAllSyntheticLockups
		for i, synthlock := range s.App.LockupKeeper.GetAllSyntheticLockups(s.Ctx) {
			s.Require().Equal(synthlock, synthlocks[i])
		}
		// by GetSyntheticLockupByUnderlyingLockId
		for i := uint64(1); i <= 3; i++ {
			synthlockByLockup, found, err := s.App.LockupKeeper.GetSyntheticLockupByUnderlyingLockId(s.Ctx, i)
			s.Require().NoError(err)
			s.Require().True(found)
			s.Require().Equal(synthlockByLockup, synthlocks[(int(i)-1)*3+int(i)])

		}
		// by GetAllSyntheticLockupsByAddr
		for i, synthlock := range s.App.LockupKeeper.GetAllSyntheticLockupsByAddr(s.Ctx, addr1) {
			s.Require().Equal(synthlock, synthlocks[i])
		}
		// by GetPeriodLocksAccumulation
		for i := 1; i <= 3; i++ {
			for j := 1; j <= 3; j++ {
				// get accumulation with always-qualifying condition
				acc := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
					Denom:    fmt.Sprintf("synth%d/%d", j, i),
					Duration: time.Second / 10,
				})
				// amount retrieved should be equal with underlying lock's locked amount
				s.Require().Equal(acc.Int64(), amounts[i])

				// get accumulation with non-qualifying condition
				acc = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
					Denom:    fmt.Sprintf("synth%d/%d", j, i),
					Duration: time.Second * 100,
				})
				s.Require().Equal(acc.Int64(), 0)
			}
		}
	}

	checkSynthlocks([]uint64{10, 20, 30})

	// call AddTokensToLock
	for i := uint64(1); i <= 3; i++ {
		coins := sdk.NewInt64Coin("stake", int64(i)*10)
		s.FundAcc(addr1, sdk.Coins{coins})
		_, err := s.App.LockupKeeper.AddTokensToLockByID(s.Ctx, i, addr1, coins)
		s.Require().NoError(err)
	}

	// check if all invariants holds after calling AddTokensToLock
	checkSynthlocks([]uint64{20, 40, 60})
}

func (s *KeeperTestSuite) TestEndblockerWithdrawAllMaturedLockups() {
	s.SetupTest()
	clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(1)

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.NewCoins(sdk.NewInt64Coin("stake", 10), sdk.NewInt64Coin(clPoolDenom, 20))
	totalCoins := coins.Add(coins...).Add(coins...)

	// lock coins for 5 second, 1 seconds, and 3 seconds in that order
	times := []time.Duration{time.Second * 5, time.Second, time.Second * 3}
	sortedTimes := []time.Duration{time.Second, time.Second * 3, time.Second * 5}
	sortedTimesIndex := []uint64{2, 3, 1}
	unbondBlockTimes := make([]time.Time, len(times))

	// setup locks for 5 second, 1 second, and 3 seconds, and begin unbonding them.
	setupInitLocks := func() {
		for i := 0; i < len(times); i++ {
			unbondBlockTimes[i] = s.Ctx.BlockTime().Add(sortedTimes[i])
		}

		for i := 0; i < len(times); i++ {
			s.LockTokens(addr1, coins, times[i])
		}

		// consistency check locks
		locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
		s.Require().NoError(err)
		s.Require().Len(locks, 3)
		for i := 0; i < len(times); i++ {
			s.Require().Equal(locks[i].EndTime, time.Time{})
			s.Require().Equal(locks[i].IsUnlocking(), false)
		}

		// begin unlock
		locks, err = s.App.LockupKeeper.BeginUnlockAllNotUnlockings(s.Ctx, addr1)
		unlockedCoins := s.App.LockupKeeper.GetCoinsFromLocks(locks)
		s.Require().NoError(err)
		s.Require().Len(locks, len(times))
		s.Require().Equal(unlockedCoins, totalCoins)
		for i := 0; i < len(times); i++ {
			s.Require().Equal(sortedTimesIndex[i], locks[i].ID)
		}

		// check locks, these should now be sorted by unbonding completion time
		locks, err = s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
		s.Require().NoError(err)
		s.Require().Len(locks, 3)
		for i := 0; i < 3; i++ {
			s.Require().NotEqual(locks[i].EndTime, time.Time{})
			s.Require().Equal(locks[i].EndTime, unbondBlockTimes[i])
			s.Require().Equal(locks[i].IsUnlocking(), true)
		}
	}
	setupInitLocks()

	// try withdrawing before mature
	s.App.LockupKeeper.WithdrawAllMaturedLocks(s.Ctx)
	locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(locks, 3)

	// withdraw at 1 sec, 3 sec, and 5 sec intervals, check automatically withdrawn
	for i := 0; i < len(times); i++ {
		s.App.LockupKeeper.WithdrawAllMaturedLocks(s.Ctx.WithBlockTime(unbondBlockTimes[i]))
		locks, err = s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
		s.Require().NoError(err)
		s.Require().Len(locks, len(times)-i-1)
	}

	// We expect that only non-CL locks (i.e. locks that do not have the CL token prefix) send tokens back to the user's balance when mature. This is because CL tokens get burned after the lock matures.
	expectedCoins := sdk.NewCoins()
	for _, coin := range totalCoins {
		if !strings.HasPrefix(coin.Denom, cltypes.ConcentratedLiquidityTokenPrefix) {
			expectedCoins = expectedCoins.Add(coin)
		}
	}
	s.Require().Equal(expectedCoins, s.App.BankKeeper.GetAllBalances(s.Ctx, addr1))

	s.SetupTest()
	setupInitLocks()
	// now withdraw all locks and ensure all got withdrawn
	s.App.LockupKeeper.WithdrawAllMaturedLocks(s.Ctx.WithBlockTime(unbondBlockTimes[len(times)-1]))
	s.Require().Len(locks, 0)
}

func (s *KeeperTestSuite) TestLockAccumulationStore() {
	s.SetupTest()

	// initial check
	locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(locks, 0)

	// lock coins
	addr := sdk.AccAddress([]byte("addr1---------------"))

	// 1 * time.Second: 10 + 20
	// 2 * time.Second: 20 + 30
	// 3 * time.Second: 30
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr, coins, time.Second)
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 20)}
	s.LockTokens(addr, coins, time.Second)
	s.LockTokens(addr, coins, time.Second*2)
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 30)}
	s.LockTokens(addr, coins, time.Second*2)
	s.LockTokens(addr, coins, time.Second*3)

	// check accumulations
	acc := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: 0,
	})
	s.Require().Equal(int64(110), acc.Int64())
	acc = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 1,
	})
	s.Require().Equal(int64(110), acc.Int64())
	acc = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 2,
	})
	s.Require().Equal(int64(80), acc.Int64())
	acc = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 3,
	})
	s.Require().Equal(int64(30), acc.Int64())
	acc = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 4,
	})
	s.Require().Equal(int64(0), acc.Int64())
}

func (s *KeeperTestSuite) TestSlashTokensFromLockByID() {
	s.SetupTest()

	// initial check
	locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(locks, 0)

	// lock coins
	addr := sdk.AccAddress([]byte("addr1---------------"))

	// 1 * time.Second: 10
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr, coins, time.Second)

	// check accumulations
	acc := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second,
	})
	s.Require().Equal(int64(10), acc.Int64())

	_, err = s.App.LockupKeeper.SlashTokensFromLockByID(s.Ctx, 1, sdk.Coins{sdk.NewInt64Coin("stake", 1)})
	s.Require().NoError(err)

	acc = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second,
	})
	s.Require().Equal(int64(9), acc.Int64())

	lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, 1)
	s.Require().NoError(err)
	s.Require().Equal(lock.Coins.String(), "9stake")

	_, err = s.App.LockupKeeper.SlashTokensFromLockByID(s.Ctx, 1, sdk.Coins{sdk.NewInt64Coin("stake", 11)})
	s.Require().Error(err)

	_, err = s.App.LockupKeeper.SlashTokensFromLockByID(s.Ctx, 1, sdk.Coins{sdk.NewInt64Coin("stake1", 1)})
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestSlashTokensFromLockByIDSendUnderlyingAndBurn() {
	testCases := []struct {
		name             string
		positionCoins    sdk.Coins
		liquidityToSlash osmomath.Dec
		denomToSlash     string
		expectError      bool
	}{
		{
			name:             "happy path",
			positionCoins:    sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(1000000)), sdk.NewCoin("usdc", osmomath.NewInt(5000000000))),
			liquidityToSlash: osmomath.NewDec(10000000),
			denomToSlash:     cltypes.GetConcentratedLockupDenomFromPoolId(1),
		},
		{
			name:             "error: attempt to slash more liquidity than the lock has",
			positionCoins:    sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(1000000)), sdk.NewCoin("usdc", osmomath.NewInt(5000000000))),
			liquidityToSlash: osmomath.NewDec(100000000),
			denomToSlash:     cltypes.GetConcentratedLockupDenomFromPoolId(1),
			expectError:      true,
		},
		{
			name:             "error: attempt to slash a denom that does not exist in the lock",
			positionCoins:    sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(1000000)), sdk.NewCoin("usdc", osmomath.NewInt(5000000000))),
			denomToSlash:     cltypes.GetConcentratedLockupDenomFromPoolId(2),
			liquidityToSlash: osmomath.NewDec(10000000),
			expectError:      true,
		},
	}
	for _, tc := range testCases {
		s.SetupTest()

		// Check that there are currently no locks
		locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
		s.Require().NoError(err)
		s.Require().Len(locks, 0)

		// Fund the account we will be using
		addr := s.TestAccs[0]
		s.FundAcc(addr, tc.positionCoins)

		// Create a cl pool and a locked full range position
		clPool := s.PrepareConcentratedPool()
		clPoolId := clPool.GetId()
		positionData, concentratedLockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPoolId, addr, tc.positionCoins, time.Hour)
		s.Require().NoError(err)

		// Refetch the cl pool post full range position creation
		clPool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, clPoolId)
		s.Require().NoError(err)

		clPoolPositionDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPoolId)

		// Check the supply of the cl asset before we slash it is equal to the liquidity created
		clAssetSupplyPreSlash := s.App.BankKeeper.GetSupply(s.Ctx, clPoolPositionDenom)
		s.Require().Equal(positionData.Liquidity.TruncateInt().String(), clAssetSupplyPreSlash.Amount.String())

		// Store the cl pool balance before the slash
		clPoolBalancePreSlash := s.App.BankKeeper.GetAllBalances(s.Ctx, clPool.GetAddress())

		// Check the period locks accumulation
		acc := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
			Denom:    clPoolPositionDenom,
			Duration: time.Second,
		})
		s.Require().Equal(positionData.Liquidity.TruncateInt64(), acc.Int64())

		// The lockup module account balance before the slash should match the liquidity added to the lock
		lockupModuleBalancePreSlash := s.App.LockupKeeper.GetModuleBalance(s.Ctx)
		s.Require().Equal(sdk.NewCoins(sdk.NewCoin(clPoolPositionDenom, positionData.Liquidity.TruncateInt())), lockupModuleBalancePreSlash)

		// Slash the cl shares and the underlying assets
		// Figure out the underlying assets from the liquidity slash
		position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionData.ID)
		s.Require().NoError(err)

		concentratedPool, err := s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, position.PoolId)
		s.Require().NoError(err)

		tempPositionToCalculateUnderlyingAssets := position
		tempPositionToCalculateUnderlyingAssets.Liquidity = tc.liquidityToSlash
		asset0, asset1, err := cl.CalculateUnderlyingAssetsFromPosition(s.Ctx, tempPositionToCalculateUnderlyingAssets, concentratedPool)
		s.Require().NoError(err)

		underlyingAssetsToSlash := sdk.NewCoins(asset0, asset1)

		// The expected new liquidity is the previous liquidity minus the slashed liquidity
		expectedNewLiquidity := position.Liquidity.Sub(tc.liquidityToSlash).TruncateInt()

		// Slash the tokens from the lock
		_, err = s.App.LockupKeeper.SlashTokensFromLockByIDSendUnderlyingAndBurn(s.Ctx, concentratedLockId, sdk.Coins{sdk.NewInt64Coin(tc.denomToSlash, tc.liquidityToSlash.TruncateInt64())}, underlyingAssetsToSlash, clPool.GetAddress())
		if tc.expectError {
			s.Require().Error(err)
			continue
		} else {
			s.Require().NoError(err)
		}

		expectedNewLiquidityCoins := sdk.NewCoin(clPoolPositionDenom, expectedNewLiquidity)

		acc = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
			Denom:    tc.denomToSlash,
			Duration: time.Second,
		})
		s.Require().Equal(expectedNewLiquidityCoins.Amount.Int64(), acc.Int64())

		// The lockup module account balance after the slash should match the liquidity minus the slashed liquidity
		lockupModuleBalancePostSlash := s.App.LockupKeeper.GetModuleBalance(s.Ctx)
		s.Require().Equal(sdk.NewCoins(expectedNewLiquidityCoins), lockupModuleBalancePostSlash)

		// Check the supply of the cl asset after we slash it is equal to the liquidity created
		clAssetSupplyPostSlash := s.App.BankKeeper.GetSupply(s.Ctx, clPoolPositionDenom)
		s.Require().Equal(expectedNewLiquidityCoins.Amount.String(), clAssetSupplyPostSlash.Amount.String())

		// The lock itself should have been slashed
		lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, concentratedLockId)
		s.Require().NoError(err)
		s.Require().Equal(expectedNewLiquidityCoins.String(), lock.Coins.String())

		// The cl pool should be missing the underlying assets that were slashed
		clPoolBalancePostSlash := s.App.BankKeeper.GetAllBalances(s.Ctx, clPool.GetAddress())
		s.Require().Equal(clPoolBalancePreSlash.Sub(underlyingAssetsToSlash...), clPoolBalancePostSlash)
	}
}

func (s *KeeperTestSuite) TestEditLockup() {
	s.SetupTest()

	// initial check
	locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(locks, 0)

	// lock coins
	addr := sdk.AccAddress([]byte("addr1---------------"))

	// 1 * time.Second: 10
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr, coins, time.Second)

	// check accumulations
	acc := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second,
	})
	s.Require().Equal(int64(10), acc.Int64())

	lock, _ := s.App.LockupKeeper.GetLockByID(s.Ctx, 1)

	// duration decrease should fail
	err = s.App.LockupKeeper.ExtendLockup(s.Ctx, lock.ID, addr, time.Second/2)
	s.Require().Error(err)
	// extending lock with same duration should fail
	err = s.App.LockupKeeper.ExtendLockup(s.Ctx, lock.ID, addr, time.Second)
	s.Require().Error(err)

	// duration increase should success
	err = s.App.LockupKeeper.ExtendLockup(s.Ctx, lock.ID, addr, time.Second*2)
	s.Require().NoError(err)

	// check queries
	lock, _ = s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
	s.Require().Equal(lock.Duration, time.Second*2)
	s.Require().Equal(uint64(1), lock.ID)
	s.Require().Equal(coins, lock.Coins)

	locks = s.App.LockupKeeper.GetLocksLongerThanDurationDenom(s.Ctx, "stake", time.Second)
	s.Require().Equal(len(locks), 1)

	locks = s.App.LockupKeeper.GetLocksLongerThanDurationDenom(s.Ctx, "stake", time.Second*2)
	s.Require().Equal(len(locks), 1)

	// check accumulations
	acc = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second,
	})
	s.Require().Equal(int64(10), acc.Int64())
	acc = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 2,
	})
	s.Require().Equal(int64(10), acc.Int64())
	acc = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 3,
	})
	s.Require().Equal(int64(0), acc.Int64())
}

func (s *KeeperTestSuite) TestForceUnlock() {
	addr1 := sdk.AccAddress([]byte("addr1---------------"))

	testCases := []struct {
		name          string
		postLockSetup func()
	}{
		{
			name: "happy path",
		},
		{
			name: "superfluid staked",
			postLockSetup: func() {
				err := s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, 1, "testDenom", time.Minute, true)
				s.Require().NoError(err)
			},
		},
	}
	for _, tc := range testCases {
		// set up test and create default lock
		s.SetupTest()
		coinsToLock := sdk.NewCoins(sdk.NewCoin("stake", osmomath.NewInt(10000000)))
		s.FundAcc(addr1, sdk.NewCoins(coinsToLock...))
		lock, err := s.App.LockupKeeper.CreateLock(s.Ctx, addr1, coinsToLock, time.Minute)
		s.Require().NoError(err)

		// post lock setup
		if tc.postLockSetup != nil {
			tc.postLockSetup()
		}

		err = s.App.LockupKeeper.ForceUnlock(s.Ctx, lock)
		s.Require().NoError(err)

		// check that accumulation store has decreased
		accum := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
			LockQueryType: types.ByDuration,
			Denom:         "foo",
			Duration:      time.Minute,
		})
		s.Require().Equal(accum.String(), "0")

		// check balance of lock account to confirm
		balances := s.App.BankKeeper.GetAllBalances(s.Ctx, addr1)
		s.Require().Equal(balances, coinsToLock)

		// if it was superfluid delegated lock,
		// confirm that we don't have associated synth lock
		synthLock, found, err := s.App.LockupKeeper.GetSyntheticLockupByUnderlyingLockId(s.Ctx, lock.ID)
		s.Require().NoError(err)
		s.Require().False(found)
		s.Require().Equal((lockuptypes.SyntheticLock{}), synthLock)

		// check if lock is deleted by checking trying to get lock ID
		_, err = s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
		s.Require().Error(err)
	}
}

func (s *KeeperTestSuite) TestPartialForceUnlock() {
	addr1 := sdk.AccAddress([]byte("addr1---------------"))

	defaultDenomToLock := "stake"
	defaultAmountToLock := osmomath.NewInt(10000000)
	coinsToLock := sdk.NewCoins(sdk.NewCoin("stake", defaultAmountToLock))

	testCases := []struct {
		name               string
		coinsToForceUnlock sdk.Coins
		expectedPass       bool
	}{
		{
			name:               "unlock full amount",
			coinsToForceUnlock: coinsToLock,
			expectedPass:       true,
		},
		{
			name:               "partial unlock",
			coinsToForceUnlock: sdk.Coins{sdk.NewCoin(defaultDenomToLock, defaultAmountToLock.Quo(osmomath.NewInt(2)))},
			expectedPass:       true,
		},
		{
			name:               "unlock more than locked",
			coinsToForceUnlock: sdk.Coins{sdk.NewCoin(defaultDenomToLock, defaultAmountToLock.Add(osmomath.NewInt(2)))},
			expectedPass:       false,
		},
		{
			name:               "try unlocking with empty coins",
			coinsToForceUnlock: sdk.Coins{},
			expectedPass:       true,
		},
	}
	for _, tc := range testCases {
		// set up test and create default lock
		s.SetupTest()

		s.FundAcc(addr1, sdk.NewCoins(coinsToLock...))

		lock, err := s.App.LockupKeeper.CreateLock(s.Ctx, addr1, coinsToLock, time.Minute)
		s.Require().NoError(err)

		err = s.App.LockupKeeper.PartialForceUnlock(s.Ctx, lock, tc.coinsToForceUnlock)

		if tc.expectedPass {
			s.Require().NoError(err)

			// check balance
			balanceAfterForceUnlock := s.App.BankKeeper.GetBalance(s.Ctx, addr1, "stake")

			if tc.coinsToForceUnlock.Empty() {
				tc.coinsToForceUnlock = coinsToLock
			}
			s.Require().Equal(tc.coinsToForceUnlock, sdk.Coins{balanceAfterForceUnlock})
		} else {
			s.Require().Error(err)

			// check balance
			balanceAfterForceUnlock := s.App.BankKeeper.GetBalance(s.Ctx, addr1, "stake")
			s.Require().Equal(osmomath.NewInt(0), balanceAfterForceUnlock.Amount)
		}
	}
}
