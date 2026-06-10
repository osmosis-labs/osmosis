package keeper_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v31/x/mint/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v31/x/txfees/types"
)

// mintDenom is the bond/mint denom used by the app test harness.
func (s *KeeperTestSuite) mintDenom() string {
	return s.App.MintKeeper.GetParams(s.Ctx).MintDenom
}

// rawSupply returns the bank raw supply (GetSupply) of the mint denom.
func (s *KeeperTestSuite) rawSupply() osmomath.Int {
	return s.App.BankKeeper.GetSupply(s.Ctx, s.mintDenom()).Amount
}

// TestGetBurnedSupply asserts burned supply equals the balance of the null
// (burn) address, and tracks coins sent there.
func (s *KeeperTestSuite) TestGetBurnedSupply() {
	s.SetupTest()
	denom := s.mintDenom()

	// Initially the burn address holds nothing of the mint denom.
	s.Require().True(s.App.MintKeeper.GetBurnedSupply(s.Ctx).IsZero())

	// Mint to the module, then send a known amount to the burn address.
	burnAmt := osmomath.NewInt(5_550_000)
	s.MintCoins(sdk.NewCoins(sdk.NewCoin(denom, burnAmt)))
	err := s.App.BankKeeper.SendCoinsFromModuleToAccount(
		s.Ctx, types.ModuleName, txfeestypes.DefaultNullAddress, sdk.NewCoins(sdk.NewCoin(denom, burnAmt)))
	s.Require().NoError(err)

	s.Require().Equal(burnAmt, s.App.MintKeeper.GetBurnedSupply(s.Ctx))
}

// TestGetTotalSupply asserts total supply == raw GetSupply - burned.
func (s *KeeperTestSuite) TestGetTotalSupply() {
	s.SetupTest()
	denom := s.mintDenom()

	mintAmt := osmomath.NewInt(100_000_000)
	burnAmt := osmomath.NewInt(4_000_000)
	s.MintCoins(sdk.NewCoins(sdk.NewCoin(denom, mintAmt)))
	rawBefore := s.rawSupply()

	err := s.App.BankKeeper.SendCoinsFromModuleToAccount(
		s.Ctx, types.ModuleName, txfeestypes.DefaultNullAddress, sdk.NewCoins(sdk.NewCoin(denom, burnAmt)))
	s.Require().NoError(err)

	// Sending to the burn address does not change raw supply; total = raw - burned.
	expectedTotal := rawBefore.Sub(burnAmt)
	s.Require().Equal(expectedTotal, s.App.MintKeeper.GetTotalSupply(s.Ctx))
}

// TestCirculatingSupplyIdentity asserts the core accounting identity holds for
// every state we construct: circulating == total - restricted == minted -
// burned - restricted.
func (s *KeeperTestSuite) TestCirculatingSupplyIdentity() {
	s.SetupTest()
	denom := s.mintDenom()

	s.MintCoins(sdk.NewCoins(sdk.NewCoin(denom, osmomath.NewInt(50_000_000))))
	err := s.App.BankKeeper.SendCoinsFromModuleToAccount(
		s.Ctx, types.ModuleName, txfeestypes.DefaultNullAddress, sdk.NewCoins(sdk.NewCoin(denom, osmomath.NewInt(1_000_000))))
	s.Require().NoError(err)

	total := s.App.MintKeeper.GetTotalSupply(s.Ctx)
	restricted := s.App.MintKeeper.GetRestrictedSupply(s.Ctx)
	circulating := s.App.MintKeeper.GetCirculatingSupply(s.Ctx)

	// Identity 1: circulating == total - restricted.
	s.Require().Equal(total.Sub(restricted), circulating)

	// Identity 2: circulating == minted - burned - restricted.
	minted := s.rawSupply()
	burned := s.App.MintKeeper.GetBurnedSupply(s.Ctx)
	s.Require().Equal(minted.Sub(burned).Sub(restricted), circulating)
}

