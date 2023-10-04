package poolmanager_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
)

// validates that the pool manager keeper can charge taker fees correctly.
// If the sender is whitelisted, then the taker fee is not charged.
// Otherwise, the taker fee is charged.
func (s *KeeperTestSuite) TestChargeTakerFee() {

	var (
		defaultTakerFee = osmomath.MustNewDecFromStr("0.01")
		defaultAmount   = sdk.NewInt(100)
	)

	tests := map[string]struct {
		shouldSetSenderWhitelist bool
		tokenIn                  sdk.Coin
		tokenOutDenom            string
		sender                   sdk.AccAddress
		exactIn                  bool
		takerFee                 osmomath.Dec

		expectedResult sdk.Coin
		expectError    error
	}{
		"fee charged on token in": {
			takerFee:      defaultTakerFee,
			tokenIn:       sdk.NewCoin(apptesting.ETH, defaultAmount),
			tokenOutDenom: apptesting.USDC,
			sender:        s.TestAccs[0],
			exactIn:       true,

			expectedResult: sdk.NewCoin(apptesting.ETH, defaultAmount.ToLegacyDec().Mul(osmomath.OneDec().Sub(defaultTakerFee)).TruncateInt()),
		},
		"fee bypassed due to sender being whitelisted": {
			takerFee:                 defaultTakerFee,
			tokenIn:                  sdk.NewCoin(apptesting.ETH, defaultAmount),
			tokenOutDenom:            apptesting.USDC,
			sender:                   s.TestAccs[0],
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
				poolManagerParams.TakerFeeParams.BypassWhitelist = []string{tc.sender.String()}
				poolManager.SetParams(s.Ctx, poolManagerParams)
			}

			// Create pool.
			s.PrepareConcentratedPool()

			// Set taker fee.
			poolManager.SetDenomPairTakerFee(s.Ctx, tc.tokenIn.Denom, tc.tokenOutDenom, tc.takerFee)

			// Pre-fund owner.
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(tc.tokenIn))

			// System under test.
			tokenInAfterTakerFee, err := poolManager.ChargeTakerFee(s.Ctx, tc.tokenIn, tc.tokenOutDenom, tc.sender, tc.exactIn)

			if tc.expectError != nil {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// Validate results.
			s.Require().Equal(tc.expectedResult.String(), tokenInAfterTakerFee.String())
		})
	}
}
