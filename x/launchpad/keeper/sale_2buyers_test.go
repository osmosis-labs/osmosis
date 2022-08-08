package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/launchpad/types"
)

var zero = sdk.ZeroInt()

type TwoBuyersSuite struct {
	SaleSuite
	u1, u2 *types.UserPosition
	p      *types.Sale

	staked1, staked2, totalStaked          sdk.Int
	totalOut, inPerRound, outPerRound      sdk.Int
	u1PurchasePerRound, u2PurchasePerRound sdk.Int
}

func (s *TwoBuyersSuite) SetupTest() {
	s.p = s.createSale()
	u1, u2 := newUserPosition(), newUserPosition()
	s.u1, s.u2 = &u1, &u2
	s.staked1 = sdk.NewInt(20)
	s.staked2 = s.staked1.MulRaw(2)
	s.totalStaked = s.staked1.Add(s.staked2)
	s.totalOut = s.p.OutRemaining
	s.inPerRound = s.totalStaked.QuoRaw(10)
	s.outPerRound = s.totalOut.QuoRaw(10)
	s.u1PurchasePerRound = s.outPerRound.QuoRaw(3)
	s.u2PurchasePerRound = s.u1PurchasePerRound.MulRaw(2)
}

func (s *TwoBuyersSuite) Test2Buyers() {
	require := s.Require()
	log := s.T().Log

	subscribe(s.p, s.u1, s.staked1, s.before)
	checkSale(require, s.p, 0, s.totalOut, zero, zero, s.staked1, zero, s.staked1)
	subscribe(s.p, s.u2, s.staked2, s.before)

	checkUser(require, s.u1, s.staked1, s.staked1, zero, zero, "user1")
	checkUser(require, s.u2, s.staked2, s.staked2, zero, zero, "user2")

	// ping before start shouldn't change anything
	pingSale(s.p, s.before2)
	checkSale(require, s.p, 0, s.totalOut, zero, zero, s.totalStaked, zero, s.totalStaked)
	pingSale(s.p, s.before2.Add(types.ROUND))
	checkSale(require, s.p, 0, s.totalOut, zero, zero, s.totalStaked, zero, s.totalStaked)

	// ###############################################
	// at the start, the round is still zero and we don't do a sale yet
	pingSale(s.p, s.start)
	// NOTE: we don't test out per share
	//outPerShare := s.p.OutPerShare
	checkSale(require, s.p, 0, s.totalOut, zero, zero, s.totalStaked, zero, s.totalStaked)
	triggerUserPurchase(s.p, s.u1)
	// triggerUserPurchase shouldn't change Sale
	checkSale(require, s.p, 0, s.totalOut, zero, zero, s.totalStaked, zero, s.totalStaked)

	checkUser(require, s.u1, s.staked1, s.staked1, s.u1.OutPerShare, zero, "user1 zero round")

	// ###############################################
	// at the beginning of the first round (start + round) we should do the first sale
	now := s.start.Add(types.ROUND)
	pingSale(s.p, now)
	//checkSale(require, s.p, 1, s.totalOut.Sub(s.outPerRound), s.outPerRound, outPerShare, s.totalStaked.Sub(s.inPerRound), s.inPerRound, s.totalStaked)

	// second ping shouldn't change anything
	pingSale(s.p, now)
	checkSale(require, s.p, 1, s.totalOut.Sub(s.outPerRound), s.outPerRound, s.p.OutPerShare, s.totalStaked.Sub(s.inPerRound), s.inPerRound, s.totalStaked)

	// check user purchase
	log("\n### u1 triggers purchase in 1st round ###\n")
	triggerUserPurchase(s.p, s.u1)
	//// sale shouldn't change
	checkSale(require, s.p, 1, s.totalOut.Sub(s.outPerRound), s.outPerRound, s.u1.OutPerShare, s.totalStaked.Sub(s.inPerRound), s.inPerRound, s.totalStaked)
	checkUser(require, s.u1, s.staked1, s.staked1, s.u1.OutPerShare, s.u1PurchasePerRound, "user1 first round")
	// second purchase in the same round shouldn't make any effect
	triggerUserPurchase(s.p, s.u1)
	checkUser(require, s.u1, s.staked1, s.staked1, s.u1.OutPerShare, s.u1PurchasePerRound, "user1 first round")
	// second purchase in the middle of the first round shouldn't make any effect
	pingSale(s.p, now.Add(types.ROUND/2))
	triggerUserPurchase(s.p, s.u1)
	checkUser(require, s.u1, s.staked1, s.staked1, s.u1.OutPerShare, s.u1PurchasePerRound, "user1 first round")

	// ###############################################
	// round 3: user2 triggers a purchase
	now = s.start.Add(3 * types.ROUND)
	log("\n### u2 triggers purchase in 3rd round ###\n")
	u2PurchasePerRound := s.u1PurchasePerRound.MulRaw(2)
	pingSale(s.p, now)
	triggerUserPurchase(s.p, s.u2)
	checkSale(require, s.p, 3, s.totalOut.Sub(s.outPerRound.MulRaw(3)), s.outPerRound.MulRaw(3), s.u1.OutPerShare.MulRaw(3), s.totalStaked.Sub(s.inPerRound.MulRaw(3)), s.inPerRound.MulRaw(3), s.totalStaked)
	checkUser(require, s.u2, s.staked2, s.staked2, s.u2.OutPerShare, u2PurchasePerRound.MulRaw(3), "user2 3rd round")

	log("\n### u1 triggers purchase in 3rd round ###\n")
	pingSale(s.p, now)
	triggerUserPurchase(s.p, s.u1)
	// sale shouldn't change
	checkSale(require, s.p, 3, s.totalOut.Sub(s.outPerRound.MulRaw(3)), s.outPerRound.MulRaw(3), s.u1.OutPerShare, s.totalStaked.Sub(s.inPerRound.MulRaw(3)), s.inPerRound.MulRaw(3), s.totalStaked)
	checkUser(require, s.u1, s.staked1, s.staked1, s.u1.OutPerShare, s.u1PurchasePerRound.MulRaw(3), "user1 3rd round")

	// ###############################################
	// last by one round
	log("\n### u1 triggers purchase in the last round ###\n")
	now = s.end.Add(-types.ROUND)
	pingSale(s.p, now)
	triggerUserPurchase(s.p, s.u1)
	checkSale(require, s.p, 9, s.outPerRound, s.outPerRound.MulRaw(9), s.u1.OutPerShare, s.totalStaked.Sub(s.inPerRound.MulRaw(9)), s.inPerRound.MulRaw(9), s.totalStaked)
	checkUser(require, s.u1, s.staked1, s.staked1, s.u1.OutPerShare, s.u1PurchasePerRound.MulRaw(9), "user1 10th round")

	// ###############################################
	// Last round
	log("\n### u1 triggers purchase in the end ###\n")
	now = s.end
	pingSale(s.p, now)
	triggerUserPurchase(s.p, s.u1)
	// user 1 bough everything so p.Shares decerased by user shares
	checkSale(require, s.p, 10, zero, s.totalOut, s.u1.OutPerShare, zero, s.inPerRound.MulRaw(10), s.u2.Shares)
	checkUser(require, s.u1, zero, s.staked1, s.u1.OutPerShare, s.u1PurchasePerRound.MulRaw(10), "user1 10th round")

	// ###############################################
	// checking after the end - shouldn't make any effect
	now = s.after
	pingSale(s.p, now)
	triggerUserPurchase(s.p, s.u1)
	checkSale(require, s.p, 10, zero, s.totalOut, s.u1.OutPerShare, zero, s.inPerRound.MulRaw(10), s.u2.Shares)
	checkUser(require, s.u1, zero, s.staked1, s.u1.OutPerShare, s.u1PurchasePerRound.MulRaw(10), "user1 10th round")

	pingSale(s.p, now)
	triggerUserPurchase(s.p, s.u2)
	// only shares should change
	checkSale(require, s.p, 10, zero, s.totalOut, s.u1.OutPerShare, zero, s.inPerRound.MulRaw(10), zero)
	checkUser(require, s.u2, zero, s.staked2, s.u2.OutPerShare, u2PurchasePerRound.MulRaw(10), "user2 10th round")
}

