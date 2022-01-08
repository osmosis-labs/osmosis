package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
)

// Returns the round number since lbp `start`.
// If now < start  return 0.
// If now == start return 0.
// if now == start + ROUND return 1...
// if now > end return round the round at end.
func currentRound(p *api.LBP, now time.Time) uint64 {
	if now.Before(p.StartTime) {
		return 0
	}
	if !p.EndTime.After(now) { // end <= now
		now = p.EndTime
		// NOTE: add adjustment if round
	}
	return uint64(now.Sub(p.EndTime) / api.ROUND)
}

// pingPool updates the accumulators based on the current round
func pingPool(p *api.LBP, round uint64) {
	// check if we started or we already pinged in the same round
	// note: we don't update the accumulated values based on the updates in the current round.
	if p.Round == round || round <= 0 {
		return
	}
	roundDiff := sdk.NewIntFromUint64(round - p.Round)
	if !p.IncomeRate.IsZero() {
		p.Income = p.Income.Add(roundDiff.Mul(p.IncomeRate))
	}
	// fast forward. Also catch an edge case when the pool started, but doesn't have a stake.
	if p.Staked.IsZero() {
		p.Round = round
		return
	}
	// amount of tokens sold between the rounds per unit of the stake
	// TODO: factor by an overflow to avoid zero
	diff := roundDiff.Mul(p.Rate).Quo(p.Staked)
	if !diff.IsZero() {
		p.AccumulatorOut = p.AccumulatorOut.Add(diff)
		p.Round = round
	}
	return
}

// pingUser updates user purchase based on the pool accumulator in the current round.
func pingUser(u *api.UserPosition, round uint64, accumulator sdk.Int) {
	// return if we didn't started or we already updated based on the current accumulator
	if round <= 0 || u.Accumulator.GTE(accumulator) {
		return
	}
	// TODO: factor by an overflow
	purchased := u.Staked.Mul(accumulator.Sub(u.Accumulator))
	u.Purchased = u.Purchased.Add(purchased)
	u.Accumulator = accumulator
	// TODO: need to decrease the user stake and move it to the pool treasury!
}

func stakeInPool(p *api.LBP, u *api.UserPosition, amount sdk.Int, now time.Time) {
	round := currentRound(p, now)
	// TODO: maybe we should return error?
	if round >= p.EndRound {
		return
	}
	pingPool(p, round)
	pingUser(u, round, p.AccumulatorOut)

	// user stake will only be accounted in the next round
	u.Staked = u.Staked.Add(amount)
	p.Staked = p.Staked.Add(amount)
	remainingRounds := sdk.NewIntFromUint64(p.EndRound - round)
	p.IncomeRate.Add(amount.Quo(remainingRounds))
}

func unstakeFromPool(p *api.LBP, u *api.UserPosition, amount sdk.Int, now time.Time) error {
	round := currentRound(p, now)
	pingPool(p, round)
	pingUser(u, round, p.AccumulatorOut)

	if amount.GT(u.Staked) {
		return errors.Wrapf(errors.ErrInvalidRequest, "unstake amount (%v) must not be bigger than available stake (%v)", amount, u.Staked)
	}

	// user stake will only be accounted in the next round
	u.Staked = u.Staked.Sub(amount)
	p.Staked = p.Staked.Sub(amount)
	remainingRounds := sdk.NewIntFromUint64(p.EndRound - round)
	p.IncomeRate.Sub(amount.Quo(remainingRounds))
	return nil
}
