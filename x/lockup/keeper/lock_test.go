package keeper_test

import (
	"fmt"
	"strings"
	"time"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestBeginUnlocking() { // test for all unlockable coins
	suite.SetupTest()

	// initial check
	locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err = suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].EndTime, time.Time{})
	suite.Require().Equal(locks[0].IsUnlocking(), false)

	// begin unlock
	locks, err = suite.App.LockupKeeper.BeginUnlockAllNotUnlockings(suite.Ctx, addr1)
	unlockedCoins := suite.App.LockupKeeper.GetCoinsFromLocks(locks)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(unlockedCoins, coins)
	suite.Require().Equal(locks[0].ID, uint64(1))

	// check locks
	locks, err = suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().NotEqual(locks[0].EndTime, time.Time{})
	suite.Require().NotEqual(locks[0].IsUnlocking(), false)
}

func (suite *KeeperTestSuite) TestBeginForceUnlock() {
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
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// initial check
			locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
			suite.Require().NoError(err)
			suite.Require().Len(locks, 0)

			// lock coins
			addr1 := sdk.AccAddress([]byte("addr1---------------"))
			suite.LockTokens(addr1, tc.coins, time.Second)

			// check locks
			locks, err = suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
			suite.Require().NoError(err)
			suite.Require().True(len(locks) > 0)

			for _, lock := range locks {
				suite.Require().Equal(lock.EndTime, time.Time{})
				suite.Require().Equal(lock.IsUnlocking(), false)

				lockID, err := suite.App.LockupKeeper.BeginForceUnlock(suite.Ctx, lock.ID, tc.unlockCoins)
				suite.Require().NoError(err)

				if tc.expectSameLockID {
					suite.Require().Equal(lockID, lock.ID)
				} else {
					suite.Require().Equal(lockID, lock.ID+1)
				}

				// new or updated lock
				newLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockID)
				suite.Require().NoError(err)
				suite.Require().True(newLock.IsUnlocking())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetPeriodLocks() {
	suite.SetupTest()

	// initial check
	locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err = suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
}

func (suite *KeeperTestSuite) TestUnlock() {
	suite.SetupTest()
	initialLockCoins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	concentratedShareCoins := sdk.NewCoins(sdk.NewCoin(cltypes.GetConcentratedLockupDenomFromPoolId(1), sdk.NewInt(10)))

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
		suite.SetupTest()
		lockupKeeper := suite.App.LockupKeeper
		bankKeeper := suite.App.BankKeeper
		ctx := suite.Ctx

		addr1 := sdk.AccAddress([]byte("addr1---------------"))
		lock := types.NewPeriodLock(1, addr1, time.Second, time.Time{}, tc.fundAcc)

		// lock with balance
		suite.FundAcc(addr1, tc.fundAcc)
		lock, err := lockupKeeper.CreateLock(ctx, addr1, tc.fundAcc, time.Second)
		suite.Require().NoError(err)

		// store in variable if we're testing partial unlocking for future use
		partialUnlocking := tc.unlockingCoins.IsAllLT(tc.fundAcc) && tc.unlockingCoins != nil

		// begin unlocking
		unlockingLock, err := lockupKeeper.BeginUnlock(ctx, lock.ID, tc.unlockingCoins)

		if tc.expectedBeginUnlockPass {
			suite.Require().NoError(err)

			if tc.isPartial {
				suite.Require().Equal(unlockingLock, lock.ID+1)
			}

			// check unlocking coins. When a lock is a partial lock
			// (i.e. tc.unlockingCoins is not nit and less than tc.fundAcc),
			// we only unlock the partial amount of tc.unlockingCoins
			expectedUnlockingCoins := tc.unlockingCoins
			if expectedUnlockingCoins == nil {
				expectedUnlockingCoins = tc.fundAcc
			}
			actualUnlockingCoins := suite.App.LockupKeeper.GetAccountUnlockingCoins(suite.Ctx, addr1)
			suite.Require().Equal(len(actualUnlockingCoins), 1)
			suite.Require().Equal(expectedUnlockingCoins[0].Amount, actualUnlockingCoins[0].Amount)

			lock = lockupKeeper.GetAccountPeriodLocks(ctx, addr1)[0]

			// if it is partial unlocking, get the new partial lock id
			if partialUnlocking {
				lock = lockupKeeper.GetAccountPeriodLocks(ctx, addr1)[1]
			}

			// check lock state
			suite.Require().Equal(ctx.BlockTime().Add(lock.Duration), lock.EndTime)
			suite.Require().Equal(true, lock.IsUnlocking())

		} else {
			suite.Require().Error(err)

			// check unlocking coins, should not be unlocking any coins
			unlockingCoins := suite.App.LockupKeeper.GetAccountUnlockingCoins(suite.Ctx, addr1)
			suite.Require().Equal(len(unlockingCoins), 0)

			lockedCoins := suite.App.LockupKeeper.GetAccountLockedCoins(suite.Ctx, addr1)
			suite.Require().Equal(len(lockedCoins), 1)
			suite.Require().Equal(tc.fundAcc[0], lockedCoins[0])
		}

		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(tc.passedTime))

		err = lockupKeeper.UnlockMaturedLock(ctx, lock.ID)
		if tc.expectedUnlockMaturedLockPass {
			suite.Require().NoError(err)

			unlockings := lockupKeeper.GetAccountUnlockingCoins(ctx, addr1)
			suite.Require().Equal(len(unlockings), 0)
		} else {
			suite.Require().Error(err)
			// things to test if unlocking has started
			if tc.expectedBeginUnlockPass {
				// should still be unlocking if `UnlockMaturedLock` failed
				actualUnlockingCoins := suite.App.LockupKeeper.GetAccountUnlockingCoins(suite.Ctx, addr1)
				suite.Require().Equal(len(actualUnlockingCoins), 1)

				expectedUnlockingCoins := tc.unlockingCoins
				if tc.unlockingCoins == nil {
					actualUnlockingCoins = tc.fundAcc
				}
				suite.Require().Equal(expectedUnlockingCoins, actualUnlockingCoins)
			}
		}

		balance := bankKeeper.GetAllBalances(ctx, addr1)
		suite.Require().Equal(tc.balanceAfterUnlock, balance)
	}
}