func (s *TwoBuyersSuite) Test2BuyersEnd1() {
	require := s.Require()
	subscribe(s.p, s.u1, s.staked1, s.before)
	subscribe(s.p, s.u2, s.staked2, s.before)

	pingSale(s.p, s.end.Add(-types.ROUND)) // last by one purchase
	triggerUserPurchase(s.p, s.u1)
	checkUser(require, s.u1, s.staked1, s.staked1, s.u1.OutPerShare, s.u1PurchasePerRound.MulRaw(9), "user1 @ end")

	pingSale(s.p, s.end) // last purchase
	triggerUserPurchase(s.p, s.u1)
	checkUser(require, s.u1, zero, s.staked1, s.u1.OutPerShare, s.totalOut.QuoRaw(3), "user1 @ end")

	// after the  last purchase no change should be made
	pingSale(s.p, s.after)
	triggerUserPurchase(s.p, s.u1)
	checkUser(require, s.u1, zero, s.staked1, s.u1.OutPerShare, s.totalOut.QuoRaw(3), "user1 @ end")

	// withdraw everything - which is zero because we are at the end of the sale
	amount, err := withdraw(s.p, s.u1, nil, s.after)
	require.NoError(err)
	require.True(amount.IsZero())
}

func (s *TwoBuyersSuite) Test2BuyersEnd2() {
	require := s.Require()
	subscribe(s.p, s.u1, s.staked1, s.before)
	subscribe(s.p, s.u2, s.staked2, s.before)

	pingSale(s.p, s.end)
	triggerUserPurchase(s.p, s.u1)
	checkUser(require, s.u1, zero, s.staked1, s.u1.OutPerShare, s.totalOut.QuoRaw(3), "user1 @ end")
}

