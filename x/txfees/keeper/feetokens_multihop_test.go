package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v31/x/txfees/types"
)

// setupTwoHopFooViaBar creates balancer pools for FOO/BAR and BAR/OSMO and registers them in
// protorev so that find2HopRoute can discover the path. Pool reserves are chosen so that
// 1 FOO is worth 0.5 BAR and 1 BAR is worth 2 OSMO, giving a compounded spot price of
// 1 FOO = 1 OSMO. Adds BAR to FeeSwapIntermediaryDenomList.
//
// Returns (fooBarPoolId, barOsmoPoolId).
func (s *KeeperTestSuite) setupTwoHopFooViaBar(baseDenom string) (uint64, uint64) {
	// FOO/BAR pool: 200 FOO, 100 BAR. SpotPrice(quote=BAR, base=FOO) = 100/200 = 0.5.
	fooBarPoolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin("foo", 200),
		sdk.NewInt64Coin("bar", 100),
	)
	// BAR/OSMO pool: 100 BAR, 200 OSMO. SpotPrice(quote=OSMO, base=BAR) = 200/100 = 2.
	barOsmoPoolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin("bar", 100),
		sdk.NewInt64Coin(baseDenom, 200),
	)

	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, "foo", "bar", fooBarPoolId)
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, "bar", baseDenom, barOsmoPoolId)

	params := s.App.TxFeesKeeper.GetParams(s.Ctx)
	params.FeeSwapIntermediaryDenomList = []string{"bar"}
	s.App.TxFeesKeeper.SetParams(s.Ctx, params)

	return fooBarPoolId, barOsmoPoolId
}

// TestValidateFeeToken_MultiHopFallback verifies that a fee token without a direct OSMO pool
// is accepted when a 2-hop route can be discovered via protorev, and rejected when neither a
// direct pool nor a 2-hop route exists.
func (s *KeeperTestSuite) TestValidateFeeToken_MultiHopFallback() {
	baseDenom := sdk.DefaultBondDenom

	s.Run("accepts foo when only a 2-hop route via bar exists", func() {
		s.SetupTest(false)
		fooBarPoolId, _ := s.setupTwoHopFooViaBar(baseDenom)

		// PoolID = fooBarPoolId. This pool does NOT contain baseDenom, so the direct-route
		// validation will fail and the validator must fall back to 2-hop discovery.
		feeToken := types.FeeToken{Denom: "foo", PoolID: fooBarPoolId}

		err := s.App.TxFeesKeeper.ValidateFeeToken(s.Ctx, feeToken)
		s.Require().NoError(err)
	})

	s.Run("rejects foo when no direct pool and no whitelisted intermediary route", func() {
		s.SetupTest(false)
		// Create the FOO/BAR pool but do NOT whitelist BAR as an intermediary, and do NOT
		// register the BAR/OSMO leg in protorev. Direct validation fails; fallback also fails.
		fooBarPoolId := s.PrepareBalancerPoolWithCoins(
			sdk.NewInt64Coin("foo", 200),
			sdk.NewInt64Coin("bar", 100),
		)

		feeToken := types.FeeToken{Denom: "foo", PoolID: fooBarPoolId}

		err := s.App.TxFeesKeeper.ValidateFeeToken(s.Ctx, feeToken)
		s.Require().Error(err)
	})

	s.Run("rejects basedenom regardless of route availability", func() {
		s.SetupTest(false)
		s.setupTwoHopFooViaBar(baseDenom)
		feeToken := types.FeeToken{Denom: baseDenom, PoolID: 1}

		err := s.App.TxFeesKeeper.ValidateFeeToken(s.Ctx, feeToken)
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "cannot add basedenom")
	})

	s.Run("accepts foo via direct route even when 2-hop fallback would fail", func() {
		s.SetupTest(false)
		// Direct foo/OSMO pool. No FeeSwapIntermediaryDenomList configured and no
		// protorev registrations, so the 2-hop fallback would always fail. The
		// direct-route path should still succeed without ever consulting the fallback.
		directPoolId := s.PrepareBalancerPoolWithCoins(
			sdk.NewInt64Coin("foo", 100),
			sdk.NewInt64Coin(baseDenom, 100),
		)

		feeToken := types.FeeToken{Denom: "foo", PoolID: directPoolId}

		err := s.App.TxFeesKeeper.ValidateFeeToken(s.Ctx, feeToken)
		s.Require().NoError(err)
	})
}