func (suite *KeeperTestSuite) TestUnlockMaturedLockInternalLogic() {

	testCases := []struct {
		name                       string
		coinsLocked, coinsBurned   sdk.Coins
		expectedFinalCoinsSentBack sdk.Coins

		expectedError bool
	}{
		{
			name:                       "unlock lock with gamm shares",
			coinsLocked:                sdk.NewCoins(sdk.NewCoin("gamm/pool/1", sdk.NewInt(100))),
			coinsBurned:                sdk.NewCoins(),
			expectedFinalCoinsSentBack: sdk.NewCoins(sdk.NewCoin("gamm/pool/1", sdk.NewInt(100))),
			expectedError:              false,
		},
		{
			name:                       "unlock lock with cl shares",
			coinsLocked:                sdk.NewCoins(sdk.NewCoin(cltypes.GetConcentratedLockupDenomFromPoolId(1), sdk.NewInt(100))),
			coinsBurned:                sdk.NewCoins(sdk.NewCoin(cltypes.GetConcentratedLockupDenomFromPoolId(1), sdk.NewInt(100))),
			expectedFinalCoinsSentBack: sdk.NewCoins(),
			expectedError:              false,
		},
		{
			name:                       "unlock lock with gamm and cl shares (should not be possible)",
			coinsLocked:                sdk.NewCoins(sdk.NewCoin("gamm/pool/1", sdk.NewInt(100)), sdk.NewCoin("cl/pool/1/1", sdk.NewInt(100))),
			coinsBurned:                sdk.NewCoins(sdk.NewCoin("cl/pool/1/1", sdk.NewInt(100))),
			expectedFinalCoinsSentBack: sdk.NewCoins(sdk.NewCoin("gamm/pool/1", sdk.NewInt(100))),
			expectedError:              false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			ctx := suite.Ctx
			lockupKeeper := suite.App.LockupKeeper
			bankKeeper := suite.App.BankKeeper
			owner := suite.TestAccs[0]

			// Fund the account with lp shares we intend to lock
			suite.FundAcc(owner, tc.coinsLocked)

			// Note the supply of the coins being locked
			assetsSupplyAtLockStart := sdk.Coins{}
			for _, coin := range tc.coinsLocked {
				assetSupplyAtLockStart := suite.App.BankKeeper.GetSupply(suite.Ctx, coin.Denom)
				assetsSupplyAtLockStart = assetsSupplyAtLockStart.Add(assetSupplyAtLockStart)
			}

			// Lock the shares
			lockCreated, err := lockupKeeper.CreateLock(ctx, owner, tc.coinsLocked, time.Hour)
			suite.Require().NoError(err)

			// Begin unlocking the lock
			_, err = lockupKeeper.BeginUnlock(ctx, lockCreated.ID, lockCreated.Coins)
			suite.Require().NoError(err)

			// Note the balance of the lockup module before the unlock
			lockupModuleBalancePre := lockupKeeper.GetModuleBalance(ctx)

			// System under test
			err = lockupKeeper.UnlockMaturedLockInternalLogic(ctx, lockCreated)
			suite.Require().NoError(err)

			// Check that the correct coins were sent back to the owner
			actualFinalCoinsSentBack := bankKeeper.GetAllBalances(ctx, owner)
			suite.Require().Equal(tc.expectedFinalCoinsSentBack.String(), actualFinalCoinsSentBack.String())

			// Ensure that the lock was deleted
			_, err = lockupKeeper.GetLockByID(ctx, lockCreated.ID)
			suite.Require().ErrorIs(err, types.ErrLockupNotFound)

			// Ensure that the lock refs were deleted from the unlocking queue
			allLocks, err := lockupKeeper.GetPeriodLocks(ctx)
			suite.Require().NoError(err)
			suite.Require().Empty(allLocks)

			// Ensure that the correct coins left the module account
			lockupModuleBalancePost := lockupKeeper.GetModuleBalance(ctx)
			coinsRemovedFromModuleAccount := lockupModuleBalancePre.Sub(lockupModuleBalancePost)
			suite.Require().Equal(tc.coinsLocked, coinsRemovedFromModuleAccount)

			// Note the supply of the coins after the lock has matured
			assetsSupplyAtLockEnd := sdk.Coins{}
			for _, coin := range tc.coinsLocked {
				assetSupplyAtLockEnd := suite.App.BankKeeper.GetSupply(suite.Ctx, coin.Denom)
				assetsSupplyAtLockEnd = assetsSupplyAtLockEnd.Add(assetSupplyAtLockEnd)
			}

			for _, coin := range tc.coinsLocked {
				if coin.Denom == "gamm/pool/1" {
					// The supply should be the same as before the lock matured
					suite.Require().Equal(assetsSupplyAtLockStart.AmountOf(coin.Denom).String(), assetsSupplyAtLockEnd.AmountOf(coin.Denom).String())
				} else if coin.Denom == "cl/pool/1/1" {
					// The supply should be zero
					suite.Require().Equal(sdk.ZeroInt().String(), assetsSupplyAtLockEnd.AmountOf(coin.Denom).String())
				}
			}

		})
	}
}

