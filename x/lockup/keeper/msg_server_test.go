package keeper_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
)

func (s *KeeperTestSuite) TestMsgLockTokens() {
	type param struct {
		coinsToLock         sdk.Coins
		lockOwner           sdk.AccAddress
		duration            time.Duration
		coinsInOwnerAddress sdk.Coins
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "creation of lock via lockTokens",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: true,
		},
		{
			name: "locking more coins than are in the address",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 20)},       // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		s.SetupTest()

		s.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)

		msgServer := keeper.NewMsgServerImpl(s.App.LockupKeeper)
		c := s.Ctx
		_, err := msgServer.LockTokens(c, types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))

		if test.expectPass {
			// creation of lock via LockTokens
			msgServer := keeper.NewMsgServerImpl(s.App.LockupKeeper)
			_, _ = msgServer.LockTokens(s.Ctx, types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))

			// Check Locks
			locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
			s.Require().NoError(err)
			s.Require().Len(locks, 1)
			s.Require().Equal(locks[0].Coins, test.param.coinsToLock)

			// check accumulation store is correctly updated
			accum := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
				LockQueryType: types.ByDuration,
				Denom:         "stake",
				Duration:      test.param.duration,
			})
			s.Require().Equal(accum.String(), "10")

			// add more tokens to lock via LockTokens
			s.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)

			_, err = msgServer.LockTokens(s.Ctx, types.NewMsgLockTokens(test.param.lockOwner, locks[0].Duration, test.param.coinsToLock))
			s.Require().NoError(err)

			// check locks after adding tokens to lock
			locks, err = s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
			s.Require().NoError(err)
			s.Require().Len(locks, 1)
			s.Require().Equal(locks[0].Coins, test.param.coinsToLock.Add(test.param.coinsToLock...))

			// check accumulation store is correctly updated
			accum = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
				LockQueryType: types.ByDuration,
				Denom:         "stake",
				Duration:      test.param.duration,
			})
			s.Require().Equal(accum.String(), "20")
		} else {
			// Fail simple lock token
			s.Require().Error(err)
		}
	}
}

func (s *KeeperTestSuite) TestMsgBeginUnlocking() {
	type param struct {
		coinsToLock         sdk.Coins
		isSyntheticLockup   bool
		coinsToUnlock       sdk.Coins
		lockOwner           sdk.AccAddress
		duration            time.Duration
		coinsInOwnerAddress sdk.Coins
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
		isPartial  bool
	}{
		{
			name: "unlock full amount of tokens via begin unlock",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup:   false,
				coinsToUnlock:       sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: true,
		},
		{
			name: "unlock partial amount of tokens via begin unlock",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup:   false,
				coinsToUnlock:       sdk.Coins{sdk.NewInt64Coin("stake", 5)},        // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			isPartial:  true,
			expectPass: true,
		},
		{
			name: "unlock zero amount of tokens via begin unlock",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup:   false,
				coinsToUnlock:       sdk.Coins{},                                    // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: true,
		},
		{
			name: "unlock partial amount of tokens via begin unlock for lockup with synthetic versions",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup:   true,
				coinsToUnlock:       sdk.Coins{sdk.NewInt64Coin("stake", 5)},        // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: false,
			isPartial:  true,
		},
	}

	for _, test := range tests {
		s.SetupTest()

		s.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)

		msgServer := keeper.NewMsgServerImpl(s.App.LockupKeeper)
		goCtx := s.Ctx
		resp, err := msgServer.LockTokens(goCtx, types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))
		s.Require().NoError(err)

		if test.param.isSyntheticLockup {
			err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, resp.ID, "synthetic", time.Second, false)
			s.Require().NoError(err)
		}

		unlockingResponse, err := msgServer.BeginUnlocking(goCtx, types.NewMsgBeginUnlocking(test.param.lockOwner, resp.ID, test.param.coinsToUnlock))

		if test.expectPass {
			s.Require().NoError(err)
			s.AssertEventEmitted(s.Ctx, types.TypeEvtBeginUnlock, 1)
			s.Require().True(unlockingResponse.Success)
			if test.isPartial {
				s.Require().Equal(unlockingResponse.UnlockingLockID, resp.ID+1)
			} else {
				s.Require().Equal(unlockingResponse.UnlockingLockID, resp.ID)
			}
		} else {
			s.Require().Error(err)
			s.AssertEventEmitted(s.Ctx, types.TypeEvtBeginUnlock, 0)
		}
	}
}

