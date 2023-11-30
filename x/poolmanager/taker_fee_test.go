package poolmanager_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/app/apptesting"
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
		sendCoins      bool
	}{
		"fee charged on token in": {
			takerFee:       defaultTakerFee,
			tokenIn:        sdk.NewCoin(apptesting.ETH, defaultAmount),
			tokenOutDenom:  apptesting.USDC,
			senderIndex:    whitelistedSenderIndex,
			exactIn:        true,
			expectedResult: sdk.NewCoin(apptesting.ETH, defaultAmount.ToLegacyDec().Mul(osmomath.OneDec().Sub(defaultTakerFee)).TruncateInt()),
		},
		"fee charged on token in due to different address being whitelisted": {
			takerFee:                 defaultTakerFee,
			tokenIn:                  sdk.NewCoin(apptesting.ETH, defaultAmount),
			tokenOutDenom:            apptesting.USDC,
			senderIndex:              nonWhitelistedSenderIndex,
			exactIn:                  true,
			shouldSetSenderWhitelist: true,
			expectedResult:           sdk.NewCoin(apptesting.ETH, defaultAmount.ToLegacyDec().Mul(osmomath.OneDec().Sub(defaultTakerFee)).TruncateInt()),
		},
		"fee bypassed due to sender being whitelisted": {
			takerFee:                 defaultTakerFee,
			tokenIn:                  sdk.NewCoin(apptesting.ETH, defaultAmount),
			tokenOutDenom:            apptesting.USDC,
			senderIndex:              whitelistedSenderIndex,
			exactIn:                  true,
			shouldSetSenderWhitelist: true,
			expectedResult:           sdk.NewCoin(apptesting.ETH, defaultAmount),
		},
		"fee charged on token out": {
			takerFee:      defaultTakerFee,
			tokenIn:       sdk.NewCoin(apptesting.ETH, defaultAmount),
			tokenOutDenom: apptesting.USDC,
			senderIndex:   whitelistedSenderIndex,
			exactIn:       false,

			expectedResult: sdk.NewCoin(apptesting.ETH, defaultAmount.ToLegacyDec().Quo(osmomath.OneDec().Sub(defaultTakerFee)).Ceil().TruncateInt()),
		},
		"fee charged on token out due to different address being whitelisted": {
			takerFee:                 defaultTakerFee,
			tokenIn:                  sdk.NewCoin(apptesting.ETH, defaultAmount),
			tokenOutDenom:            apptesting.USDC,
			senderIndex:              nonWhitelistedSenderIndex,
			exactIn:                  false,
			shouldSetSenderWhitelist: true,

			expectedResult: sdk.NewCoin(apptesting.ETH, defaultAmount.ToLegacyDec().Quo(osmomath.OneDec().Sub(defaultTakerFee)).Ceil().TruncateInt()),
		},
		"sender does not have enough coins in": {
			takerFee:                 defaultTakerFee,
			tokenIn:                  sdk.NewCoin(apptesting.ETH, defaultAmount),
			tokenOutDenom:            apptesting.USDC,
			senderIndex:              nonWhitelistedSenderIndex,
			exactIn:                  true,
			shouldSetSenderWhitelist: true,

			sendCoins:   true,
			expectError: fmt.Errorf("insufficient funds"),
		},
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

			// Send coins.
			if tc.sendCoins {
				s.App.BankKeeper.SendCoins(s.Ctx, s.TestAccs[nonWhitelistedSenderIndex], s.TestAccs[whitelistedSenderIndex], sdk.NewCoins(tc.tokenIn))
			}

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

			var expectedTotalTakerFee osmomath.Int
			if tc.exactIn {
				expectedTotalTakerFee = defaultAmount.Sub(tc.expectedResult.Amount)
			} else {
				expectedTotalTakerFee = tc.expectedResult.Amount.Sub(defaultAmount)
			}
			expectedTakerFeeToStakersAmount := expectedTotalTakerFee.ToLegacyDec().Mul(params.TakerFeeParams.NonOsmoTakerFeeDistribution.StakingRewards)
			expectedTakerFeeToCommunityPoolAmount := expectedTotalTakerFee.ToLegacyDec().Mul(params.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool)

			roundup := func(d sdkmath.LegacyDec) sdkmath.Int {
				if d.Sub(sdkmath.LegacyNewDecFromInt(d.TruncateInt())).GT(sdkmath.LegacyZeroDec()) {
					return d.TruncateInt().Add(sdkmath.NewInt(1))
				}
				return d.TruncateInt()
			}
			expectedTakerFeeToStakers := sdk.NewCoin(tc.expectedResult.Denom, roundup(expectedTakerFeeToStakersAmount))
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
