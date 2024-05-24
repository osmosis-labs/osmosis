package poolmanager_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v25/app/apptesting"
	txfeestypes "github.com/osmosis-labs/osmosis/v25/x/txfees/types"
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
		defaultAmount   = osmomath.NewInt(10000000)
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

			// System under test.
			tokenInAfterTakerFee, err := poolManager.ChargeTakerFee(s.Ctx, tc.tokenIn, tc.tokenOutDenom, s.TestAccs[tc.senderIndex], tc.exactIn)

			if tc.expectError != nil {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			var takerFeeTaken sdk.Coin
			if tc.exactIn {
				takerFeeTaken = tc.tokenIn.Sub(tokenInAfterTakerFee)
			} else {
				takerFeeTaken = tokenInAfterTakerFee.Sub(tc.tokenIn)
			}
			takerFeeModuleAccBal := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(txfeestypes.TakerFeeCollectorName))
			s.Require().True(sdk.NewCoins(takerFeeTaken).Equal(takerFeeModuleAccBal))
		})
	}
}
