package poolmanager_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v25/app/apptesting"
	"github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
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
			tokenInAfterTakerFee, _, err := poolManager.ChargeTakerFee(s.Ctx, tc.tokenIn, tc.tokenOutDenom, s.TestAccs[tc.senderIndex], tc.exactIn)

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

func (s *KeeperTestSuite) TestTakerFeeSkim() {

	tests := map[string]struct {
		alloyedPoolSetup                 func() []string
		expectedTakerFeeSkimAccumulators []types.TakerFeeSkimAccumulator
		expectedError                    error
	}{
		"one denomShareAgreement denom, one alloyedAssetShareAgreement denom, should be skimmed to denomShareAgreement": {
			alloyedPoolSetup: func() []string {
				return []string{"testA", s.setupAndRegisterAlloyedPool([]string{"testA", "testB"}, []uint16{1, 1})}
			},
			expectedTakerFeeSkimAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            "testA",
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin("testA", osmomath.NewInt(10000000)), sdk.NewCoin("testB", osmomath.NewInt(10000000)), sdk.NewCoin("testC", osmomath.NewInt(10000000))),
				},
			},
		},
		"two denomShareAgreement denoms, one alloyedAssetShareAgreement denom, should be skimmed to both denomShareAgreements": {
			alloyedPoolSetup: func() []string {
				return []string{"testA", "testB", s.setupAndRegisterAlloyedPool([]string{"testA", "testB"}, []uint16{1, 1})}
			},
			expectedTakerFeeSkimAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            "testA",
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin("testA", osmomath.NewInt(10000000)), sdk.NewCoin("testB", osmomath.NewInt(10000000)), sdk.NewCoin("testC", osmomath.NewInt(10000000))),
				},
				{
					Denom:            "testB",
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin("testA", osmomath.NewInt(20000000)), sdk.NewCoin("testB", osmomath.NewInt(20000000)), sdk.NewCoin("testC", osmomath.NewInt(20000000))),
				},
			},
		},
		"zero denomShareAgreement denoms, one alloyedAssetShareAgreement denom, should be skimmed to alloyedAssetShareAgreement": {
			alloyedPoolSetup: func() []string {
				return []string{s.setupAndRegisterAlloyedPool([]string{"testA", "testC"}, []uint16{1, 1})}
			},
			expectedTakerFeeSkimAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            "testA",
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin("testA", osmomath.NewInt(5000000)), sdk.NewCoin("testB", osmomath.NewInt(5000000)), sdk.NewCoin("testC", osmomath.NewInt(5000000))),
				},
			},
		},
		"zero denomShareAgreement denoms, zero alloyedAssetShareAgreement denoms, should not be skimmed": {
			alloyedPoolSetup: func() []string {
				return []string{"testC", "testD", "testE"}
			},
			expectedTakerFeeSkimAccumulators: []types.TakerFeeSkimAccumulator{},
		},
		"zero denomShareAgreement denoms, two alloyedAssetShareAgreement denoms, should be skimmed to both alloyedAssetShareAgreements": {
			alloyedPoolSetup: func() []string {
				alloyedDenom1 := s.setupAndRegisterAlloyedPool([]string{"testA", "testC"}, []uint16{1, 1})
				alloyedDenom2 := s.setupAndRegisterAlloyedPool([]string{"testB", "testC"}, []uint16{1, 1})
				return []string{alloyedDenom1, alloyedDenom2}
			},
			expectedTakerFeeSkimAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            "testA",
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin("testA", osmomath.NewInt(5000000)), sdk.NewCoin("testB", osmomath.NewInt(5000000)), sdk.NewCoin("testC", osmomath.NewInt(5000000))),
				},
				{
					Denom:            "testB",
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin("testA", osmomath.NewInt(10000000)), sdk.NewCoin("testB", osmomath.NewInt(10000000)), sdk.NewCoin("testC", osmomath.NewInt(10000000))),
				},
			},
		},
		"zero denomShareAgreement denoms, one alloyedAssetShareAgreement denom, but the alloyedAssetShareAgreement consists of no active denomShareAgreements, should not be skimmed": {
			alloyedPoolSetup: func() []string {
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3CustomRatio(s.TestAccs[0], []string{"testC", "testD", "testE"}, []uint16{1, 1, 1}, "osmosis", "x/cosmwasmpool/bytecode")
				alloyedDenom := fmt.Sprintf("factory/%s/alloyed/testdenom", cwPool.GetAddress().String())
				return []string{"testC", "testD", "testE", alloyedDenom}
			},
			expectedTakerFeeSkimAccumulators: []types.TakerFeeSkimAccumulator{},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			denomsInvolvedInSwap := tc.alloyedPoolSetup()

			s.setupTakerFeeShareAgreement("testA", "0.01", "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn")
			s.setupTakerFeeShareAgreement("testB", "0.02", "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2")

			takerFeeCoins := sdk.NewCoins(sdk.NewCoin("testA", osmomath.NewInt(1000000000)), sdk.NewCoin("testB", osmomath.NewInt(1000000000)), sdk.NewCoin("testC", osmomath.NewInt(1000000000)))

			err := s.App.PoolManagerKeeper.TakerFeeSkim(s.Ctx, denomsInvolvedInSwap, takerFeeCoins)
			if tc.expectedError != nil {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			takerFeeShareAccumulators := s.App.PoolManagerKeeper.GetAllTakerFeeShareAccumulators(s.Ctx)
			s.Require().Equal(tc.expectedTakerFeeSkimAccumulators, takerFeeShareAccumulators)
		})
	}
}

func (s *KeeperTestSuite) setupAndRegisterAlloyedPool(denoms []string, ratios []uint16) string {
	cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3CustomRatio(s.TestAccs[0], denoms, ratios, "osmosis", "x/cosmwasmpool/bytecode")
	err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
	s.Require().NoError(err)
	return fmt.Sprintf("factory/%s/alloyed/testdenom", cwPool.GetAddress().String())
}

func (s *KeeperTestSuite) setupTakerFeeShareAgreement(denom string, skimPercent string, skimAddress string) {
	s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
		Denom:       denom,
		SkimPercent: osmomath.MustNewDecFromStr(skimPercent),
		SkimAddress: skimAddress,
	})
}
