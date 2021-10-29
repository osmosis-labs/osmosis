package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/epochs"
	"github.com/osmosis-labs/osmosis/x/epochs/types"
	incentivestypes "github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

var (
	DayDuration          = time.Hour * 24
	SevenDaysDuration    = time.Hour * 24 * 7
	FourteenDaysDuration = time.Hour * 24 * 14
	OwnerAddr            = "addr1---------------"
	User1Addr            = "user1---------------"
	User2Addr            = "user2---------------"
)

func (suite *KeeperTestSuite) TestF1Distribute() {
	now, height := suite.setupEpochAndLockableDurations()

	//next epoch
	suite.nextEpoch(&now, &height)

	//new gauge
	defaultGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: DayDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin("stake", 1000)},
	}
	gauges := suite.SetupGauges([]perpGaugeDesc{defaultGauge})
	initGauge := gauges[0]
	denom := initGauge.DistributeTo.Denom
	duration := initGauge.DistributeTo.Duration
	owner := sdk.AccAddress(OwnerAddr)

	suite.nextBlock(&now, &height)

	//1st lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, DayDuration)

	//next epoch
	suite.nextEpoch(&now, &height)

	currentReward, err := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, duration)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt64Coin("lptoken", 10), currentReward.Coin)
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 1000)}, currentReward.Rewards)

	//2nd lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 40)}, DayDuration)

	//next epoch
	suite.nextEpoch(&now, &height)

	currentReward, err = suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, duration)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt64Coin("lptoken", 50), currentReward.Coin)
	suite.Require().Equal(sdk.Coins(nil), currentReward.Rewards)

	prevHistoricalReward, err := suite.app.IncentivesKeeper.GetHistoricalReward(suite.ctx, denom, duration, currentReward.Period-1)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt64DecCoin("stake", 100), prevHistoricalReward.CumulativeRewardRatio[0])
}

func (suite *KeeperTestSuite) setupEpochAndLockableDurations() (time.Time, int64) {
	epochInfos := suite.app.EpochsKeeper.AllEpochInfos(suite.ctx)
	for _, epochInfo := range epochInfos {
		suite.app.EpochsKeeper.DeleteEpochInfo(suite.ctx, epochInfo.Identifier)
	}

	//now, _ := time.Parse("2006-01-02", "2021-10-01")
	now := time.Now()
	height := int64(1)
	suite.ctx = suite.ctx.WithBlockHeight(height).WithBlockTime(now)

	epochs.InitGenesis(suite.ctx, suite.app.EpochsKeeper, types.GenesisState{
		Epochs: []types.EpochInfo{
			{
				Identifier:            "day",
				StartTime:             now,
				Duration:              DayDuration,
				CurrentEpoch:          0,
				CurrentEpochStartTime: time.Time{},
				EpochCountingStarted:  false,
			},
		},
	})

	// add additional lockable durations
	suite.app.IncentivesKeeper.SetParams(suite.ctx, incentivestypes.Params{DistrEpochIdentifier: "day"})
	suite.app.IncentivesKeeper.SetLockableDurations(suite.ctx,
		[]time.Duration{
			time.Hour * 24,
			time.Hour * 24 * 7,
			time.Hour * 24 * 14})

	return now, height
}

func (suite *KeeperTestSuite) nextEpoch(now *time.Time, height *int64) {
	epochInfo := suite.app.EpochsKeeper.AllEpochInfos(suite.ctx)[0]
	if epochInfo.StartTime.After(epochInfo.CurrentEpochStartTime) {
		*now = (*now).Add(DayDuration + time.Second)
	} else {
		*now = epochInfo.CurrentEpochStartTime.Add(DayDuration + time.Second)
	}
	*height = *height + 1
	suite.ctx = suite.ctx.WithBlockHeight(*height).WithBlockTime(*now)
	suite.app.LockupKeeper.WithdrawAllMaturedLocks(suite.ctx)
	epochs.BeginBlocker(suite.ctx, suite.app.EpochsKeeper)
}

func (suite *KeeperTestSuite) nextBlock(now *time.Time, height *int64) {
	*now = (*now).Add(time.Second)
	*height = *height + 1
	suite.ctx = suite.ctx.WithBlockHeight(*height).WithBlockTime(*now)
	suite.app.LockupKeeper.WithdrawAllMaturedLocks(suite.ctx)
	epochs.BeginBlocker(suite.ctx, suite.app.EpochsKeeper)
}

func (suite *KeeperTestSuite) setupNonPerpetualGauge(owner sdk.AccAddress, now time.Time, duration time.Duration) (string, time.Duration) {
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 1400)}
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      duration,
	}
	numEpochsPaidOver := uint64(14)
	_, initGauge := suite.CreateGauge(false, owner, coins, distrTo, now, numEpochsPaidOver)
	return initGauge.DistributeTo.Denom, initGauge.DistributeTo.Duration
}