func (s *TwoBuyersSuite) Test2BuyersEnd3() {
	require := s.Require()
	subscribe(s.p, s.u1, s.staked1, s.before)
	subscribe(s.p, s.u2, s.staked2, s.before)

	pingSale(s.p, s.after)
	triggerUserPurchase(s.p, s.u1)
	checkUser(require, s.u1, zero, s.staked1, s.u1.OutPerShare, s.totalOut.QuoRaw(3), "user1 @ end")
}

func (s *TwoBuyersSuite) Test2BuyersEnd_mid1() {
	require := s.Require()
	end := s.end.Add(types.ROUND / 2) // half round after normal end
	s.p.EndRound = currentRound(s.start, end, end)
	subscribe(s.p, s.u1, s.staked1, s.before)
	subscribe(s.p, s.u2, s.staked2, s.before)

	pingSale(s.p, s.end.Add(-types.ROUND/2)) // last purchase still happens in the end_round
	triggerUserPurchase(s.p, s.u1)
	checkUser(require, s.u1, s.staked1, s.staked1, s.u1.OutPerShare, s.totalOut.QuoRaw(3).Sub(s.u1PurchasePerRound), "user1 @ end")

	pingSale(s.p, s.end) // round(s.end) == round(end) == round(s.end + round/2)
	triggerUserPurchase(s.p, s.u1)
	checkUser(require, s.u1, zero, s.staked1, s.u1.OutPerShare, s.totalOut.QuoRaw(3), "user1 @ end")

	pingSale(s.p, end)
	triggerUserPurchase(s.p, s.u1)
	checkUser(require, s.u1, zero, s.staked1, s.u1.OutPerShare, s.totalOut.QuoRaw(3), "user1 @ end")

	// withdraw everything - which is zero because we are at the end of the sale
	amount, err := withdraw(s.p, s.u1, nil, end)
	require.NoError(err)
	require.True(amount.IsZero())
}

// subscribe at the beginning and trigger purchase after the end
func (s *TwoBuyersSuite) Test2BuyersEnd_mid2() {
	require := s.Require()
	end := s.end.Add(types.ROUND / 2) // half round after normal end
	s.p.EndRound = currentRound(s.start, end, end)
	subscribe(s.p, s.u1, s.staked1, s.before)
	subscribe(s.p, s.u2, s.staked2, s.before)

	pingSale(s.p, s.after)
	triggerUserPurchase(s.p, s.u1)
	checkUser(require, s.u1, zero, s.staked1, s.u1.OutPerShare, s.totalOut.QuoRaw(3), "user1 @ end")
}