func (suite *KeeperTestSuite) TestModuleLockedCoins() {
	suite.SetupTest()

	// initial check
	lockedCoins := suite.App.LockupKeeper.GetModuleLockedCoins(suite.Ctx)
	suite.Require().Equal(lockedCoins, sdk.Coins(nil))

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// final check
	lockedCoins = suite.App.LockupKeeper.GetModuleLockedCoins(suite.Ctx)
	suite.Require().Equal(lockedCoins, coins)
}

func (suite *KeeperTestSuite) TestLocksPastTimeDenom() {
	suite.SetupTest()

	now := time.Now()
	suite.Ctx = suite.Ctx.WithBlockTime(now)

	// initial check
	locks := suite.App.LockupKeeper.GetLocksPastTimeDenom(suite.Ctx, "stake", now)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// final check
	locks = suite.App.LockupKeeper.GetLocksPastTimeDenom(suite.Ctx, "stake", now)
	suite.Require().Len(locks, 1)
}

func (suite *KeeperTestSuite) TestLocksLongerThanDurationDenom() {
	suite.SetupTest()

	// initial check
	duration := time.Second
	locks := suite.App.LockupKeeper.GetLocksLongerThanDurationDenom(suite.Ctx, "stake", duration)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// final check
	locks = suite.App.LockupKeeper.GetLocksLongerThanDurationDenom(suite.Ctx, "stake", duration)
	suite.Require().Len(locks, 1)
}

func (suite *KeeperTestSuite) TestCreateLock() {
	suite.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

	// test locking without balance
	_, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, addr1, coins, time.Second)
	suite.Require().Error(err)

	suite.FundAcc(addr1, coins)

	lock, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, addr1, coins, time.Second)
	suite.Require().NoError(err)

	// check new lock
	suite.Require().Equal(coins, lock.Coins)
	suite.Require().Equal(time.Second, lock.Duration)
	suite.Require().Equal(time.Time{}, lock.EndTime)
	suite.Require().Equal(uint64(1), lock.ID)

	lockID := suite.App.LockupKeeper.GetLastLockID(suite.Ctx)
	suite.Require().Equal(uint64(1), lockID)

	// check accumulation store
	accum := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")

	// create new lock
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 20)}
	suite.FundAcc(addr1, coins)

	lock, err = suite.App.LockupKeeper.CreateLock(suite.Ctx, addr1, coins, time.Second)
	suite.Require().NoError(err)

	lockID = suite.App.LockupKeeper.GetLastLockID(suite.Ctx)
	suite.Require().Equal(uint64(2), lockID)

	// check accumulation store
	accum = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "30")

	// check balance
	balance := suite.App.BankKeeper.GetBalance(suite.Ctx, addr1, "stake")
	suite.Require().Equal(sdk.ZeroInt(), balance.Amount)

	acc := suite.App.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)
	balance = suite.App.BankKeeper.GetBalance(suite.Ctx, acc.GetAddress(), "stake")
	suite.Require().Equal(sdk.NewInt(30), balance.Amount)
}

