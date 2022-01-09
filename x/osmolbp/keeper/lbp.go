package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
)

// TODO: verify if this is enough!
var multiplayer = sdk.NewInt(2 << 61)

// Returns the round number since lbp `start`.
// If now < start  return 0.
// If now == start return 0.
// if now == start + ROUND return 1...
// if now > end return round the round at end.
// distribution happens at the beginning of each round
func currentRound(p *api.LBP, now time.Time) uint64 {
	if now.Before(p.StartTime) {
		return 0
	}
	if !p.EndTime.After(now) { // end <= now
		now = p.EndTime
	}
	return uint64(now.Sub(p.StartTime) / api.ROUND)
}

func lbpRemainigBalance(p *api.LBP, userShares sdk.Int) sdk.Int {
	if userShares.IsZero() {
		return sdk.ZeroInt()
	}
	return p.Staked.Mul(userShares).Quo(p.Shares)
}

// compute amount of shares that should be minted for a new subscription amount
// TODO: caller must assert that the sale didn't finish:
//     inRemaining >0 and not ended
func computeSharesAmount(p *api.LBP, amountIn sdk.Int, roundUp bool) sdk.Int {
	if p.Shares.IsZero() {
		return amountIn
	}
	shares := amountIn.Mul(p.Shares)
	if roundUp {
		shares = shares.Add(p.Staked).AddRaw(-1).Quo(p.Staked)
	} else {
		shares = shares.Quo(p.Staked)
	}
	return shares
}

func lbpHasEnded(p *api.LBP, round uint64) bool {
	return p.EndRound >= round
}

func subscribe(p *api.LBP, u *api.UserPosition, amount sdk.Int) {
	triggerUserPurchase(p, u)
	remaining := lbpRemainigBalance(p, u.Shares)
	u.SpentInWithoutShares = u.SpentInWithoutShares.Add(u.Staked).Sub(remaining)
	shares := computeSharesAmount(p, amount, false)
	u.Shares = u.Shares.Add(shares)
	p.Shares = p.Shares.Add(shares)
	p.Staked = p.Staked.Add(amount)

	u.Staked = lbpRemainigBalance(p, u.Shares)
}

// TODO: maybe we can merge it with pingUser?
func triggerUserPurchase(p *api.LBP, u *api.UserPosition) {
	purchased := pingUser(u, p.OutPerShare)
	if purchased.IsPositive() {
		u.Purchased = u.Purchased.Add(purchased)
	}
	if u.Shares.IsPositive() {
		if lbpRemainigBalance(p, u.Shares).IsZero() {
			p.Shares = p.Shares.Sub(u.Shares)
			u.Shares = sdk.ZeroInt()
		}
	}
}

func withdraw(p *api.LBP, u *api.UserPosition, amount *sdk.Int, now time.Time) error {
	triggerUserPurchase(p, u)
	remaining := lbpRemainigBalance(p, u.Shares)
	if amount == nil {
		*amount = remaining
	} else if remaining.GT(*amount) {
		return errors.ErrInvalidRequest.Wrapf("Not enough balance, available balance: %s", remaining)
	}

	shares := computeSharesAmount(p, *amount, true)
	u.SpentInWithoutShares = u.SpentInWithoutShares.Add(u.Staked).Sub(remaining)
	u.Shares = u.Shares.Sub(shares)
	p.Shares = p.Shares.Sub(shares)
	p.Staked = p.Staked.Sub(*amount)

	return nil
}

// pingUser updates purchase rate and returns amount of tokens user purchased since the last time.
// `rate` is the current LBP per share distribution rate
func pingUser(u *api.UserPosition, ratePerShare sdk.Int) sdk.Int {
	out := sdk.ZeroInt()
	if !ratePerShare.IsZero() {
		diff := ratePerShare.Sub(u.Rate)
		out = diff.Mul(u.Shares).Quo(multiplayer)
	}
	u.Rate = ratePerShare
	return out
}

// TODO: rename to: finalize LBP - will send paid tokens to the LBP treasury
// This p.InPaidUnclaimed should be merged with p.InPaid
func distributeUnclaimedTokens(p *api.LBP, u *api.UserPosition) {
	// TODO:
	// 1. only after sale ends
	// 2. send tokens to the treasury / owner
	// 3. merge
}