// subscribe before the beginning and withdraw user 1 after 2 rounds
func (s *TwoBuyersSuite) Test2Buyers_withdraw1() {
	require := s.Require()
	subscribe(s.p, s.u1, s.staked1, s.before)
	subscribe(s.p, s.u2, s.staked2, s.before)

	r2 := s.start.Add(types.ROUND * 2)
	amount, err := withdraw(s.p, s.u1, nil, r2)
	require.NoError(err)
	expectedU1Spent := s.staked1.MulRaw(8).QuoRaw(10)
	expectedU1TokenOut := s.totalOut.QuoRaw(3).MulRaw(2).QuoRaw(10)
	require.Equal(amount, expectedU1Spent, "we should withdraw 8/10 of our initial stake")
	checkUser(require, s.u1, zero, s.staked1, s.u1.OutPerShare, expectedU1TokenOut, "user1 should receive 1/3*2/10 of total purchase")
	triggerUserPurchase(s.p, s.u2)
	expectedU2TokenOut := s.totalOut.QuoRaw(3).MulRaw(4).QuoRaw(10)
	checkUser(require, s.u2, s.staked2, s.staked2, s.u2.OutPerShare, expectedU2TokenOut,
		"user2 purchase")

	pingSale(s.p, s.after)
	triggerUserPurchase(s.p, s.u1)
	triggerUserPurchase(s.p, s.u2)
	checkUser(require, s.u1, zero, s.staked1, s.u1.OutPerShare, expectedU1TokenOut, "no new purchase for u1")
	expectedU2TokenOut = expectedU2TokenOut.Add(s.totalOut.MulRaw(8).QuoRaw(10))
	checkUser(require, s.u2, zero, s.staked2, s.u2.OutPerShare, expectedU2TokenOut, "total purchase of u2")
}

// subscribe before the beginning and withdraw user 2 half stake after 2 rounds
func (s *TwoBuyersSuite) Test2Buyers_withdraw2() {
	require := s.Require()
	subscribe(s.p, s.u1, s.staked1, s.before)
	subscribe(s.p, s.u2, s.staked2, s.before)

	r2 := s.start.Add(types.ROUND * 2)
	u2RemainingHalfStake := s.staked2.MulRaw(8).QuoRaw(10).QuoRaw(2)

	amountOut, err := withdraw(s.p, s.u2, &u2RemainingHalfStake, r2)
	require.NoError(err)
	require.Equal(u2RemainingHalfStake.String(), amountOut.String())

	triggerUserPurchase(s.p, s.u1) // trigger for u2 is done in `withdraw` method
	expectedU1TokenOut := s.totalOut.QuoRaw(3).MulRaw(2).QuoRaw(10)
	expectedU2TokenOut := expectedU1TokenOut.MulRaw(2)
	checkUser(require, s.u1, s.staked1, s.staked1, s.u1.OutPerShare, expectedU1TokenOut, "user1 should receive 1/3*2/10 of total purchase")
	checkUser(require, s.u2, s.staked2.QuoRaw(2), s.staked2, s.u2.OutPerShare, expectedU2TokenOut,
		"user2 purchase")
	require.Equal(s.u1.Shares, s.u2.Shares, "after withdraw both users should have the same amount of shares")

	pingSale(s.p, s.after)
	triggerUserPurchase(s.p, s.u1)
	triggerUserPurchase(s.p, s.u2)
	remainingSale := s.totalOut.MulRaw(8).QuoRaw(10).QuoRaw(2)
	expectedU1TokenOut = expectedU1TokenOut.Add(remainingSale)
	checkUser(require, s.u1, zero, s.staked1, s.u1.OutPerShare, expectedU1TokenOut, "total purchase for u1")
	checkUser(require, s.u2, zero, s.staked2, s.u2.OutPerShare, expectedU2TokenOut.Add(remainingSale), "total purchase of u2")
}
