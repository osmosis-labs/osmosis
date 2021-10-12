package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/epochs"
	"github.com/osmosis-labs/osmosis/x/epochs/types"
	"time"
)

var (
	AnEpochDuration = time.Hour * 24 * 7
	OwnerAddr       = "addr1---------------"
	User1Addr       = "user1---------------"
)

func (suite *KeeperTestSuite) TestF1Distribute() {
	now, height := suite.setupEpochAndLockableDurations()

	//next epoch
	now, height = suite.nextEpoch(now, height)

	//new gauge
	defaultGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: AnEpochDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin("stake", 1000)},
	}
	gauges := suite.SetupGauges([]perpGaugeDesc{defaultGauge})
	initGauge := gauges[0]
	denom := initGauge.DistributeTo.Denom
	duration := initGauge.DistributeTo.Duration
	owner := sdk.AccAddress(OwnerAddr)

	//next epoch
	now, height = suite.nextEpoch(now, height)

	//1st lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, AnEpochDuration)

	//next epoch
	now, height = suite.nextEpoch(now, height)

	currentReward, err := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, duration)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt64Coin("lptoken", 10), currentReward.Coin)
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 1000)}, currentReward.Rewards)
	suite.T().Logf("current_reward=%v", currentReward)

	//2nd lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 40)}, AnEpochDuration)

	//next epoch
	now, height = suite.nextEpoch(now, height)

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

	now := time.Now()
	height := int64(1)
	suite.ctx = suite.ctx.WithBlockHeight(height).WithBlockTime(now)

	epochs.InitGenesis(suite.ctx, suite.app.EpochsKeeper, types.GenesisState{
		Epochs: []types.EpochInfo{
			{
				Identifier:            "week",
				StartTime:             now.Add(AnEpochDuration),
				Duration:              AnEpochDuration,
				CurrentEpoch:          0,
				CurrentEpochStartTime: time.Time{},
				EpochCountingStarted:  false,
			},
		},
	})

	// add additional lockable durations
	suite.app.IncentivesKeeper.SetLockableDurations(suite.ctx,
		[]time.Duration{
			time.Hour * 24 * 7,
			time.Hour * 24 * 14,
			time.Hour * 24 * 21})

	return now, height
}

func (suite *KeeperTestSuite) nextEpoch(now time.Time, height int64) (time.Time, int64) {
	now = now.Add(AnEpochDuration).Add(time.Second * 10)
	height = height + 1
	suite.ctx = suite.ctx.WithBlockHeight(height).WithBlockTime(now)
	epochs.BeginBlocker(suite.ctx, suite.app.EpochsKeeper)
	return now, height
}