func (s *KeeperTestSuite) TestMsgBeginUnlockingAll() {
	type param struct {
		coinsToLock         sdk.Coins
		isSyntheticLockup   bool
		lockOwner           sdk.AccAddress
		duration            time.Duration
		coinsInOwnerAddress sdk.Coins
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "unlock all lockups",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup:   false,
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: true,
		},
		{
			name: "unlock all when synthetic versions exists",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup:   true,
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		s.SetupTest()

		s.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)

		msgServer := keeper.NewMsgServerImpl(s.App.LockupKeeper)
		c := s.Ctx
		resp, err := msgServer.LockTokens(c, types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))
		s.Require().NoError(err)

		if test.param.isSyntheticLockup {
			err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, resp.ID, "synthetic", time.Second, false)
			s.Require().NoError(err)
		}

		_, err = msgServer.BeginUnlockingAll(c, types.NewMsgBeginUnlockingAll(test.param.lockOwner))

		if test.expectPass {
			s.Require().NoError(err)
		} else {
			s.Require().Error(err)
		}
	}
}

func (s *KeeperTestSuite) TestMsgEditLockup() {
	type param struct {
		coinsToLock       sdk.Coins
		isSyntheticLockup bool
		lockOwner         sdk.AccAddress
		duration          time.Duration
		newDuration       time.Duration
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "edit lockups by duration",
			param: param{
				coinsToLock:       sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup: false,
				lockOwner:         sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:          time.Second,
				newDuration:       time.Second * 2,
			},
			expectPass: true,
		},
		{
			name: "edit lockups by lesser duration",
			param: param{
				coinsToLock:       sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup: false,
				lockOwner:         sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:          time.Second,
				newDuration:       time.Second / 2,
			},
			expectPass: false,
		},
		{
			name: "disallow edit when synthetic lockup exists",
			param: param{
				coinsToLock:       sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup: true,
				lockOwner:         sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:          time.Second,
				newDuration:       time.Second * 2,
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		s.SetupTest()

		err := testutil.FundAccount(s.Ctx, s.App.BankKeeper, test.param.lockOwner, test.param.coinsToLock)
		s.Require().NoError(err)

		msgServer := keeper.NewMsgServerImpl(s.App.LockupKeeper)
		c := s.Ctx
		resp, err := msgServer.LockTokens(c, types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))
		s.Require().NoError(err)

		if test.param.isSyntheticLockup {
			err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, resp.ID, "synthetic", time.Second, false)
			s.Require().NoError(err)
		}

		_, err = msgServer.ExtendLockup(c, types.NewMsgExtendLockup(test.param.lockOwner, resp.ID, test.param.newDuration))

		if test.expectPass {
			s.Require().NoError(err, test.name)
		} else {
			s.Require().Error(err, test.name)
		}
	}
}

