package poolmanager_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/app/apptesting"
)

// validates that the pool manager keeper can charge taker fees correctly.
// If the sender is whitelisted, then the taker fee is not charged.
// Otherwise, the taker fee is charged.
func (s *KeeperTestSuite) TestChargeTakerFee() {

	const (
		whitelistedSenderIndex = iota
		nonWhitelistedSenderIndex
	)

	var (
		defaultTakerFee = osmomath.MustNewDecFromStr("0.01")
		defaultAmount   = sdk.NewInt(10000000)
	)

	tests := map[string]struct {
		shouldSetSenderWhitelist bool
		tokenIn                  sdk.Coin
		tokenOutDenom            string
		senderIndex              int
		exactIn                  bool
		takerFee                 osmomath.Dec

		expectedResult sdk.Coin
		expectError    error
	}{
		"fee charged on token in": {
			takerFee:      defaultTakerFee,
			tokenIn:       sdk.NewCoin(apptesting.ETH, defaultAmount),
			tokenOutDenom: apptesting.USDC,
			senderIndex:   whitelistedSenderIndex,
			exactIn:       true,

			expectedResult: sdk.NewCoin(apptesting.ETH, defaultAmount.ToLegacyDec().Mul(osmomath.OneDec().Sub(defaultTakerFee)).TruncateInt()),
		},
		"fee charged on token in due to different address being whitelisted": {
			takerFee:                 defaultTakerFee,
			tokenIn:                  sdk.NewCoin(apptesting.ETH, defaultAmount),
			tokenOutDenom:            apptesting.USDC,
			senderIndex:              nonWhitelistedSenderIndex,
			exactIn:                  true,
			shouldSetSenderWhitelist: true,

			expectedResult: sdk.NewCoin(apptesting.ETH, defaultAmount.ToLegacyDec().Mul(osmomath.OneDec().Sub(defaultTakerFee)).TruncateInt()),
		},
		"fee bypassed due to sender being whitelisted": {
			takerFee:                 defaultTakerFee,
			tokenIn:                  sdk.NewCoin(apptesting.ETH, defaultAmount),
			tokenOutDenom:            apptesting.USDC,
			senderIndex:              whitelistedSenderIndex,
			exactIn:                  true,
			shouldSetSenderWhitelist: true,

			expectedResult: sdk.NewCoin(apptesting.ETH, defaultAmount),
		},
		// TODO: under more test cases
		// https://github.com/osmosis-labs/osmosis/issues/6633
		// - exactOut: false
		// - sender does not have enough coins
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			poolManager := s.App.PoolManagerKeeper

			// Set whitelist.
			if tc.shouldSetSenderWhitelist {
				poolManagerParams := poolManager.GetParams(s.Ctx)
				poolManagerParams.TakerFeeParams.ReducedFeeWhitelist = []string{s.TestAccs[whitelistedSenderIndex].String()}
				poolManager.SetParams(s.Ctx, poolManagerParams)
			}

			// Create pool.
			s.PrepareConcentratedPool()

			// Set taker fee.
			poolManager.SetDenomPairTakerFee(s.Ctx, tc.tokenIn.Denom, tc.tokenOutDenom, tc.takerFee)

			// Pre-fund owner.
			s.FundAcc(s.TestAccs[tc.senderIndex], sdk.NewCoins(tc.tokenIn))

			// Check the taker fee tracker before the taker fee is charged.
			takerFeeTrackerForStakersBefore := poolManager.GetTakerFeeTrackerForStakers(s.Ctx)
			takerFeeTrackerForCommunityPoolBefore := poolManager.GetTakerFeeTrackerForCommunityPool(s.Ctx)

			// System under test.
			tokenInAfterTakerFee, err := poolManager.ChargeTakerFee(s.Ctx, tc.tokenIn, tc.tokenOutDenom, s.TestAccs[tc.senderIndex], tc.exactIn)

			// Check the taker fee tracker after the taker fee is charged.
			takerFeeTrackerForStakersAfter := poolManager.GetTakerFeeTrackerForStakers(s.Ctx)
			takerFeeTrackerForCommunityPoolAfter := poolManager.GetTakerFeeTrackerForCommunityPool(s.Ctx)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().Equal(takerFeeTrackerForStakersBefore, takerFeeTrackerForStakersAfter)
				s.Require().Equal(takerFeeTrackerForCommunityPoolBefore, takerFeeTrackerForCommunityPoolAfter)
				return
			}
			s.Require().NoError(err)

			params := s.App.PoolManagerKeeper.GetParams(s.Ctx)
			expectedTotalTakerFee := defaultAmount.Sub(tc.expectedResult.Amount)
			expectedTakerFeeToStakersAmount := expectedTotalTakerFee.ToLegacyDec().Mul(params.TakerFeeParams.NonOsmoTakerFeeDistribution.StakingRewards)
			expectedTakerFeeToCommunityPoolAmount := expectedTotalTakerFee.ToLegacyDec().Mul(params.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool)
			expectedTakerFeeToStakers := sdk.NewCoin(tc.expectedResult.Denom, expectedTakerFeeToStakersAmount.TruncateInt())
			expectedTakerFeeToCommunityPool := sdk.NewCoin(tc.expectedResult.Denom, expectedTakerFeeToCommunityPoolAmount.TruncateInt())

			// Validate results.
			s.Require().Equal(tc.expectedResult.String(), tokenInAfterTakerFee.String())
			expectedTakerFeeTrackerForStakersAfter := takerFeeTrackerForStakersBefore.Add(expectedTakerFeeToStakers)
			if expectedTakerFeeTrackerForStakersAfter.Empty() {
				expectedTakerFeeTrackerForStakersAfter = sdk.Coins(nil)
			}
			s.Require().Equal(expectedTakerFeeTrackerForStakersAfter, takerFeeTrackerForStakersAfter)
			expectedTakerFeeTrackerForCommunityPoolAfter := takerFeeTrackerForCommunityPoolBefore.Add(expectedTakerFeeToCommunityPool)
			if expectedTakerFeeTrackerForCommunityPoolAfter.Empty() {
				expectedTakerFeeTrackerForCommunityPoolAfter = sdk.Coins(nil)
			}
			s.Require().Equal(expectedTakerFeeTrackerForCommunityPoolAfter, takerFeeTrackerForCommunityPoolAfter)
		})
	}
}
