package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochtypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

func (suite *KeeperTestSuite) TestCalculateHistoricalRewards() {
	lastProcessedEpoch := int64(1)
	totalStake := sdk.NewCoin("stake", sdk.NewInt(1000))
	rewardCoin := sdk.NewCoin("reward", sdk.NewInt(10000))
	totalReward := sdk.NewCoins(rewardCoin)
	k := suite.app.IncentivesKeeper
	duration := k.GetLockableDurations(suite.ctx)[0]
	epochInfo := k.GetEpochInfo(suite.ctx)
	currentReward := types.CurrentReward{
		LastProcessedEpoch: lastProcessedEpoch,
		TotalShares:        totalStake,
		Rewards:            totalReward,
	}

	prevCummulativeRewardRatio := sdk.NewDecCoin("reward", sdk.NewInt(1000))
	prevCumulativeRewardRatioCoins := sdk.NewDecCoins(prevCummulativeRewardRatio)
	k.SetHistoricalReward(suite.ctx, prevCumulativeRewardRatioCoins, "stake", duration, lastProcessedEpoch)

	expectedCummulativeReward := sdk.NewDecCoin("reward", sdk.NewInt(1000+10))
	expectedHistoricalRewardCumulativeRewardRatio := sdk.NewDecCoins(expectedCummulativeReward)

	resultHistoricalRewardCumulativeRewardRatio, err := k.CalculateCumulativeRewardRatio(suite.ctx, currentReward, "stake", duration, epochInfo.CurrentEpoch)
	suite.Require().NoError(err)
	suite.Require().Equal(resultHistoricalRewardCumulativeRewardRatio, expectedHistoricalRewardCumulativeRewardRatio)

	resultAmount := prevCummulativeRewardRatio.Amount.Add(rewardCoin.Amount.ToDec().Quo(totalStake.Amount.ToDec()))
	suite.Require().Equal(resultHistoricalRewardCumulativeRewardRatio, sdk.NewDecCoins(sdk.NewDecCoinFromDec("reward", resultAmount)))
}

func (suite *KeeperTestSuite) TestCalculateRewardBetweenPeriod() {
	k := suite.app.IncentivesKeeper
	duration := k.GetLockableDurations(suite.ctx)[0]

	prevCummulativeRewardRatio := sdk.NewDecCoin("reward", sdk.NewInt(1000))
	prevCumulativeRewardRatioCoins := sdk.NewDecCoins(prevCummulativeRewardRatio)
	k.SetHistoricalReward(suite.ctx, prevCumulativeRewardRatioCoins, "stake", duration, 1)

	currCummulativeReward := sdk.NewDecCoin("reward", sdk.NewInt(2000))
	currCumulativeRewardRatioCoins := sdk.NewDecCoins(currCummulativeReward)
	k.SetHistoricalReward(suite.ctx, currCumulativeRewardRatioCoins, "stake", duration, 100)

	numStake := sdk.NewInt(100)
	resultCoins, err := k.CalculateRewardBetweenEpoch(suite.ctx, "stake", duration, numStake, 1, 100)
	suite.Require().NoError(err)
	// expectedAmt := currCummulativeReward.Amount.Sub(prevCummulativeReward.Amount).MulInt(numStake).TruncateInt()
	expectedAmt := sdk.NewInt((2000 - 1000) * 100)
	expectedCoin := sdk.NewCoins(sdk.NewCoin("reward", expectedAmt))
	suite.Require().Equal(expectedCoin, resultCoins)
}

func (suite *KeeperTestSuite) TestCalculateRewardForLock() {
	k := suite.app.IncentivesKeeper

	numStake := sdk.NewInt(100)
	lockedCoin := sdk.NewCoin("stake", numStake)
	lockedCoins := sdk.NewCoins(lockedCoin)

	duration := k.GetLockableDurations(suite.ctx)[0]

	lock := lockuptypes.PeriodLock{
		ID:    1,
		Coins: lockedCoins,
	}

	lockReward := types.PeriodLockReward{
		LockId:             lock.ID,
		LastEligibleEpochs: []*types.LastEligibleEpochByDurationAndDenom{},
		Rewards:            sdk.Coins{},
	}

	prevCummulativeRewardRatio := sdk.NewDecCoin("reward", sdk.NewInt(1000))
	prevCumulativeRewardRatioCoins := sdk.NewDecCoins(prevCummulativeRewardRatio)
	k.SetHistoricalReward(suite.ctx, prevCumulativeRewardRatioCoins, "stake", duration, 1)
	lastEligibleEpoch := types.LastEligibleEpochByDurationAndDenom{
		LockDuration: duration,
		Denom:        "stake",
		Epoch:        1,
	}
	lockReward.LastEligibleEpochs = append(lockReward.LastEligibleEpochs, &lastEligibleEpoch)

	currCummulativeReward := sdk.NewDecCoin("reward", sdk.NewInt(2000))
	currCumulativeRewardRatioCoins := sdk.NewDecCoins(currCummulativeReward)
	k.SetHistoricalReward(suite.ctx, currCumulativeRewardRatioCoins, "stake", duration, 100)
	expectedAmount := sdk.NewInt((2000 - 1000) * 100)
	expectedRewards := sdk.NewCoins(sdk.NewCoin("reward", expectedAmount))

	epochInfo := epochtypes.EpochInfo{
		CurrentEpoch: 101,
	}

	currentReward := types.CurrentReward{
		LastProcessedEpoch: 100,
		TotalShares:        lockedCoin,
	}

	k.SetCurrentReward(suite.ctx, currentReward, "stake", duration)

	lockReward, err := k.CalculateRewardForLock(suite.ctx, lock, lockReward, epochInfo, duration, false)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedRewards, lockReward.Rewards)

	k.SetCurrentReward(suite.ctx, currentReward, "stake", duration)
}
