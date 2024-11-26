package poolmanager_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v27/x/txfees/types"
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
				return []string{denomA, s.setupAndRegisterAlloyedPool([]string{denomA, denomB}, []uint16{1, 1})}
			},
			expectedTakerFeeSkimAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            denomA,
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(10000000)), sdk.NewCoin(denomB, osmomath.NewInt(10000000)), sdk.NewCoin(denomC, osmomath.NewInt(10000000))),
				},
			},
		},
		"two denomShareAgreement denoms, one alloyedAssetShareAgreement denom, should be skimmed to both denomShareAgreements": {
			alloyedPoolSetup: func() []string {
				return []string{denomA, denomB, s.setupAndRegisterAlloyedPool([]string{denomA, denomB}, []uint16{1, 1})}
			},
			expectedTakerFeeSkimAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            denomA,
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(10000000)), sdk.NewCoin(denomB, osmomath.NewInt(10000000)), sdk.NewCoin(denomC, osmomath.NewInt(10000000))),
				},
				{
					Denom:            denomB,
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(20000000)), sdk.NewCoin(denomB, osmomath.NewInt(20000000)), sdk.NewCoin(denomC, osmomath.NewInt(20000000))),
				},
			},
		},
		"zero denomShareAgreement denoms, one alloyedAssetShareAgreement denom, should be skimmed to alloyedAssetShareAgreement": {
			alloyedPoolSetup: func() []string {
				return []string{s.setupAndRegisterAlloyedPool([]string{denomA, denomC}, []uint16{1, 1})}
			},
			expectedTakerFeeSkimAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            denomA,
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(5000000)), sdk.NewCoin(denomB, osmomath.NewInt(5000000)), sdk.NewCoin(denomC, osmomath.NewInt(5000000))),
				},
			},
		},
		"zero denomShareAgreement denoms, zero alloyedAssetShareAgreement denoms, should not be skimmed": {
			alloyedPoolSetup: func() []string {
				return []string{denomC, OSMO, ATOM}
			},
			expectedTakerFeeSkimAccumulators: []types.TakerFeeSkimAccumulator{},
		},
		"zero denomShareAgreement denoms, two alloyedAssetShareAgreement denoms, should be skimmed to both alloyedAssetShareAgreements": {
			alloyedPoolSetup: func() []string {
				alloyedDenom1 := s.setupAndRegisterAlloyedPool([]string{denomA, denomC}, []uint16{1, 1})
				alloyedDenom2 := s.setupAndRegisterAlloyedPool([]string{denomB, denomC}, []uint16{1, 1})
				return []string{alloyedDenom1, alloyedDenom2}
			},
			expectedTakerFeeSkimAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            denomA,
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(5000000)), sdk.NewCoin(denomB, osmomath.NewInt(5000000)), sdk.NewCoin(denomC, osmomath.NewInt(5000000))),
				},
				{
					Denom:            denomB,
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(10000000)), sdk.NewCoin(denomB, osmomath.NewInt(10000000)), sdk.NewCoin(denomC, osmomath.NewInt(10000000))),
				},
			},
		},
		"zero denomShareAgreement denoms, one alloyedAssetShareAgreement denom, but the alloyedAssetShareAgreement consists of no active denomShareAgreements, should not be skimmed": {
			alloyedPoolSetup: func() []string {
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{denomC, OSMO, ATOM}, nil)
				alloyedDenom := createAlloyedDenom(cwPool.GetAddress().String())
				return []string{denomC, OSMO, ATOM, alloyedDenom}
			},
			expectedTakerFeeSkimAccumulators: []types.TakerFeeSkimAccumulator{},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			denomsInvolvedInSwap := tc.alloyedPoolSetup()

			s.setupTakerFeeShareAgreement(denomA, "0.01", "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn")
			s.setupTakerFeeShareAgreement(denomB, "0.02", "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2")

			takerFeeCoins := sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(1000000000)), sdk.NewCoin(denomB, osmomath.NewInt(1000000000)), sdk.NewCoin(denomC, osmomath.NewInt(1000000000)))

			err := s.App.PoolManagerKeeper.TakerFeeSkim(s.Ctx, denomsInvolvedInSwap, takerFeeCoins)
			if tc.expectedError != nil {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			takerFeeShareAccumulators, err := s.App.PoolManagerKeeper.GetAllTakerFeeShareAccumulators(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedTakerFeeSkimAccumulators, takerFeeShareAccumulators)
		})
	}
}