func (suite *KeeperTestSuite) TestAddTokensToLock() {
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
			tokenToAdd:     sdk.NewCoin("unknown", sdk.NewInt(10)),
			duration:       time.Second,
			lockingAddress: addr1,
		},
		{
			name:           "token to add exceeds balance",
			tokenToAdd:     sdk.NewCoin("stake", sdk.NewInt(20)),
			duration:       time.Second,
			lockingAddress: addr1,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		// lock with balance
		suite.FundAcc(addr1, sdk.Coins{initialLockCoin})
		originalLock, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, addr1, sdk.Coins{initialLockCoin}, time.Second)
		suite.Require().NoError(err)

		suite.FundAcc(addr1, sdk.Coins{initialLockCoin})
		balanceBeforeLock := suite.App.BankKeeper.GetAllBalances(suite.Ctx, tc.lockingAddress)

		lockID, err := suite.App.LockupKeeper.AddToExistingLock(suite.Ctx, tc.lockingAddress, tc.tokenToAdd, tc.duration)

		if tc.expectAddTokensToLockSuccess {
			suite.Require().NoError(err)

			// get updated lock
			lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockID)
			suite.Require().NoError(err)

			// check that tokens have been added successfully to the lock
			suite.Require().True(sdk.Coins{initialLockCoin.Add(tc.tokenToAdd)}.IsEqual(lock.Coins))

			// check balance has decreased
			balanceAfterLock := suite.App.BankKeeper.GetAllBalances(suite.Ctx, tc.lockingAddress)
			suite.Require().True(balanceBeforeLock.IsEqual(balanceAfterLock.Add(tc.tokenToAdd)))

			// check accumulation store
			accum := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
				LockQueryType: types.ByDuration,
				Denom:         "stake",
				Duration:      time.Second,
			})
			suite.Require().Equal(initialLockCoin.Amount.Add(tc.tokenToAdd.Amount), accum)
		} else {
			suite.Require().Error(err)
			suite.Require().Equal(uint64(0), lockID)

			lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, originalLock.ID)
			suite.Require().NoError(err)

			// check that locked coins haven't changed
			suite.Require().True(lock.Coins.IsEqual(sdk.Coins{initialLockCoin}))

			// check accumulation store didn't change
			accum := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
				LockQueryType: types.ByDuration,
				Denom:         "stake",
				Duration:      time.Second,
			})
			suite.Require().Equal(initialLockCoin.Amount, accum)
		}
	}
}

func (suite *KeeperTestSuite) TestHasLock() {
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
			denomToQuery:    "uosmo",
			durationToQuery: time.Minute,
			expectedHas:     false,
		},
		{
			name:            "same token, same duration, different address",
			tokenLocked:     sdk.NewInt64Coin("stake", 10),
			durationLocked:  time.Minute,
			lockAddr:        addr2,
			denomToQuery:    "uosmo",
			durationToQuery: time.Minute,
			expectedHas:     false,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest()

		suite.FundAcc(tc.lockAddr, sdk.Coins{tc.tokenLocked})
		_, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, tc.lockAddr, sdk.Coins{tc.tokenLocked}, tc.durationLocked)
		suite.Require().NoError(err)

		hasLock := suite.App.LockupKeeper.HasLock(suite.Ctx, addr1, tc.denomToQuery, tc.durationToQuery)
		suite.Require().Equal(tc.expectedHas, hasLock)
	}
}

func (suite *KeeperTestSuite) TestLock() {
	suite.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

	lock := types.PeriodLock{
		ID:       1,
		Owner:    addr1.String(),
		Duration: time.Second,
		EndTime:  time.Time{},
		Coins:    coins,
	}

	// test locking without balance
	err := suite.App.LockupKeeper.Lock(suite.Ctx, lock, coins)
	suite.Require().Error(err)

	// check accumulation store
	accum := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "0")

	suite.FundAcc(addr1, coins)
	err = suite.App.LockupKeeper.Lock(suite.Ctx, lock, coins)
	suite.Require().NoError(err)

	// check accumulation store
	accum = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")

	balance := suite.App.BankKeeper.GetBalance(suite.Ctx, addr1, "stake")
	suite.Require().Equal(sdk.ZeroInt(), balance.Amount)

	acc := suite.App.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)
	balance = suite.App.BankKeeper.GetBalance(suite.Ctx, acc.GetAddress(), "stake")
	suite.Require().Equal(sdk.NewInt(10), balance.Amount)
}

