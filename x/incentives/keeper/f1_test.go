package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/epochs"
	"github.com/osmosis-labs/osmosis/x/epochs/types"
	incentivestypes "github.com/osmosis-labs/osmosis/x/incentives/types"
	"time"
)

var (
	DayDuration = time.Hour * 24
	OwnerAddr   = "addr1---------------"
	User1Addr   = "user1---------------"
)

func (suite *KeeperTestSuite) TestF1Distribute() {
	now, height := suite.setupEpochAndLockableDurations()

	//make sure that now passed an epoch
	now = now.Add(time.Second)

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
	suite.T().Logf("current_reward=%v", currentReward)

	//2nd lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 40)}, DayDuration)

	//next epoch
	suite.nextEpoch(&now, &height)

	currentReward, err = suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, duration)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt64Coin("lptoken", 50), currentReward.Coin)
	suite.Require().Equal(sdk.Coins(nil), currentReward.Rewards)
	suite.T().Logf("current_reward=%v", currentReward)

	prevHistoricalReward, err := suite.app.IncentivesKeeper.GetHistoricalReward(suite.ctx, denom, duration, currentReward.Period-1)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt64DecCoin("stake", 100), prevHistoricalReward.CummulativeRewardRatio[0])
	suite.T().Logf("historical_reward=%v", prevHistoricalReward)
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
	*now = (*now).Add(DayDuration)
	*height = *height + 1
	suite.ctx = suite.ctx.WithBlockHeight(*height).WithBlockTime(*now)
	epochs.BeginBlocker(suite.ctx, suite.app.EpochsKeeper)
}

func (suite *KeeperTestSuite) nextBlock(now *time.Time, height *int64) {
	*now = (*now).Add(time.Second)
	*height = *height + 1
	suite.ctx = suite.ctx.WithBlockHeight(*height).WithBlockTime(*now)
	epochs.BeginBlocker(suite.ctx, suite.app.EpochsKeeper)
}
