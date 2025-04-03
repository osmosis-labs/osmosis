package keeper

import (
	"fmt"
	"github.com/osmosis-labs/osmosis/osmomath"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"

	"github.com/osmosis-labs/osmosis/v27/x/oracle/types"
)

// RewardBallotWinners implements
// at the end of every VotePeriod, give out a portion of spread fees collected in the oracle reward pool
//
//	to the oracle voters that voted faithfully.
func (k Keeper) RewardBallotWinners(
	ctx sdk.Context,
	votePeriod int64,
	rewardDistributionWindow int64,
	voteTargets map[string]osmomath.Dec,
	ballotWinners map[string]types.Claim,
) {
	// Add Note explicitly for oracle account balance coming from the market swap fee
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
	distributionRatio := osmomath.NewDec(votePeriod).QuoInt64(rewardDistributionWindow)

	var periodRewards sdk.DecCoins
	for _, denom := range rewardDenoms {
		rewardPool := k.GetRewardPool(ctx, denom)

		// return if there's no rewards to give out
		if rewardPool.IsZero() {
			continue
		}

		periodRewards = periodRewards.Add(sdk.NewDecCoinFromDec(
			denom,
			osmomath.NewDecFromInt(rewardPool.Amount).Mul(distributionRatio),
		))
	}

	logger := k.Logger(ctx)
	logger.Debug("RewardBallotWinner", "periodRewards", periodRewards)

	// Dole out rewards
	var distributedReward sdk.Coins
	for _, winner := range ballotWinners {

		// Reflects contribution
		rewardCoins, _ := periodRewards.MulDec(osmomath.NewDec(winner.Weight).QuoInt64(ballotPowerSum)).TruncateDecimal()

		receiverVal, err := k.StakingKeeper.GetValidator(ctx, winner.Recipient)
		// In case absence of the validator, we just skip distribution
		if err != nil && !rewardCoins.IsZero() {
			err = k.distrKeeper.AllocateTokensToValidator(ctx, receiverVal, sdk.NewDecCoinsFromCoins(rewardCoins...))
			if err != nil {
				panic("could not allocate tokens to validator")
			}
			distributedReward = distributedReward.Add(rewardCoins...)
		} else {
			valAddress, err := sdk.ValAddressFromBech32(receiverVal.GetOperator())
			if err != nil {
				valAddress = []byte{}
			}
			logger.Debug(fmt.Sprintf("no reward %s(%s)",
				receiverVal.GetMoniker(),
				receiverVal.GetOperator()),
				"miss", k.GetMissCounter(ctx, valAddress),
				"wincount", winner.WinCount)
		}
	}

	// Move distributed reward to distribution module
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.distrName, distributedReward)
	if err != nil {
		panic(fmt.Sprintf("[oracle] Failed to send coins to distribution module %s", err.Error()))
	}
}