func (suite *KeeperTestSuite) AddTokensToLockForSynth() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// lock coins on other durations
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 20)}
	suite.LockTokens(addr1, coins, time.Second*2)
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 30)}
	suite.LockTokens(addr1, coins, time.Second*3)

	synthlocks := []types.SyntheticLock{}
	// make three synthetic locks on each locks
	for i := uint64(1); i <= 3; i++ {
		// testing not unlocking synthlock, with same duration with underlying
		synthlock := types.SyntheticLock{
			UnderlyingLockId: i,
			SynthDenom:       fmt.Sprintf("synth1/%d", i),
			Duration:         time.Second * time.Duration(i),
		}
		err := suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, i, synthlock.SynthDenom, synthlock.Duration, false)
		suite.Require().NoError(err)
		synthlocks = append(synthlocks, synthlock)

		// testing not unlocking synthlock, different duration with underlying
		synthlock.SynthDenom = fmt.Sprintf("synth2/%d", i)
		synthlock.Duration = time.Second * time.Duration(i) / 2
		err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, i, synthlock.SynthDenom, synthlock.Duration, false)
		suite.Require().NoError(err)
		synthlocks = append(synthlocks, synthlock)

		// testing unlocking synthlock, different duration with underlying
		synthlock.SynthDenom = fmt.Sprintf("synth3/%d", i)
		err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, i, synthlock.SynthDenom, synthlock.Duration, true)
		suite.Require().NoError(err)
		synthlocks = append(synthlocks, synthlock)
	}

	// check synthlocks are all set
	checkSynthlocks := func(amounts []uint64) {
		// by GetAllSyntheticLockups
		for i, synthlock := range suite.App.LockupKeeper.GetAllSyntheticLockups(suite.Ctx) {
			suite.Require().Equal(synthlock, synthlocks[i])
		}
		// by GetAllSyntheticLockupsByLockup
		for i := uint64(1); i <= 3; i++ {
			for j, synthlockByLockup := range suite.App.LockupKeeper.GetAllSyntheticLockupsByLockup(suite.Ctx, i) {
				suite.Require().Equal(synthlockByLockup, synthlocks[(int(i)-1)*3+j])
			}
		}
		// by GetAllSyntheticLockupsByAddr
		for i, synthlock := range suite.App.LockupKeeper.GetAllSyntheticLockupsByAddr(suite.Ctx, addr1) {
			suite.Require().Equal(synthlock, synthlocks[i])
		}
		// by GetPeriodLocksAccumulation
		for i := 1; i <= 3; i++ {
			for j := 1; j <= 3; j++ {
				// get accumulation with always-qualifying condition
				acc := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
					Denom:    fmt.Sprintf("synth%d/%d", j, i),
					Duration: time.Second / 10,
				})
				// amount retrieved should be equal with underlying lock's locked amount
				suite.Require().Equal(acc.Int64(), amounts[i])

				// get accumulation with non-qualifying condition
				acc = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
					Denom:    fmt.Sprintf("synth%d/%d", j, i),
					Duration: time.Second * 100,
				})
				suite.Require().Equal(acc.Int64(), 0)
			}
		}
	}

	checkSynthlocks([]uint64{10, 20, 30})

	// call AddTokensToLock
	for i := uint64(1); i <= 3; i++ {
		coins := sdk.NewInt64Coin("stake", int64(i)*10)
		suite.FundAcc(addr1, sdk.Coins{coins})
		_, err := suite.App.LockupKeeper.AddTokensToLockByID(suite.Ctx, i, addr1, coins)
		suite.Require().NoError(err)
	}

	// check if all invariants holds after calling AddTokensToLock
	checkSynthlocks([]uint64{20, 40, 60})
}

func (suite *KeeperTestSuite) TestEndblockerWithdrawAllMaturedLockups() {
	suite.SetupTest()
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
			unbondBlockTimes[i] = suite.Ctx.BlockTime().Add(sortedTimes[i])
		}

		for i := 0; i < len(times); i++ {
			suite.LockTokens(addr1, coins, times[i])
		}

		// consistency check locks
		locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
		suite.Require().NoError(err)
		suite.Require().Len(locks, 3)
		for i := 0; i < len(times); i++ {
			suite.Require().Equal(locks[i].EndTime, time.Time{})
			suite.Require().Equal(locks[i].IsUnlocking(), false)
		}

		// begin unlock
		locks, err = suite.App.LockupKeeper.BeginUnlockAllNotUnlockings(suite.Ctx, addr1)
		unlockedCoins := suite.App.LockupKeeper.GetCoinsFromLocks(locks)
		suite.Require().NoError(err)
		suite.Require().Len(locks, len(times))
		suite.Require().Equal(unlockedCoins, totalCoins)
		for i := 0; i < len(times); i++ {
			suite.Require().Equal(sortedTimesIndex[i], locks[i].ID)
		}

		// check locks, these should now be sorted by unbonding completion time
		locks, err = suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
		suite.Require().NoError(err)
		suite.Require().Len(locks, 3)
		for i := 0; i < 3; i++ {
			suite.Require().NotEqual(locks[i].EndTime, time.Time{})
			suite.Require().Equal(locks[i].EndTime, unbondBlockTimes[i])
			suite.Require().Equal(locks[i].IsUnlocking(), true)
		}
	}
	setupInitLocks()

	// try withdrawing before mature
	suite.App.LockupKeeper.WithdrawAllMaturedLocks(suite.Ctx)
	locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 3)

	// withdraw at 1 sec, 3 sec, and 5 sec intervals, check automatically withdrawn
	for i := 0; i < len(times); i++ {
		suite.App.LockupKeeper.WithdrawAllMaturedLocks(suite.Ctx.WithBlockTime(unbondBlockTimes[i]))
		locks, err = suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
		suite.Require().NoError(err)
		suite.Require().Len(locks, len(times)-i-1)
	}

	// We expect that only non-CL locks (i.e. locks that do not have the CL token prefix) send tokens back to the user's balance when mature. This is because CL tokens get burned after the lock matures.
	expectedCoins := sdk.NewCoins()
	for _, coin := range totalCoins {
		if !strings.HasPrefix(coin.Denom, cltypes.ClTokenPrefix) {
			expectedCoins = expectedCoins.Add(coin)
		}
	}
	suite.Require().Equal(addr1.String(), suite.App.BankKeeper.GetAccountsBalances(suite.Ctx)[1].Address)
	suite.Require().Equal(expectedCoins, suite.App.BankKeeper.GetAccountsBalances(suite.Ctx)[1].Coins)

	suite.SetupTest()
	setupInitLocks()
	// now withdraw all locks and ensure all got withdrawn
	suite.App.LockupKeeper.WithdrawAllMaturedLocks(suite.Ctx.WithBlockTime(unbondBlockTimes[len(times)-1]))
	suite.Require().Len(locks, 0)
}