// TestRestrictedSupplyIncludesDevVesting asserts the developer-vesting module
// account balance is part of restricted supply.
func (s *KeeperTestSuite) TestRestrictedSupplyIncludesDevVesting() {
	s.SetupTest()
	denom := s.mintDenom()

	devVestingAddr := s.App.AccountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	s.Require().NotNil(devVestingAddr)
	devVestingBalance := s.App.BankKeeper.GetBalance(s.Ctx, devVestingAddr, denom).Amount

	restricted := s.App.MintKeeper.GetRestrictedSupply(s.Ctx)
	// Restricted supply must be at least the dev-vesting balance (other
	// components are >= 0).
	s.Require().True(restricted.GTE(devVestingBalance),
		"restricted (%s) should include dev-vesting balance (%s)", restricted, devVestingBalance)
}

// TestDevVestingCountedExactlyOnce is the regression guard for the double-count
// bug. It proves the dev-vesting balance contributes net-zero to circulating
// supply: it is present once in raw GetSupply (the base for total) and removed
// once via restricted supply. Equivalently, circulating computed on the raw
// base must equal (GetSupplyWithOffset - burned - non-devVesting-restricted),
// NOT (GetSupplyWithOffset - burned - restricted) which would subtract
// dev-vesting twice.
func (s *KeeperTestSuite) TestDevVestingCountedExactlyOnce() {
	s.SetupTest()
	denom := s.mintDenom()

	devVestingAddr := s.App.AccountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	devVestingBalance := s.App.BankKeeper.GetBalance(s.Ctx, devVestingAddr, denom).Amount

	raw := s.rawSupply()
	withOffset := s.App.BankKeeper.GetSupplyWithOffset(s.Ctx, denom).Amount

	// Sanity: the offset nets out exactly the current dev-vesting balance once.
	s.Require().Equal(raw.Sub(devVestingBalance), withOffset,
		"GetSupplyWithOffset should equal raw supply minus dev-vesting balance")

	circulating := s.App.MintKeeper.GetCirculatingSupply(s.Ctx)
	restricted := s.App.MintKeeper.GetRestrictedSupply(s.Ctx)
	burned := s.App.MintKeeper.GetBurnedSupply(s.Ctx)

	// The correct circulating equals raw - burned - restricted.
	s.Require().Equal(raw.Sub(burned).Sub(restricted), circulating)

	// The buggy circulating (double-count) would be withOffset - burned -
	// restricted, which is exactly devVestingBalance lower. Assert we are NOT
	// that value (unless dev-vesting is zero, in which case they coincide).
	buggy := withOffset.Sub(burned).Sub(restricted)
	if devVestingBalance.IsPositive() {
		s.Require().NotEqual(buggy, circulating,
			"circulating must not double-subtract dev-vesting")
		s.Require().Equal(circulating.Sub(buggy), devVestingBalance,
			"the difference between correct and buggy circulating must equal exactly the dev-vesting balance")
	}
}