func (s *KeeperTestSuite) TestGetTakerFeeShareAgreements() {
	tests := map[string]struct {
		setupFunc             func() []string
		expectedDenomShares   []types.TakerFeeShareAgreement
		expectedAlloyedShares []types.TakerFeeShareAgreement
	}{
		"one denomShareAgreement denom": {
			setupFunc: func() []string {
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements)
				return []string{denomA}
			},
			expectedDenomShares:   []types.TakerFeeShareAgreement{defaultTakerFeeShareAgreements[0]},
			expectedAlloyedShares: []types.TakerFeeShareAgreement{},
		},
		"one alloyedAssetShareAgreement denom": {
			setupFunc: func() []string {
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
				alloyedDenom := s.setupAndRegisterAlloyedPool([]string{denomA, denomB}, []uint16{1, 1})
				return []string{alloyedDenom}
			},
			expectedDenomShares:   []types.TakerFeeShareAgreement{},
			expectedAlloyedShares: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{osmomath.MustNewDecFromStr("0.5"), osmomath.MustNewDecFromStr("0.5")}),
		},
		"multiple denomShareAgreement denoms": {
			setupFunc: func() []string {
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
				return []string{denomA, denomB}
			},
			expectedDenomShares:   defaultTakerFeeShareAgreements[:2],
			expectedAlloyedShares: []types.TakerFeeShareAgreement{},
		},
		"multiple alloyedAssetShareAgreement denoms": {
			setupFunc: func() []string {
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements)
				alloyedDenom1 := s.setupAndRegisterAlloyedPool([]string{denomA, denomC}, []uint16{1, 1})
				alloyedDenom2 := s.setupAndRegisterAlloyedPool([]string{denomB, denomC}, []uint16{1, 1})
				return []string{alloyedDenom1, alloyedDenom2}
			},
			expectedDenomShares: []types.TakerFeeShareAgreement{},
			expectedAlloyedShares: modifySkimPercent([]types.TakerFeeShareAgreement{
				defaultTakerFeeShareAgreements[0], defaultTakerFeeShareAgreements[2], defaultTakerFeeShareAgreements[1], defaultTakerFeeShareAgreements[2],
			}, []osmomath.Dec{
				osmomath.MustNewDecFromStr("0.5"),
				osmomath.MustNewDecFromStr("0.5"),
				osmomath.MustNewDecFromStr("0.5"),
				osmomath.MustNewDecFromStr("0.5"),
			}),
		},
		"multiple denomShareAgreement denoms and multiple alloyedAssetShareAgreements": {
			setupFunc: func() []string {
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
				alloyedDenom1 := s.setupAndRegisterAlloyedPool([]string{denomA, denomC}, []uint16{1, 1})
				alloyedDenom2 := s.setupAndRegisterAlloyedPool([]string{denomB, denomC}, []uint16{1, 1})
				return []string{denomA, denomB, alloyedDenom1, alloyedDenom2}
			},
			expectedDenomShares:   defaultTakerFeeShareAgreements[:2],
			expectedAlloyedShares: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{osmomath.MustNewDecFromStr("0.5"), osmomath.MustNewDecFromStr("0.5")}),
		},
		"multiple denomShareAgreement denoms and multiple alloyedAssetShareAgreements, alloyed denoms first": {
			setupFunc: func() []string {
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
				alloyedDenom1 := s.setupAndRegisterAlloyedPool([]string{denomA, denomC}, []uint16{1, 1})
				alloyedDenom2 := s.setupAndRegisterAlloyedPool([]string{denomB, denomC}, []uint16{1, 1})
				return []string{alloyedDenom1, alloyedDenom2, denomA, denomB}
			},
			expectedDenomShares:   defaultTakerFeeShareAgreements[:2],
			expectedAlloyedShares: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{osmomath.MustNewDecFromStr("0.5"), osmomath.MustNewDecFromStr("0.5")}),
		},
		"no agreements": {
			setupFunc: func() []string {
				return []string{denomA, denomB}
			},
			expectedDenomShares:   []types.TakerFeeShareAgreement{},
			expectedAlloyedShares: []types.TakerFeeShareAgreement{},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			denomsInvolvedInRoute := tc.setupFunc()

			denomShares, alloyedShares := s.App.PoolManagerKeeper.GetTakerFeeShareAgreements(denomsInvolvedInRoute)
			s.Require().Equal(tc.expectedDenomShares, denomShares)
			s.Require().Equal(tc.expectedAlloyedShares, alloyedShares)
		})
	}
}