func (suite *KeeperTestSuite) TestLockAccumulationStore() {
	suite.SetupTest()

	// initial check
	locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr := sdk.AccAddress([]byte("addr1---------------"))

	// 1 * time.Second: 10 + 20
	// 2 * time.Second: 20 + 30
	// 3 * time.Second: 30
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr, coins, time.Second)
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 20)}
	suite.LockTokens(addr, coins, time.Second)
	suite.LockTokens(addr, coins, time.Second*2)
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 30)}
	suite.LockTokens(addr, coins, time.Second*2)
	suite.LockTokens(addr, coins, time.Second*3)

	// check accumulations
	acc := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: 0,
	})
	suite.Require().Equal(int64(110), acc.Int64())
	acc = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 1,
	})
	suite.Require().Equal(int64(110), acc.Int64())
	acc = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 2,
	})
	suite.Require().Equal(int64(80), acc.Int64())
	acc = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 3,
	})
	suite.Require().Equal(int64(30), acc.Int64())
	acc = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 4,
	})
	suite.Require().Equal(int64(0), acc.Int64())
}

func (suite *KeeperTestSuite) TestSlashTokensFromLockByID() {
	suite.SetupTest()

	// initial check
	locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr := sdk.AccAddress([]byte("addr1---------------"))

	// 1 * time.Second: 10
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr, coins, time.Second)

	// check accumulations
	acc := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second,
	})
	suite.Require().Equal(int64(10), acc.Int64())

	suite.App.LockupKeeper.SlashTokensFromLockByID(suite.Ctx, 1, sdk.Coins{sdk.NewInt64Coin("stake", 1)})
	acc = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second,
	})
	suite.Require().Equal(int64(9), acc.Int64())

	lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(lock.Coins.String(), "9stake")

	_, err = suite.App.LockupKeeper.SlashTokensFromLockByID(suite.Ctx, 1, sdk.Coins{sdk.NewInt64Coin("stake", 11)})
	suite.Require().Error(err)

	_, err = suite.App.LockupKeeper.SlashTokensFromLockByID(suite.Ctx, 1, sdk.Coins{sdk.NewInt64Coin("stake1", 1)})
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestSlashTokensFromLockByIDSendUnderlyingAndBurn() {
	testCases := []struct {
		name             string
		positionCoins    sdk.Coins
		liquidityToSlash sdk.Dec
		denomToSlash     string
		expectError      bool
	}{
		{
			name:             "happy path",
			positionCoins:    sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(1000000)), sdk.NewCoin("usdc", sdk.NewInt(5000000000))),
			liquidityToSlash: sdk.NewDec(10000000),
			denomToSlash:     cltypes.GetConcentratedLockupDenomFromPoolId(1),
		},
		{
			name:             "error: attempt to slash more liquidity than the lock has",
			positionCoins:    sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(1000000)), sdk.NewCoin("usdc", sdk.NewInt(5000000000))),
			liquidityToSlash: sdk.NewDec(100000000),
			denomToSlash:     cltypes.GetConcentratedLockupDenomFromPoolId(1),
			expectError:      true,
		},
		{
			name:             "error: attempt to slash a denom that does not exist in the lock",
			positionCoins:    sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(1000000)), sdk.NewCoin("usdc", sdk.NewInt(5000000000))),
			denomToSlash:     cltypes.GetConcentratedLockupDenomFromPoolId(2),
			liquidityToSlash: sdk.NewDec(10000000),
			expectError:      true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest()

		// Check that there are currently no locks
		locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
		suite.Require().NoError(err)
		suite.Require().Len(locks, 0)

		// Fund the account we will be using
		addr := suite.TestAccs[0]
		suite.FundAcc(addr, tc.positionCoins)

		// Create a cl pool and a locked full range position
		clPool := suite.PrepareConcentratedPool()
		clPoolId := clPool.GetId()
		positionID, _, _, liquidity, _, concentratedLockId, err := suite.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(suite.Ctx, clPoolId, addr, tc.positionCoins, time.Hour)

		// Refetch the cl pool post full range position creation
		clPool, err = suite.App.ConcentratedLiquidityKeeper.GetPoolFromPoolIdAndConvertToConcentrated(suite.Ctx, clPoolId)
		suite.Require().NoError(err)

		clPoolPositionDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPoolId)

		// Check the supply of the cl asset before we slash it is equal to the liquidity created
		clAssetSupplyPreSlash := suite.App.BankKeeper.GetSupply(suite.Ctx, clPoolPositionDenom)
		suite.Require().Equal(liquidity.TruncateInt().String(), clAssetSupplyPreSlash.Amount.String())

		// Store the cl pool balance before the slash
		clPoolBalancePreSlash := suite.App.BankKeeper.GetAllBalances(suite.Ctx, clPool.GetAddress())

		// Check the period locks accumulation
		acc := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
			Denom:    clPoolPositionDenom,
			Duration: time.Second,
		})
		suite.Require().Equal(liquidity.TruncateInt64(), acc.Int64())

		// The lockup module account balance before the slash should match the liquidity added to the lock
		lockupModuleBalancePreSlash := suite.App.LockupKeeper.GetModuleBalance(suite.Ctx)
		suite.Require().Equal(sdk.NewCoins(sdk.NewCoin(clPoolPositionDenom, liquidity.TruncateInt())), lockupModuleBalancePreSlash)

		// Slash the cl shares and the underlying assets
		// Figure out the underlying assets from the liquidity slash
		position, err := suite.App.ConcentratedLiquidityKeeper.GetPosition(suite.Ctx, positionID)
		suite.Require().NoError(err)

		concentratedPool, err := suite.App.ConcentratedLiquidityKeeper.GetPoolFromPoolIdAndConvertToConcentrated(suite.Ctx, position.PoolId)
		suite.Require().NoError(err)

		tempPositionToCalculateUnderlyingAssets := position
		tempPositionToCalculateUnderlyingAssets.Liquidity = tc.liquidityToSlash
		asset0, asset1, err := cl.CalculateUnderlyingAssetsFromPosition(suite.Ctx, tempPositionToCalculateUnderlyingAssets, concentratedPool)
		suite.Require().NoError(err)

		underlyingAssetsToSlash := sdk.NewCoins(asset0, asset1)

		// The expected new liquidity is the previous liquidity minus the slashed liquidity
		expectedNewLiquidity := position.Liquidity.Sub(tc.liquidityToSlash).TruncateInt()

		// Slash the tokens from the lock
		_, err = suite.App.LockupKeeper.SlashTokensFromLockByIDSendUnderlyingAndBurn(suite.Ctx, concentratedLockId, sdk.Coins{sdk.NewInt64Coin(tc.denomToSlash, tc.liquidityToSlash.TruncateInt64())}, underlyingAssetsToSlash, clPool.GetAddress())
		if tc.expectError {
			suite.Require().Error(err)
			continue
		} else {
			suite.Require().NoError(err)
		}

		expectedNewLiquidityCoins := sdk.NewCoin(clPoolPositionDenom, expectedNewLiquidity)

		acc = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
			Denom:    tc.denomToSlash,
			Duration: time.Second,
		})
		suite.Require().Equal(expectedNewLiquidityCoins.Amount.Int64(), acc.Int64())

		// The lockup module account balance after the slash should match the liquidity minus the slashed liquidity
		lockupModuleBalancePostSlash := suite.App.LockupKeeper.GetModuleBalance(suite.Ctx)
		suite.Require().Equal(sdk.NewCoins(expectedNewLiquidityCoins), lockupModuleBalancePostSlash)

		// Check the supply of the cl asset after we slash it is equal to the liquidity created
		clAssetSupplyPostSlash := suite.App.BankKeeper.GetSupply(suite.Ctx, clPoolPositionDenom)
		suite.Require().Equal(expectedNewLiquidityCoins.Amount.String(), clAssetSupplyPostSlash.Amount.String())

		// The lock itself should have been slashed
		lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, concentratedLockId)
		suite.Require().NoError(err)
		suite.Require().Equal(expectedNewLiquidityCoins.String(), lock.Coins.String())

		// The cl pool should be missing the underlying assets that were slashed
		clPoolBalancePostSlash := suite.App.BankKeeper.GetAllBalances(suite.Ctx, clPool.GetAddress())
		suite.Require().Equal(clPoolBalancePreSlash.Sub(underlyingAssetsToSlash), clPoolBalancePostSlash)
	}
}