func (suite *KeeperTestSuite) TestLockFor1Day() {
	now, height := suite.setupEpochAndLockableDurations()
	suite.nextEpoch(&now, &height)

	//new non-perpetual gauge
	owner := sdk.AccAddress(OwnerAddr)
	denom, duration := suite.setupNonPerpetualGauge(owner, now, DayDuration)

	suite.nextBlock(&now, &height)

	//lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, DayDuration)

	for i := 2; i <= 14; i++ {
		suite.nextEpoch(&now, &height)
		currentReward, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, duration)
		suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(100*(i-1)))}, currentReward.Rewards)
	}
	//gauge is finished
	suite.nextEpoch(&now, &height)
	currentReward, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, duration)
	suite.Require().Equal(sdk.Coins(nil), currentReward.Rewards)
}

func (suite *KeeperTestSuite) TestLockFor14Days() {
	now, height := suite.setupEpochAndLockableDurations()
	suite.nextEpoch(&now, &height)

	//new non-perpetual gauge
	owner := sdk.AccAddress(OwnerAddr)
	denom, _ := suite.setupNonPerpetualGauge(owner, now, DayDuration)
	suite.setupNonPerpetualGauge(owner, now, SevenDaysDuration)
	suite.setupNonPerpetualGauge(owner, now, FourteenDaysDuration)

	suite.nextBlock(&now, &height)

	//lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, FourteenDaysDuration)

	for i := 2; i <= 14; i++ {
		suite.nextEpoch(&now, &height)
		currentReward1Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, DayDuration)
		currentReward7Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, SevenDaysDuration)
		currentReward14Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, FourteenDaysDuration)
		suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(100*(i-1)))}, currentReward1Day.Rewards)
		suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(100*(i-1)))}, currentReward7Day.Rewards)
		suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(100*(i-1)))}, currentReward14Day.Rewards)
	}
	suite.nextEpoch(&now, &height)

	//test gauge is finished
	currentReward1Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, DayDuration)
	currentReward7Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, SevenDaysDuration)
	currentReward14Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, FourteenDaysDuration)
	suite.Require().Equal(sdk.Coins(nil), currentReward1Day.Rewards)
	suite.Require().Equal(sdk.Coins(nil), currentReward7Day.Rewards)
	suite.Require().Equal(sdk.Coins(nil), currentReward14Day.Rewards)
}

func (suite *KeeperTestSuite) TestLockAndUnlockFor14Days() {
	now, height := suite.setupEpochAndLockableDurations()
	suite.nextEpoch(&now, &height)

	//new non-perpetual gauge
	owner := sdk.AccAddress(OwnerAddr)
	denom, _ := suite.setupNonPerpetualGauge(owner, now, DayDuration)
	suite.setupNonPerpetualGauge(owner, now, SevenDaysDuration)
	suite.setupNonPerpetualGauge(owner, now, FourteenDaysDuration)

	suite.nextBlock(&now, &height)

	//lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, FourteenDaysDuration)

	//unlock
	suite.app.LockupKeeper.BeginUnlockPeriodLockByID(suite.ctx, 1)

	reward := sdk.Coins{}
	for i := 2; i <= 15; i++ {
		suite.nextEpoch(&now, &height)
		currentReward1Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, DayDuration)
		currentReward7Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, SevenDaysDuration)
		currentReward14Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, FourteenDaysDuration)
		if i >= 15 {
			suite.Require().Equal(sdk.Coins(nil), currentReward1Day.Rewards)
		} else {
			reward = reward.Add(sdk.NewCoin("stake", sdk.NewInt(100)))
		}
		if i >= 9 {
			suite.Require().Equal(sdk.Coins(nil), currentReward7Day.Rewards)
		} else {
			reward = reward.Add(sdk.NewCoin("stake", sdk.NewInt(100)))
		}
		if i >= 2 {
			suite.Require().Equal(sdk.Coins(nil), currentReward14Day.Rewards)
		} else {
			reward = reward.Add(sdk.NewCoin("stake", sdk.NewInt(100)))
		}

		lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, 1)
		estLockReward, err := suite.app.IncentivesKeeper.EstimateLockReward(suite.ctx, *lock)
		suite.Require().NoError(err)
		suite.Require().Equal(reward, estLockReward.Rewards)
	}

	//moving onto next epoch should completely finish unlocking and claim rewards
	suite.nextEpoch(&now, &height)

	lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, 1)
	suite.Require().Equal((*lockuptypes.PeriodLock)(nil), lock)

	stakeBalance := suite.app.BankKeeper.GetBalance(suite.ctx, owner, "stake")
	suite.Require().Equal(reward.AmountOf("stake"), stakeBalance.Amount)
}

