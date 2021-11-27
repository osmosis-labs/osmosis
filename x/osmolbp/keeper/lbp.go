package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/x/osmolbp/proto"
)

// Returns the round number since lbp `start`.
// If now < start  return 0.
// If now == start return 0.
// if now == start + ROUND return 1...
// if now > end return round the round at end.
func currentRound(p *proto.LBP, now time.Time) uint64 {
	if now.Before(p.Start) {
		return 0
	}
	if !p.End.After(now) { // end <= now
		now = p.End
		// NOTE: add adjustment if round
	}
	return uint64(now.Sub(p.End) / proto.ROUND)
}

// pingPool updates the accumulators based on the current round
func pingPool(p *proto.LBP, round uint64) {
	// check if we started or we already pinged in the same round
	// note: we don't update the accumulated values based on the updates in the current round.
	if p.AccumulatorR == round || round <= 0 {
		return
	}
	// fast forward. Also catch an edge case when the pool started, but doesn't have a stake.
	if p.Staked.IsZero() {
		p.AccumulatorR = round
		return
	}
	// amount of tokens sold between the rounds per unit of the stake
	// TODO: factor by an overflow to avoid zero
	diff := sdk.NewIntFromUint64(round - p.AccumulatorR).Mul(p.Rate).Quo(p.Staked)
	if !diff.IsZero() {
		p.Accumulator = p.Accumulator.Add(diff)
		p.AccumulatorR = round
	}
}

// pingUser updates user purchase based on the pool accumulator in the current round.
func pingUser(u *proto.UserPosition, round uint64, accumulator sdk.Int) {
	// return if we didn't started or we already updated based on the current accumulator
	if round <= 0 || u.Accumulator.GTE(accumulator) {
		return
	}
	// TODO: factor by an overflow
	purchased := u.Staked.Mul(accumulator.Sub(u.Accumulator))
	u.Purchased = u.Purchased.Add(purchased)
	u.Accumulator = accumulator
}

func stakeInPool(p *proto.LBP, u *proto.UserPosition, amount sdk.Int, now time.Time) {
	round := currentRound(p, now)
	pingPool(p, round)
	pingUser(u, round, p.Accumulator)

	// user stake will only be accounted in the next round
	u.Staked = u.Staked.Add(amount)
	p.Staked = p.Staked.Add(amount)
}

func unstakeFromPool(p *proto.LBP, u *proto.UserPosition, amount sdk.Int, now time.Time) error {
	round := currentRound(p, now)
	pingPool(p, round)
	pingUser(u, round, p.Accumulator)

	if amount.GT(u.Staked) {
		return errors.Wrapf(errors.ErrInvalidRequest, "unstake amount (%v) must not be bigger than available stake (%v)", amount, u.Staked)
	}

	// user stake will only be accounted in the next round
	// TODO: need to decrease the user stake and move it to the pool treasury!
	u.Staked = u.Staked.Sub(amount)
	p.Staked = p.Staked.Sub(amount)
	return nil
}