func (suite *KeeperTestSuite) TestEditLockup() {
	suite.SetupTest()

	// initial check
	locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr := sdk.AccAddress([]byte("addr1---------------"))

	// 1 * time.Second: 10
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr, coins, time.Second)

	// check accumulations
	acc := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second,
	})
	suite.Require().Equal(int64(10), acc.Int64())

	lock, _ := suite.App.LockupKeeper.GetLockByID(suite.Ctx, 1)

	// duration decrease should fail
	err = suite.App.LockupKeeper.ExtendLockup(suite.Ctx, lock.ID, addr, time.Second/2)
	suite.Require().Error(err)
	// extending lock with same duration should fail
	err = suite.App.LockupKeeper.ExtendLockup(suite.Ctx, lock.ID, addr, time.Second)
	suite.Require().Error(err)

	// duration increase should success
	err = suite.App.LockupKeeper.ExtendLockup(suite.Ctx, lock.ID, addr, time.Second*2)
	suite.Require().NoError(err)

	// check queries
	lock, _ = suite.App.LockupKeeper.GetLockByID(suite.Ctx, lock.ID)
	suite.Require().Equal(lock.Duration, time.Second*2)
	suite.Require().Equal(uint64(1), lock.ID)
	suite.Require().Equal(coins, lock.Coins)

	locks = suite.App.LockupKeeper.GetLocksLongerThanDurationDenom(suite.Ctx, "stake", time.Second)
	suite.Require().Equal(len(locks), 1)

	locks = suite.App.LockupKeeper.GetLocksLongerThanDurationDenom(suite.Ctx, "stake", time.Second*2)
	suite.Require().Equal(len(locks), 1)

	// check accumulations
	acc = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second,
	})
	suite.Require().Equal(int64(10), acc.Int64())
	acc = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 2,
	})
	suite.Require().Equal(int64(10), acc.Int64())
	acc = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 3,
	})
	suite.Require().Equal(int64(0), acc.Int64())
}