func (s *KeeperTestSuite) TestMsgForceUnlock() {
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	defaultPoolID, defaultLockID := uint64(1), uint64(1)
	defaultLockAmount := osmomath.NewInt(1000000000)

	tests := []struct {
		name                      string
		forceUnlockAllowedAddress types.Params
		postLockSetup             func()
		forceUnlockAmount         osmomath.Int
		expectPass                bool
	}{
		{
			"happy path",
			types.Params{ForceUnlockAllowedAddresses: []string{addr1.String()}},
			func() {},
			defaultLockAmount,
			true,
		},
		{
			"force unlock superfluid delegated lock",
			types.Params{ForceUnlockAllowedAddresses: []string{addr1.String()}},
			func() {
				err := s.SuperfluidDelegateToDefaultVal(addr1, defaultPoolID, defaultLockID)
				s.Require().NoError(err)
			},
			defaultLockAmount,
			false,
		},
		{
			"superfluid undelegating lock",
			types.Params{ForceUnlockAllowedAddresses: []string{addr1.String()}},
			func() {
				err := s.SuperfluidDelegateToDefaultVal(addr1, defaultPoolID, defaultLockID)
				s.Require().NoError(err)

				err = s.App.SuperfluidKeeper.SuperfluidUndelegate(s.Ctx, addr1.String(), defaultLockID)
				s.Require().NoError(err)
			},
			defaultLockAmount,
			false,
		},
		{
			"partial unlock",
			types.Params{ForceUnlockAllowedAddresses: []string{addr1.String()}},
			func() {},
			// try force unlocking half of locked amount
			defaultLockAmount.Quo(osmomath.NewInt(2)),
			true,
		},
		{
			"force unlock more than what we have locked",
			types.Params{ForceUnlockAllowedAddresses: []string{addr1.String()}},
			func() {},
			// try force more than the locked amount
			defaultLockAmount.Add(osmomath.NewInt(1)),
			false,
		},
		{
			"params with different address",
			types.Params{ForceUnlockAllowedAddresses: []string{addr2.String()}},
			func() {},
			defaultLockAmount,
			false,
		},
		{
			"param with multiple addresses ",
			types.Params{ForceUnlockAllowedAddresses: []string{addr1.String(), addr2.String()}},
			func() {},
			defaultLockAmount,
			true,
		},
	}

	for _, test := range tests {
		// set up test
		s.SetupTest()
		s.App.LockupKeeper.SetParams(s.Ctx, test.forceUnlockAllowedAddress)

		// prepare pool for superfluid staking cases
		poolId := s.PrepareBalancerPoolWithCoins(sdk.NewCoin("stake", osmomath.NewInt(1000000000000)), sdk.NewCoin("foo", osmomath.NewInt(5000)))

		// lock tokens
		msgServer := keeper.NewMsgServerImpl(s.App.LockupKeeper)
		c := s.Ctx

		poolDenom := gammtypes.GetPoolShareDenom(poolId)
		coinsToLock := sdk.Coins{sdk.NewCoin(poolDenom, defaultLockAmount)}
		s.FundAcc(addr1, coinsToLock)

		stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
		s.Require().NoError(err)
		unbondingDuration := stakingParams.UnbondingTime
		resp, err := msgServer.LockTokens(c, types.NewMsgLockTokens(addr1, unbondingDuration, coinsToLock))
		s.Require().NoError(err)

		// setup env after lock tokens
		test.postLockSetup()

		// test force unlock
		_, err = msgServer.ForceUnlock(c, types.NewMsgForceUnlock(addr1, resp.ID, sdk.Coins{sdk.NewCoin(poolDenom, test.forceUnlockAmount)}))
		if test.expectPass {
			s.Require().NoError(err)

			// check that we have successfully force unlocked
			balanceAfterForceUnlock := s.App.BankKeeper.GetBalance(s.Ctx, addr1, poolDenom)
			s.Require().Equal(test.forceUnlockAmount, balanceAfterForceUnlock.Amount)
		} else {
			s.Require().Error(err)

			// check that we have successfully force unlocked
			balanceAfterForceUnlock := s.App.BankKeeper.GetBalance(s.Ctx, addr1, poolDenom)
			s.Require().NotEqual(test.forceUnlockAmount, balanceAfterForceUnlock.Amount)
			return
		}
	}
}

func (s *KeeperTestSuite) TestSetRewardReceiverAddress() {
	type param struct {
		isOwner                      bool
		isRewardReceiverAddressOwner bool
		isLockOwner                  bool
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "happy path: change reward receiver address to another address",
			param: param{
				isOwner:                      true,
				isRewardReceiverAddressOwner: false,
				isLockOwner:                  true,
			},
			expectPass: true,
		},
		{
			name: "error: attempt to try changing to same owner",
			param: param{
				isOwner:                      false,
				isRewardReceiverAddressOwner: true,
				isLockOwner:                  true,
			},
			expectPass: false,
		},
		{
			name: "error: sender is not the owner of the lock",
			param: param{
				isOwner:                      false,
				isRewardReceiverAddressOwner: false,
				isLockOwner:                  true,
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		s.SetupTest()

		defaultAmountInLock := sdk.NewCoins(sdk.NewInt64Coin("foo", 100))
		s.FundAcc(s.TestAccs[0], defaultAmountInLock)

		lock, err := s.App.LockupKeeper.CreateLock(s.Ctx, s.TestAccs[0], defaultAmountInLock, time.Minute)
		s.Require().NoError(err)

		// lock reward receiver address should initially be an empty string
		s.Require().Equal(lock.RewardReceiverAddress, "")

		msgServer := keeper.NewMsgServerImpl(s.App.LockupKeeper)
		c := s.Ctx

		owner := s.TestAccs[0]
		if !test.param.isOwner {
			owner = s.TestAccs[1]
		}

		rewardReceiver := s.TestAccs[0]
		if !test.param.isRewardReceiverAddressOwner {
			rewardReceiver = s.TestAccs[1]
		}

		lockId := lock.ID
		if !test.param.isLockOwner {
			lockId = lockId + 1
		}
		// System under test
		msg := types.NewMsgSetRewardReceiverAddress(owner, rewardReceiver, lockId)
		resp, err := msgServer.SetRewardReceiverAddress(c, msg)
		if !test.expectPass {
			s.Require().Error(err)
			s.Require().Equal(resp.Success, false)
			return
		}

		s.Require().NoError(err)
		s.Require().Equal(resp.Success, true)

		// now check if the reward receiver address has been changed
		newLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
		s.Require().NoError(err)
		if test.param.isRewardReceiverAddressOwner {
			s.Require().Equal(newLock.RewardReceiverAddress, "")
		} else {
			s.Require().Equal(s.TestAccs[1].String(), newLock.RewardReceiverAddress)
		}

	}
}
