package keeper

import (
	"cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"

	"github.com/osmosis-labs/osmosis/v27/x/oracle/types"
)

// Tally calculates the median and returns it. Sets the set of voters to be rewarded, i.e. voted within
// a reasonable spread from the weighted median to the store
// CONTRACT: pb must be sorted
func Tally(pb types.ExchangeRateBallot, rewardBand osmomath.Dec, validatorClaimMap map[string]types.Claim) (weightedMedian osmomath.Dec) {
	weightedMedian = pb.WeightedMedian()
	standardDeviation := pb.StandardDeviation(weightedMedian)
	rewardSpread := weightedMedian.Mul(rewardBand.QuoInt64(2))

	if standardDeviation.GT(rewardSpread) {
		rewardSpread = standardDeviation
	}

	for _, vote := range pb {
		// Filter ballot winners & abstain voters
		if (vote.ExchangeRate.GTE(weightedMedian.Sub(rewardSpread)) &&
			vote.ExchangeRate.LTE(weightedMedian.Add(rewardSpread))) ||
			!vote.ExchangeRate.IsPositive() {
			key := vote.Voter.String()
			claim := validatorClaimMap[key]
			claim.Weight += vote.Power
			claim.WinCount++
			validatorClaimMap[key] = claim
		}
	}

	return weightedMedian
}

// ballot for the asset is passing the threshold amount of voting power
func ballotIsPassing(ballot types.ExchangeRateBallot, thresholdVotes math.Int) (math.Int, bool) {
	ballotPower := osmomath.NewInt(ballot.Power())
	return ballotPower, !ballotPower.IsZero() && ballotPower.GTE(thresholdVotes)
}

// PickReferenceSymphony choose Reference Symphony with the highest voter turnout
// If the voting power of the two denominations is the same,
// select reference Symphony in alphabetical order.
func PickReferenceSymphony(ctx sdk.Context, k Keeper, voteTargets map[string]osmomath.Dec, voteMap map[string]types.ExchangeRateBallot) (string, error) {
	largestBallotPower := int64(0)
	referenceSymphony := ""

	totalBondedTokens, err := k.StakingKeeper.TotalBondedTokens(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get total bonded tokens: %w", err)
	}
	totalBondedPower := sdk.TokensToConsensusPower(totalBondedTokens, k.StakingKeeper.PowerReduction(ctx))
	voteThreshold := k.VoteThreshold(ctx)
	thresholdVotes := voteThreshold.MulInt64(totalBondedPower).RoundInt()

	for denom, ballot := range voteMap {
		// If denom is not in the voteTargets, or the ballot for it has failed, then skip
		// and remove it from voteMap for iteration efficiency
		if _, exists := voteTargets[denom]; !exists {
			delete(voteMap, denom)
			continue
		}

		ballotPower := int64(0)

		// If the ballot is not passed, remove it from the voteTargets array
		// to prevent slashing validators who did valid vote.
		if power, ok := ballotIsPassing(ballot, thresholdVotes); ok {
			ballotPower = power.Int64()
		} else {
			delete(voteTargets, denom)
			delete(voteMap, denom)
			continue
		}

		if ballotPower > largestBallotPower || largestBallotPower == 0 {
			referenceSymphony = denom
			largestBallotPower = ballotPower
		} else if largestBallotPower == ballotPower && referenceSymphony > denom {
			referenceSymphony = denom
		}
	}

	return referenceSymphony, nil
}
