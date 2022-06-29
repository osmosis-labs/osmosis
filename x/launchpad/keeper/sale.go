package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v7/x/launchpad/types"
)

// TODO: verify if this is enough!
var multiplayer = sdk.NewInt(1_000_000) // sdk.NewInt(2 << 61)

// Returns the round number since sale `start`.
// if now < start  return 0.
// If now == start return 0.
// if now == start + ROUND return 1
// if now == start + 1.5*ROUND return 1...
// if now == start + 2*ROUND return 2...
// if now > end return the end_round.
// distribution happens at the beginning of each round. Last distribution is at end_round
func currentRound(start, end, now time.Time) int64 {
	if now.Before(start) {
		return 0
	}
	if !end.After(now) { // !(end>now) => end<=now
		now = end
	}
	return int64(now.Sub(start) / types.ROUND)
}

func saleRemainigBalance(p *types.Sale, userShares sdk.Int) sdk.Int {
	if userShares.IsZero() {
		return sdk.ZeroInt()
	}
	return p.Staked.Mul(userShares).Quo(p.Shares)
}

// compute amount of shares that should be minted for a new subscription amount
// TODO: caller must assert that the sale didn't finish:
//     inRemaining >0 and not ended
func computeSharesAmount(p *types.Sale, amountIn sdk.Int, roundUp bool) sdk.Int {
	if p.Shares.IsZero() || amountIn.IsZero() {
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

func saleHasEnded(p *types.Sale, round int64) bool {
	return p.EndRound >= round
}

func subscribe(p *types.Sale, u *types.UserPosition, amount sdk.Int, now time.Time) {
	pingSale(p, now)
	remaining := triggerUserPurchase(p, u)
	u.Spent = u.Spent.Add(u.Staked).Sub(remaining)
	shares := computeSharesAmount(p, amount, false)
	u.Shares = u.Shares.Add(shares)
	p.Shares = p.Shares.Add(shares)
	p.Staked = p.Staked.Add(amount)
	u.Staked = saleRemainigBalance(p, u.Shares)
}

// withdraw applies withdraw requests and updates sell state.
// If amount == nil then it withdrawns all the remaining deposit.
// Returns withdrawn amount.
func withdraw(p *types.Sale, u *types.UserPosition, amount *sdk.Int, now time.Time) (sdk.Int, error) {
	pingSale(p, now)
	remaining := triggerUserPurchase(p, u)
	if amount == nil {
		amount = &remaining
	} else if amount.GT(remaining) {
		return sdk.ZeroInt(), errors.ErrInvalidRequest.Wrapf("Not enough balance, available balance: %s", remaining)
	}

	shares := computeSharesAmount(p, *amount, true)
	u.Spent = u.Spent.Add(u.Staked).Sub(remaining)
	u.Shares = u.Shares.Sub(shares)
	p.Shares = p.Shares.Sub(shares)
	p.Staked = p.Staked.Sub(*amount)

	return *amount, nil
}

func pingSale(p *types.Sale, now time.Time) {
	// Need to use round for the end check to assure we have the final distribution
	if now.Before(p.StartTime) || p.Round >= p.EndRound {
		return
	}

	round := currentRound(p.StartTime, p.EndTime, now)
	diff := round - p.Round
	if p.Shares.IsZero() || diff == 0 {
		p.Round = round
		return
	}
	// remaining rounds including the current round
	remainingRounds := p.EndRound - p.Round
	// fmt.Println("remaining rounds:", remainingRounds, " p.round:", p.Round, " c_round:", round)
	p.Round = round
	if remainingRounds == 0 {
		return
	}

	sold := p.OutRemaining.MulRaw(diff).QuoRaw(remainingRounds)
	if sold.IsPositive() {
		p.OutSold = p.OutSold.Add(sold)
		p.OutRemaining = p.OutRemaining.Sub(sold)

		perShareDiff := sold.Mul(multiplayer).Quo(p.Shares)
		p.OutPerShare = p.OutPerShare.Add(perShareDiff)
	}
	income := p.Staked.MulRaw(diff).QuoRaw(remainingRounds)
	p.Income = p.Income.Add(income)
	p.Staked = p.Staked.Sub(income)
}

// returns remaining user token_in balance
func triggerUserPurchase(p *types.Sale, u *types.UserPosition) sdk.Int {
	// TODO: reorder and optimize - we can early return
	if !p.OutPerShare.IsZero() && !u.Shares.IsZero() {
		diff := p.OutPerShare.Sub(u.OutPerShare)
		if !diff.IsZero() {
			purchased := diff.Mul(u.Shares).Quo(multiplayer)
			// fmt.Printf("p.OutPerShare=%s   u.Shares=%s,  diff=%s, purchased=%s\n",
			// 	p.OutPerShare, u.Shares, diff, purchased)
			u.Purchased = u.Purchased.Add(purchased)
		}
	}
	u.OutPerShare = p.OutPerShare
	remaining := saleRemainigBalance(p, u.Shares)
	if u.Shares.IsPositive() {
		if remaining.IsZero() {
			p.Shares = p.Shares.Sub(u.Shares)
			u.Shares = sdk.ZeroInt()
		}
	}
	// we can't compute spent amount here because of the way how  we aggregate

	return remaining
}
