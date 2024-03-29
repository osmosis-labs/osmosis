package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/osmosis-labs/osmosis/v23/app/params"

	"github.com/osmosis-labs/osmosis/v23/x/oracle/types"
)

// RewardBallotWinners implements
// at the end of every VotePeriod, give out a portion of spread fees collected in the oracle reward pool
//
//	to the oracle voters that voted faithfully.
func (k Keeper) RewardBallotWinners(
	ctx sdk.Context,
	votePeriod int64,
	rewardDistributionWindow int64,
	voteTargets map[string]sdk.Dec,
	ballotWinners map[string]types.Claim,
) {
	// Add Luna explicitly for oracle account balance coming from the market swap fee
	rewardDenoms := make([]string, len(voteTargets)+1)
	rewardDenoms[0] = appparams.BaseCoinUnit

	i := 1
	for denom := range voteTargets {
		rewardDenoms[i] = denom
		i++
	}

	// Sum weight of the claims
	ballotPowerSum := int64(0)
	for _, winner := range ballotWinners {
		ballotPowerSum += winner.Weight
	}

	// Exit if the ballot is empty
	if ballotPowerSum == 0 {
		return
	}

	// The Reward distributionRatio = votePeriod/rewardDistributionWindow
	distributionRatio := sdk.NewDec(votePeriod).QuoInt64(rewardDistributionWindow)

	var periodRewards sdk.DecCoins
	for _, denom := range rewardDenoms {
		rewardPool := k.GetRewardPool(ctx, denom)

		// return if there's no rewards to give out
		if rewardPool.IsZero() {
			continue
		}

		periodRewards = periodRewards.Add(sdk.NewDecCoinFromDec(
			denom,
			sdk.NewDecFromInt(rewardPool.Amount).Mul(distributionRatio),
		))
	}

	logger := k.Logger(ctx)
	logger.Debug("RewardBallotWinner", "periodRewards", periodRewards)

	// Dole out rewards
	var distributedReward sdk.Coins
	for _, winner := range ballotWinners {
		receiverVal := k.StakingKeeper.Validator(ctx, winner.Recipient)

		// Reflects contribution
		rewardCoins, _ := periodRewards.MulDec(sdk.NewDec(winner.Weight).QuoInt64(ballotPowerSum)).TruncateDecimal()

		// In case absence of the validator, we just skip distribution
		if receiverVal != nil && !rewardCoins.IsZero() {
			k.distrKeeper.AllocateTokensToValidator(ctx, receiverVal, sdk.NewDecCoinsFromCoins(rewardCoins...))
			distributedReward = distributedReward.Add(rewardCoins...)
		} else {
			logger.Debug(fmt.Sprintf("no reward %s(%s)",
				receiverVal.GetMoniker(),
				receiverVal.GetOperator().String()),
				"miss", k.GetMissCounter(ctx, receiverVal.GetOperator()),
				"wincount", winner.WinCount)
		}
	}

	// Move distributed reward to distribution module
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.distrName, distributedReward)
	if err != nil {
		panic(fmt.Sprintf("[oracle] Failed to send coins to distribution module %s", err.Error()))
	}
}
