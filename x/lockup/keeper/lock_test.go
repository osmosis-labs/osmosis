package keeper_test

import (
	"fmt"
	"time"

	"github.com/osmosis-labs/osmosis/v12/x/lockup/types"

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

	testCases := []struct {
		name                          string
		unlockingCoins                sdk.Coins
		expectedBeginUnlockPass       bool
		passedTime                    time.Duration
		expectedUnlockMaturedLockPass bool
		balanceAfterUnlock            sdk.Coins
	}{
		{
			name:                          "normal unlocking case",
			unlockingCoins:                initialLockCoins,
			expectedBeginUnlockPass:       true,
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: true,
			balanceAfterUnlock:            initialLockCoins,
		},
		{
			name:                          "begin unlocking with nil as unlocking coins",
			unlockingCoins:                nil,
			expectedBeginUnlockPass:       true,
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: true,
			balanceAfterUnlock:            initialLockCoins,
		},
		{
			name:                          "unlocking coins exceed what's in lock",
			unlockingCoins:                sdk.Coins{sdk.NewInt64Coin("stake", 20)},
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: false,
			balanceAfterUnlock:            sdk.Coins{},
		},
		{
			name:                          "unlocking unknown tokens",
			unlockingCoins:                sdk.Coins{sdk.NewInt64Coin("unknown", 10)},
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: false,
			balanceAfterUnlock:            sdk.Coins{},
		},
		{
			name:                          "partial unlocking",
			unlockingCoins:                sdk.Coins{sdk.NewInt64Coin("stake", 5)},
			expectedBeginUnlockPass:       true,
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: true,
			balanceAfterUnlock:            sdk.Coins{sdk.NewInt64Coin("stake", 5)},
		},
		{
			name:                          "partial unlocking unknown tokens",
			unlockingCoins:                sdk.Coins{sdk.NewInt64Coin("unknown", 5)},
			passedTime:                    time.Second,
			expectedUnlockMaturedLockPass: false,
			balanceAfterUnlock:            sdk.Coins{},
		},
		{
			name:                          "unlocking should not finish yet",
			unlockingCoins:                initialLockCoins,
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
		lock := types.NewPeriodLock(1, addr1, time.Second, time.Time{}, initialLockCoins)

		// lock with balance
		suite.FundAcc(addr1, initialLockCoins)
		lock, err := lockupKeeper.CreateLock(ctx, addr1, initialLockCoins, time.Second)
		suite.Require().NoError(err)

		// store in variable if we're testing partial unlocking for future use
		partialUnlocking := tc.unlockingCoins.IsAllLT(initialLockCoins) && tc.unlockingCoins != nil

		// begin unlocking
		err = lockupKeeper.BeginUnlock(ctx, lock.ID, tc.unlockingCoins)

		if tc.expectedBeginUnlockPass {
			suite.Require().NoError(err)

			// check unlocking coins. When a lock is a partial lock
			// (i.e. tc.unlockingCoins is not nit and less than initialLockCoins),
			// we only unlock the partial amount of tc.unlockingCoins
			expectedUnlockingCoins := tc.unlockingCoins
			if expectedUnlockingCoins == nil {
				expectedUnlockingCoins = initialLockCoins
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
			suite.Require().Equal(initialLockCoins[0], lockedCoins[0])
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
					actualUnlockingCoins = initialLockCoins
				}
				suite.Require().Equal(expectedUnlockingCoins, actualUnlockingCoins)
			}
		}

		balance := bankKeeper.GetAllBalances(ctx, addr1)
		suite.Require().Equal(tc.balanceAfterUnlock, balance)
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

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
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
	suite.Require().Equal(suite.App.BankKeeper.GetAccountsBalances(suite.Ctx)[1].Address, addr1.String())
	suite.Require().Equal(suite.App.BankKeeper.GetAccountsBalances(suite.Ctx)[1].Coins, totalCoins)

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
