package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v11/x/streamswap/types"
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

func saleRemainigBalance(s *types.Sale, userShares sdk.Int) sdk.Int {
	if userShares.IsZero() {
		return sdk.ZeroInt()
	}
	return s.Staked.Mul(userShares).Quo(s.Shares)
}

// compute amount of shares that should be minted for a new subscription amount
func computeSharesAmount(s *types.Sale, amountIn sdk.Int, roundUp bool) sdk.Int {
	if s.Shares.IsZero() || amountIn.IsZero() {
		return amountIn
	}
	shares := amountIn.Mul(s.Shares)
	if roundUp {
		shares = shares.Add(s.Staked).AddRaw(-1).Quo(s.Staked)
	} else {
		shares = shares.Quo(s.Staked)
	}
	return shares
}

func saleHasEnded(s *types.Sale, round int64) bool {
	return s.EndRound >= round
}

func subscribe(s *types.Sale, u *types.UserPosition, amount sdk.Int, now time.Time) {
	pingSale(s, now)
	if s.Round >= s.EndRound {
		return
	}
	remaining := triggerUserPurchase(s, u)
	u.Spent = u.Spent.Add(u.Staked).Sub(remaining)
	shares := computeSharesAmount(s, amount, false)
	u.Shares = u.Shares.Add(shares)
	s.Shares = s.Shares.Add(shares)
	s.Staked = s.Staked.Add(amount)
	u.Staked = saleRemainigBalance(s, u.Shares)
}

// withdraw applies withdraw requests and updates sell state.
// If amount == nil then it withdrawns all the remaining deposit.
// Returns withdrawn amount.
func withdraw(s *types.Sale, u *types.UserPosition, amount *sdk.Int, now time.Time) (sdk.Int, error) {
	pingSale(s, now)
	remaining := triggerUserPurchase(s, u)
	if amount == nil {
		amount = &remaining
	} else if amount.GT(remaining) {
		return sdk.ZeroInt(), errors.ErrInvalidRequest.Wrapf("Not enough balance, available balance: %s", remaining)
	}

	shares := computeSharesAmount(s, *amount, true)
	u.Spent = u.Spent.Add(u.Staked).Sub(remaining)
	u.Shares = u.Shares.Sub(shares)
	s.Shares = s.Shares.Sub(shares)
	s.Staked = s.Staked.Sub(*amount)

	return *amount, nil
}

func pingSale(s *types.Sale, now time.Time) {
	// Need to use round for the end check to assure we have the final distribution
	round := currentRound(s.StartTime, s.EndTime, now)
	if now.Before(s.StartTime) || s.Round >= s.EndRound {
		return
	}

	diff := round - s.Round
	if s.Shares.IsZero() || diff == 0 {
		s.Round = round
		return
	}
	// remaining rounds including the current round
	remainingRounds := s.EndRound - s.Round
	// fmt.Println("remaining rounds:", remainingRounds, " p.round:", p.Round, " c_round:", round)
	if remainingRounds <= 0 {
		return
	}

	s.Round = round
	sold := s.OutRemaining.MulRaw(diff).QuoRaw(remainingRounds)
	if sold.IsPositive() {
		s.OutSold = s.OutSold.Add(sold)
		s.OutRemaining = s.OutRemaining.Sub(sold)

		perShareDiff := sold.Mul(multiplayer).Quo(s.Shares)
		s.OutPerShare = s.OutPerShare.Add(perShareDiff)
	}
	income := s.Staked.MulRaw(diff).QuoRaw(remainingRounds)
	s.Income = s.Income.Add(income)
	s.Staked = s.Staked.Sub(income)
}

// returns remaining user token_in balance
func triggerUserPurchase(s *types.Sale, u *types.UserPosition) sdk.Int {
	// TODO: reorder and optimize - we can early return
	if !s.OutPerShare.IsZero() && !u.Shares.IsZero() {
		diff := s.OutPerShare.Sub(u.OutPerShare)
		if !diff.IsZero() {
			purchased := diff.Mul(u.Shares).Quo(multiplayer)
			// fmt.Printf("p.OutPerShare=%s   u.Shares=%s,  diff=%s, purchased=%s\n",
			// 	p.OutPerShare, u.Shares, diff, purchased)
			u.Purchased = u.Purchased.Add(purchased)
		}
	}
	u.OutPerShare = s.OutPerShare
	remaining := saleRemainigBalance(s, u.Shares)
	if u.Shares.IsPositive() {
		if remaining.IsZero() {
			s.Shares = s.Shares.Sub(u.Shares)
			u.Shares = sdk.ZeroInt()
		}
	}
	// we can't compute spent amount here because of the way how  we aggregate

	return remaining
}
