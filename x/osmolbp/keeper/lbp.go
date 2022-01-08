package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
)

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

func lbpRemainigBalance(p *api.LBP, shares sdk.Int) sdk.Int {
	if shares.IsZero() {
		return sdk.ZeroInt()
	}
	return p.InRemaining.Mul(shares).Quo(p.Shares)
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
		shares = shares.Add(p.InRemaining).AddRaw(-1).Quo(p.InRemaining)
	} else {
		shares = shares.Quo(p.InRemaining)
	}
	return shares
}

func lbpHasEnded(p *api.LBP, round uint64) bool {
	return p.EndRound >= round
}

func subscribe(p *api.LBP, u *api.UserPosition, amount sdk.Int, now time.Time) {
}

func updateSubscriptoin(p *api.LBP, u *api.UserPosition, amount sdk.Int, now time.Time) {

}

func pingUser(u *api.UserPosition, outPerShare sdk.Int) sdk.Int {
	return sdk.Int{}
}