// TestRestrictedSupplyIncludesStakedAmount asserts that OSMO delegated by a
// restricted entity (here, a dev-reward receiver address) is counted in
// restricted supply via the share->token conversion, not just its liquid
// balance.
func (s *KeeperTestSuite) TestRestrictedSupplyIncludesStakedAmount() {
	s.SetupTest()
	denom := s.mintDenom()

	// Stand up a validator to delegate to.
	valAddr := s.SetupValidator(stakingtypes.Bonded)
	validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().NoError(err)

	// Use a dev-reward receiver address as the restricted holder (already part
	// of restricted accounting), fund it, and delegate.
	holder := testAddressOne
	params := s.App.MintKeeper.GetParams(s.Ctx)
	params.WeightedDeveloperRewardsReceivers = []types.WeightedAddress{
		{Address: holder.String(), Weight: osmomath.OneDec()},
	}
	s.App.MintKeeper.SetParams(s.Ctx, params)

	delegateAmt := osmomath.NewInt(7_000_000)
	s.FundAcc(holder, sdk.NewCoins(sdk.NewCoin(denom, delegateAmt)))

	restrictedBefore := s.App.MintKeeper.GetRestrictedSupply(s.Ctx)

	_, err = s.App.StakingKeeper.Delegate(s.Ctx, holder, delegateAmt, stakingtypes.Unbonded, validator, true)
	s.Require().NoError(err)

	restrictedAfter := s.App.MintKeeper.GetRestrictedSupply(s.Ctx)

	// Delegating moves the holder's liquid balance into staked, so the aggregate
	// (balance + staked) restricted supply must be unchanged within rounding. A
	// weak lower bound (>= before - 1) would also pass if the staked term were
	// dropped entirely, so assert near-equality in BOTH directions: restricted
	// must not have fallen by anywhere near delegateAmt (which is what would
	// happen if staked weren't counted), and must not have risen either.
	one := osmomath.OneInt()
	s.Require().True(restrictedAfter.LTE(restrictedBefore.Add(one)),
		"restricted should not increase on delegate (before=%s after=%s)", restrictedBefore, restrictedAfter)
	s.Require().True(restrictedAfter.GTE(restrictedBefore.Sub(one)),
		"restricted must be flat within rounding; staked amount is counted (before=%s after=%s)", restrictedBefore, restrictedAfter)

	// Belt and suspenders: prove the holder's staked amount is non-zero and is
	// reflected, i.e. removing the holder's stake would drop restricted by ~the
	// delegated amount. (If staked weren't counted, restrictedAfter would equal
	// restrictedBefore - delegateAmt, far outside the +/-1 band asserted above.)
	dropIfStakeIgnored := restrictedBefore.Sub(delegateAmt)
	s.Require().True(restrictedAfter.GT(dropIfStakeIgnored.Add(one)),
		"restricted must include the staked term, not just liquid balance")
}

// TestSupplyQueryHandlers exercises the four gRPC handlers end to end through
// the query client.
func (s *KeeperTestSuite) TestSupplyQueryHandlers() {
	s.SetupTest()

	burnedRes, err := s.queryClient.BurnedSupply(context.Background(), &types.QueryBurnedRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.App.MintKeeper.GetBurnedSupply(s.Ctx), burnedRes.Burned)

	totalRes, err := s.queryClient.TotalSupply(context.Background(), &types.QueryTotalSupplyRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.App.MintKeeper.GetTotalSupply(s.Ctx), totalRes.TotalSupply)

	restrictedRes, err := s.queryClient.RestrictedSupply(context.Background(), &types.QueryRestrictedSupplyRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.App.MintKeeper.GetRestrictedSupply(s.Ctx), restrictedRes.RestrictedSupply)

	circulatingRes, err := s.queryClient.CirculatingSupply(context.Background(), &types.QueryCirculatingSupplyRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.App.MintKeeper.GetCirculatingSupply(s.Ctx), circulatingRes.CirculatingSupply)

	// Handler identity check.
	s.Require().Equal(
		totalRes.TotalSupply.Sub(restrictedRes.RestrictedSupply),
		circulatingRes.CirculatingSupply,
	)
}

// TestInflationUsesCirculatingSupply asserts GetInflation now divides by
// circulating supply (the denominator change), by checking the computed rate
// matches the hand-derived value against circulating, not offset-total.
func (s *KeeperTestSuite) TestInflationUsesCirculatingSupply() {
	s.SetupTest()

	circulating := s.App.MintKeeper.GetCirculatingSupply(s.Ctx)
	s.Require().True(circulating.IsPositive())

	minter := s.App.MintKeeper.GetMinter(s.Ctx)
	params := s.App.MintKeeper.GetParams(s.Ctx)
	oneMinusCommunityPool := osmomath.OneDec().Sub(params.DistributionProportions.CommunityPool)
	expected := minter.EpochProvisions.
		Mul(oneMinusCommunityPool).
		Mul(osmomath.NewDec(365)).
		Quo(circulating.ToLegacyDec())

	inflation, err := s.App.MintKeeper.GetInflation(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(expected, inflation)
}