func (s *KeeperTestSuite) TestProcessDenomShareAgreements() {
	tests := map[string]struct {
		denomShareAgreements []types.TakerFeeShareAgreement
		totalTakerFees       sdk.Coins
		expectedAccumulators []types.TakerFeeSkimAccumulator
		expectedError        error
	}{
		"valid denomShareAgreements": {
			denomShareAgreements: []types.TakerFeeShareAgreement{
				{Denom: denomA, SkimPercent: osmomath.MustNewDecFromStr("0.01")},
			},
			totalTakerFees: sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(1000000000))),
			expectedAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            denomA,
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(10000000))),
				},
			},
			expectedError: nil,
		},
		"invalid denomShareAgreements with percentage > 1": {
			denomShareAgreements: []types.TakerFeeShareAgreement{
				{Denom: denomA, SkimPercent: osmomath.MustNewDecFromStr("1.01")},
			},
			totalTakerFees:       sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(1000000000))),
			expectedAccumulators: []types.TakerFeeSkimAccumulator{},
			expectedError:        types.InvalidTakerFeeSharePercentageError{Percentage: osmomath.MustNewDecFromStr("1.01")},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			err := s.App.PoolManagerKeeper.ProcessShareAgreements(s.Ctx, tc.denomShareAgreements, tc.totalTakerFees)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError, err)
				return
			}
			s.Require().NoError(err)

			takerFeeShareAccumulators, err := s.App.PoolManagerKeeper.GetAllTakerFeeShareAccumulators(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedAccumulators, takerFeeShareAccumulators)
		})
	}
}

func (s *KeeperTestSuite) TestProcessAlloyedAssetShareAgreements() {
	tests := map[string]struct {
		alloyedAssetShareAgreements []types.TakerFeeShareAgreement
		totalTakerFees              sdk.Coins
		expectedAccumulators        []types.TakerFeeSkimAccumulator
		expectedError               error
	}{
		"valid alloyedAssetShareAgreements": {
			alloyedAssetShareAgreements: []types.TakerFeeShareAgreement{
				{Denom: denomA, SkimPercent: osmomath.MustNewDecFromStr("0.01")},
			},
			totalTakerFees: sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(1000000000))),
			expectedAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            denomA,
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(10000000))),
				},
			},
			expectedError: nil,
		},
		"invalid alloyedAssetShareAgreements with percentage > 1": {
			alloyedAssetShareAgreements: []types.TakerFeeShareAgreement{
				{Denom: denomA, SkimPercent: osmomath.MustNewDecFromStr("1.01")},
			},
			totalTakerFees:       sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(1000000000))),
			expectedAccumulators: []types.TakerFeeSkimAccumulator{},
			expectedError:        types.InvalidTakerFeeSharePercentageError{Percentage: osmomath.MustNewDecFromStr("1.01")},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			err := s.App.PoolManagerKeeper.ProcessShareAgreements(s.Ctx, tc.alloyedAssetShareAgreements, tc.totalTakerFees)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError, err)
				return
			}
			s.Require().NoError(err)

			takerFeeShareAccumulators, err := s.App.PoolManagerKeeper.GetAllTakerFeeShareAccumulators(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedAccumulators, takerFeeShareAccumulators)
		})
	}
}

func (s *KeeperTestSuite) TestValidatePercentage() {
	tests := map[string]struct {
		percentage    osmomath.Dec
		expectedError error
	}{
		"valid percentage": {
			percentage:    osmomath.MustNewDecFromStr("0.5"),
			expectedError: nil,
		},
		"percentage greater than 1": {
			percentage:    osmomath.MustNewDecFromStr("1.01"),
			expectedError: types.InvalidTakerFeeSharePercentageError{Percentage: osmomath.MustNewDecFromStr("1.01")},
		},
		"negative percentage": {
			percentage:    osmomath.MustNewDecFromStr("-0.01"),
			expectedError: types.InvalidTakerFeeSharePercentageError{Percentage: osmomath.MustNewDecFromStr("-0.01")},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			err := s.App.PoolManagerKeeper.ValidatePercentage(tc.percentage)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError, err)
				return
			}
			s.Require().NoError(err)
		})
	}
}

func (s *KeeperTestSuite) setupAndRegisterAlloyedPool(denoms []string, ratios []uint16) string {
	cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], denoms, ratios)
	err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
	s.Require().NoError(err)
	return createAlloyedDenom(cwPool.GetAddress().String())
}

func (s *KeeperTestSuite) setupTakerFeeShareAgreement(denom string, skimPercent string, skimAddress string) {
	s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
		Denom:       denom,
		SkimPercent: osmomath.MustNewDecFromStr(skimPercent),
		SkimAddress: skimAddress,
	})
}