func (suite *KeeperTestSuite) TestForceUnlock() {
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
				err := suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, 1, "testDenom", time.Minute, true)
				suite.Require().NoError(err)
			},
		},
	}
	for _, tc := range testCases {
		// set up test and create default lock
		suite.SetupTest()
		coinsToLock := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10000000)))
		suite.FundAcc(addr1, sdk.NewCoins(coinsToLock...))
		lock, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, addr1, coinsToLock, time.Minute)
		suite.Require().NoError(err)

		// post lock setup
		if tc.postLockSetup != nil {
			tc.postLockSetup()
		}

		err = suite.App.LockupKeeper.ForceUnlock(suite.Ctx, lock)
		suite.Require().NoError(err)

		// check that accumulation store has decreased
		accum := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
			LockQueryType: types.ByDuration,
			Denom:         "foo",
			Duration:      time.Minute,
		})
		suite.Require().Equal(accum.String(), "0")

		// check balance of lock account to confirm
		balances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr1)
		suite.Require().Equal(balances, coinsToLock)

		// if it was superfluid delegated lock,
		//  confirm that we don't have associated synth locks
		synthLocks := suite.App.LockupKeeper.GetAllSyntheticLockupsByLockup(suite.Ctx, lock.ID)
		suite.Require().Equal(0, len(synthLocks))

		// check if lock is deleted by checking trying to get lock ID
		_, err = suite.App.LockupKeeper.GetLockByID(suite.Ctx, lock.ID)
		suite.Require().Error(err)
	}
}

func (suite *KeeperTestSuite) TestPartialForceUnlock() {
	addr1 := sdk.AccAddress([]byte("addr1---------------"))

	defaultDenomToLock := "stake"
	defaultAmountToLock := sdk.NewInt(10000000)

	testCases := []struct {
		name               string
		coinsToForceUnlock sdk.Coins
		expectedPass       bool
	}{
		{
			name:               "unlock full amount",
			coinsToForceUnlock: sdk.Coins{sdk.NewCoin(defaultDenomToLock, defaultAmountToLock)},
			expectedPass:       true,
		},
		{
			name:               "partial unlock",
			coinsToForceUnlock: sdk.Coins{sdk.NewCoin(defaultDenomToLock, defaultAmountToLock.Quo(sdk.NewInt(2)))},
			expectedPass:       true,
		},
		{
			name:               "unlock more than locked",
			coinsToForceUnlock: sdk.Coins{sdk.NewCoin(defaultDenomToLock, defaultAmountToLock.Add(sdk.NewInt(2)))},
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
		suite.SetupTest()
		coinsToLock := sdk.NewCoins(sdk.NewCoin("stake", defaultAmountToLock))
		suite.FundAcc(addr1, sdk.NewCoins(coinsToLock...))
		// balanceBeforeLock := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr1)
		lock, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, addr1, coinsToLock, time.Minute)
		suite.Require().NoError(err)

		err = suite.App.LockupKeeper.PartialForceUnlock(suite.Ctx, lock, tc.coinsToForceUnlock)

		if tc.expectedPass {
			suite.Require().NoError(err)

			// check balance
			balanceAfterForceUnlock := suite.App.BankKeeper.GetBalance(suite.Ctx, addr1, "stake")

			if tc.coinsToForceUnlock.Empty() {
				tc.coinsToForceUnlock = coinsToLock
			}
			suite.Require().Equal(tc.coinsToForceUnlock, sdk.Coins{balanceAfterForceUnlock})
		} else {
			suite.Require().Error(err)

			// check balance
			balanceAfterForceUnlock := suite.App.BankKeeper.GetBalance(suite.Ctx, addr1, "stake")
			suite.Require().Equal(sdk.NewInt(0), balanceAfterForceUnlock.Amount)
		}
	}
}