// TestCalcFeeSpotPrice_MultiHopFallback verifies that CalcFeeSpotPrice falls back to the
// compounded 2-hop spot price when the registered direct pool does not contain baseDenom.
// Pool setup: 1 FOO = 0.5 BAR, 1 BAR = 2 OSMO => 1 FOO = 1 OSMO.
func (s *KeeperTestSuite) TestCalcFeeSpotPrice_MultiHopFallback() {
	s.SetupTest(false)
	baseDenom := sdk.DefaultBondDenom
	fooBarPoolId, _ := s.setupTwoHopFooViaBar(baseDenom)

	// Register foo with the FOO/BAR pool. Direct validation will fail (pool has no OSMO),
	// but the 2-hop fallback should succeed.
	err := s.App.TxFeesKeeper.SetFeeTokens(s.Ctx, []types.FeeToken{
		{Denom: "foo", PoolID: fooBarPoolId},
	})
	s.Require().NoError(err)

	spotPrice, err := s.App.TxFeesKeeper.CalcFeeSpotPrice(s.Ctx, "foo")
	s.Require().NoError(err)

	// Expect 1 FOO = 1 OSMO. Compounded: 0.5 (BAR/FOO) * 2 (OSMO/BAR) = 1 (OSMO/FOO).
	s.Require().True(spotPrice.Equal(osmomath.NewBigDec(1)),
		"expected 1 OSMO per FOO, got %s", spotPrice.String())
}

// TestConvertToBaseToken_MultiHopFallback verifies the end-to-end fee conversion path for a
// fee token that only has a 2-hop route to OSMO.
func (s *KeeperTestSuite) TestConvertToBaseToken_MultiHopFallback() {
	s.SetupTest(false)
	baseDenom := sdk.DefaultBondDenom
	fooBarPoolId, _ := s.setupTwoHopFooViaBar(baseDenom)

	err := s.App.TxFeesKeeper.SetFeeTokens(s.Ctx, []types.FeeToken{
		{Denom: "foo", PoolID: fooBarPoolId},
	})
	s.Require().NoError(err)

	// 100 FOO * 1 (OSMO/FOO) = 100 OSMO.
	converted, err := s.App.TxFeesKeeper.ConvertToBaseToken(s.Ctx, sdk.NewInt64Coin("foo", 100))
	s.Require().NoError(err)
	s.Require().Equal(baseDenom, converted.Denom)
	s.Require().Equal(int64(100), converted.Amount.Int64())
}

// TestCalcFeeSpotPrice_DirectRoutePreferred verifies that when a token has a direct OSMO pool
// registered, the direct route is used even if a 2-hop route would also be discoverable.
// This guards the existing direct-route behaviour from regressing under the new fallback.
func (s *KeeperTestSuite) TestCalcFeeSpotPrice_DirectRoutePreferred() {
	s.SetupTest(false)
	baseDenom := sdk.DefaultBondDenom

	// Direct foo/OSMO pool with 1:1 reserves.
	directPoolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin("foo", 100),
		sdk.NewInt64Coin(baseDenom, 100),
	)

	// Also wire up a 2-hop fallback that would give a different price if used.
	// FOO/BAR 1:1, BAR/OSMO 1:10. If the 2-hop path were taken, spot price would be 10x.
	fooBarPoolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin("foo", 100),
		sdk.NewInt64Coin("bar", 100),
	)
	barOsmoPoolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin("bar", 100),
		sdk.NewInt64Coin(baseDenom, 1000),
	)
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, "foo", "bar", fooBarPoolId)
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, "bar", baseDenom, barOsmoPoolId)
	params := s.App.TxFeesKeeper.GetParams(s.Ctx)
	params.FeeSwapIntermediaryDenomList = []string{"bar"}
	s.App.TxFeesKeeper.SetParams(s.Ctx, params)

	err := s.App.TxFeesKeeper.SetFeeTokens(s.Ctx, []types.FeeToken{
		{Denom: "foo", PoolID: directPoolId},
	})
	s.Require().NoError(err)

	spotPrice, err := s.App.TxFeesKeeper.CalcFeeSpotPrice(s.Ctx, "foo")
	s.Require().NoError(err)
	s.Require().True(spotPrice.Equal(osmomath.NewBigDec(1)),
		"expected direct route to yield 1, got %s (would be 10 via fallback)", spotPrice.String())
}