func (suite *KeeperTestSuite) TestMultipleLocksFromMultipleUsers() {
	now, height := suite.setupEpochAndLockableDurations()
	suite.nextEpoch(&now, &height)

	//new non-perpetual gauge
	owner := sdk.AccAddress(OwnerAddr)
	denom, _ := suite.setupNonPerpetualGauge(owner, now, DayDuration)
	/* no 7day gauge */
	/* suite.setupNonPerpetualGauge(owner, now, SevenDaysDuration) */
	suite.setupNonPerpetualGauge(owner, now, FourteenDaysDuration)

	suite.nextBlock(&now, &height)

	//lock: user1=10lptoken,1day; user2=10lptoken,14days
	user1 := sdk.AccAddress(User1Addr)
	user2 := sdk.AccAddress(User2Addr)
	user1Stake := sdk.NewInt(10)
	user2Stake := sdk.NewInt(10)
	totalStake := user1Stake.Add(user2Stake)
	suite.LockTokens(user1, sdk.Coins{sdk.NewInt64Coin("lptoken", user1Stake.Int64())}, DayDuration)
	suite.LockTokens(user2, sdk.Coins{sdk.NewInt64Coin("lptoken", user2Stake.Int64())}, FourteenDaysDuration)
	for i := 0; i < 3; i++ {
		suite.nextEpoch(&now, &height)
		currentReward1Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, DayDuration)
		currentReward14Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, FourteenDaysDuration)
		suite.Require().Equal(user1Stake.Add(user2Stake), currentReward1Day.Coin.Amount)
		suite.Require().Equal(user2Stake, currentReward14Day.Coin.Amount)

		user1Lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, 1)
		lock1Reward, _ := suite.app.IncentivesKeeper.EstimateLockReward(suite.ctx, *user1Lock)
		user2Lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, 2)
		lock2Reward, _ := suite.app.IncentivesKeeper.EstimateLockReward(suite.ctx, *user2Lock)
		user1Reward := currentReward1Day.Rewards.AmountOf("stake").Mul(user1Stake).Quo(totalStake)
		user2Reward := currentReward1Day.Rewards.AmountOf("stake").Mul(user2Stake).Quo(totalStake)
		user2Reward = user2Reward.Add(currentReward14Day.Rewards.AmountOf("stake"))
		suite.Require().Equal(user1Reward, lock1Reward.Rewards.AmountOf("stake"))
		suite.Require().Equal(user2Reward, lock2Reward.Rewards.AmountOf("stake"))
	}

	//unlock: user1=10lptoken,1day
	suite.app.LockupKeeper.BeginUnlockPeriodLockByID(suite.ctx, 1)
	for i := 0; i < 3; i++ {
		suite.nextEpoch(&now, &height)
		currentReward1Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, DayDuration)
		currentReward14Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, FourteenDaysDuration)
		suite.Require().Equal(user2Stake, currentReward1Day.Coin.Amount)
		suite.Require().Equal(user2Stake, currentReward14Day.Coin.Amount)
	}

	//unlock: user2=10lptoken,14days
	suite.app.LockupKeeper.BeginUnlockPeriodLockByID(suite.ctx, 2)
	for i := 0; i < 13; i++ {
		suite.nextEpoch(&now, &height)
		currentReward1Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, DayDuration)
		currentReward14Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, FourteenDaysDuration)
		suite.Require().Equal(user2Stake, currentReward1Day.Coin.Amount)
		suite.Require().Equal(sdk.Coins(nil), currentReward14Day.Rewards)
	}

	//estimate and claim reward
	user2Lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, 2)
	user2Reward, err := suite.app.IncentivesKeeper.EstimateLockReward(suite.ctx, *user2Lock)
	suite.Require().NoError(err)
	claimedReward, err := suite.app.IncentivesKeeper.ClaimLockReward(suite.ctx, *user2Lock, owner)
	suite.Require().NoError(err)
	suite.Require().Equal(user2Reward.Rewards, claimedReward)

	suite.nextEpoch(&now, &height) //move to last epoch and finish unlocking lock2

	user1Bal := suite.app.BankKeeper.GetAllBalances(suite.ctx, user1)
	user2Bal := suite.app.BankKeeper.GetAllBalances(suite.ctx, user2)
	suite.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin("lptoken", 10), sdk.NewInt64Coin("stake", 150)), user1Bal)
	suite.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin("lptoken", 10), sdk.NewInt64Coin("stake", 1850)), user2Bal)
}
