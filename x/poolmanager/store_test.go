package poolmanager_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v25/app/apptesting"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v25/x/cosmwasmpool/types"
	"github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

var DefaultTakerFeeShareAgreements = []types.TakerFeeShareAgreement{
	{
		Denom:       "uosmo",
		SkimPercent: osmomath.MustNewDecFromStr("0.01"),
		SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
	},
	{
		Denom:       "stake",
		SkimPercent: osmomath.MustNewDecFromStr("0.02"),
		SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
	},
	{
		Denom:       "nBTC",
		SkimPercent: osmomath.MustNewDecFromStr("0.03"),
		SkimAddress: "osmo1jermpr9yust7cyhfjme3cr08kt6n8jv6p35l39",
	},
}

func (s *KeeperTestSuite) TestGetAllTakerFeeShareAgreementsMap() {
	tests := map[string]struct {
		setupFunc                         func()
		expectedTakerFeeShareAgreementMap func() map[string]types.TakerFeeShareAgreement
	}{
		"single taker fee share agreement": {
			setupFunc: func() {
				numTakerFeeShareAgreements := 1
				for i, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					if i >= numTakerFeeShareAgreements {
						break
					}
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			expectedTakerFeeShareAgreementMap: func() map[string]types.TakerFeeShareAgreement {
				numTakerFeeShareAgreements := 1
				expectedTakerFeeShareAgreementMap := make(map[string]types.TakerFeeShareAgreement)
				for i, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					if i >= numTakerFeeShareAgreements {
						break
					}
					expectedTakerFeeShareAgreementMap[takerFeeShareAgreement.Denom] = takerFeeShareAgreement
				}
				return expectedTakerFeeShareAgreementMap
			},
		},
		"multiple taker fee share agreements": {
			setupFunc: func() {
				for _, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			expectedTakerFeeShareAgreementMap: func() map[string]types.TakerFeeShareAgreement {
				expectedTakerFeeShareAgreementMap := make(map[string]types.TakerFeeShareAgreement)
				for _, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					expectedTakerFeeShareAgreementMap[takerFeeShareAgreement.Denom] = takerFeeShareAgreement
				}
				return expectedTakerFeeShareAgreementMap
			},
		},
		"no taker fee share agreements": {
			setupFunc: func() {},
			expectedTakerFeeShareAgreementMap: func() map[string]types.TakerFeeShareAgreement {
				return make(map[string]types.TakerFeeShareAgreement)
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()
			expectedTakerFeeShareAgreementMap := tc.expectedTakerFeeShareAgreementMap()

			takerFeeShareAgreementMap, err := s.App.PoolManagerKeeper.GetAllTakerFeeShareAgreementsMap(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(expectedTakerFeeShareAgreementMap, takerFeeShareAgreementMap, "GetAllTakerFeeShareAgreementsMap() = %v, want %v", takerFeeShareAgreementMap, tc.expectedTakerFeeShareAgreementMap)
		})
	}
}

func (s *KeeperTestSuite) TestGetAllTakerFeesShareAgreements() {
	tests := map[string]struct {
		setupFunc                       func()
		expectedTakerFeeShareAgreements func() []types.TakerFeeShareAgreement
	}{
		"single taker fee share agreement": {
			setupFunc: func() {
				numTakerFeeShareAgreements := 1
				for i, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					if i >= numTakerFeeShareAgreements {
						break
					}
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			expectedTakerFeeShareAgreements: func() []types.TakerFeeShareAgreement {
				numTakerFeeShareAgreements := 1
				var expectedTakerFeeShareAgreements []types.TakerFeeShareAgreement
				for i, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					if i >= numTakerFeeShareAgreements {
						break
					}
					expectedTakerFeeShareAgreements = append(expectedTakerFeeShareAgreements, takerFeeShareAgreement)
				}
				return reverseSlice(expectedTakerFeeShareAgreements)
			},
		},
		"multiple taker fee share agreements": {
			setupFunc: func() {
				for _, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			expectedTakerFeeShareAgreements: func() []types.TakerFeeShareAgreement {
				return reverseSlice(DefaultTakerFeeShareAgreements)
			},
		},
		"no taker fee share agreements": {
			setupFunc: func() {},
			expectedTakerFeeShareAgreements: func() []types.TakerFeeShareAgreement {
				return []types.TakerFeeShareAgreement{}
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()
			expectedTakerFeeShareAgreements := tc.expectedTakerFeeShareAgreements()

			takerFeeShareAgreements := s.App.PoolManagerKeeper.GetAllTakerFeesShareAgreements(s.Ctx)
			s.Require().Equal(expectedTakerFeeShareAgreements, takerFeeShareAgreements, "GetAllTakerFeesShareAgreements() = %v, want %v", takerFeeShareAgreements, expectedTakerFeeShareAgreements)
		})
	}
}

func (s *KeeperTestSuite) TestSetTakerFeeShareAgreementsMapCached() {
	tests := map[string]struct {
		setupFunc                       func()
		expectedTakerFeeShareAgreements func() map[string]types.TakerFeeShareAgreement
		expectErr                       bool
	}{
		"single taker fee share agreement": {
			setupFunc: func() {
				numTakerFeeShareAgreements := 1
				for i, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					if i >= numTakerFeeShareAgreements {
						break
					}
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			expectedTakerFeeShareAgreements: func() map[string]types.TakerFeeShareAgreement {
				numTakerFeeShareAgreements := 1
				expectedTakerFeeShareAgreements := make(map[string]types.TakerFeeShareAgreement)
				for i, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					if i >= numTakerFeeShareAgreements {
						break
					}
					expectedTakerFeeShareAgreements[takerFeeShareAgreement.Denom] = takerFeeShareAgreement
				}
				return expectedTakerFeeShareAgreements
			},
			expectErr: false,
		},
		"multiple taker fee share agreements": {
			setupFunc: func() {
				for _, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			expectedTakerFeeShareAgreements: func() map[string]types.TakerFeeShareAgreement {
				expectedTakerFeeShareAgreements := make(map[string]types.TakerFeeShareAgreement)
				for _, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					expectedTakerFeeShareAgreements[takerFeeShareAgreement.Denom] = takerFeeShareAgreement
				}
				return expectedTakerFeeShareAgreements
			},
			expectErr: false,
		},
		"no taker fee share agreements": {
			setupFunc: func() {},
			expectedTakerFeeShareAgreements: func() map[string]types.TakerFeeShareAgreement {
				return make(map[string]types.TakerFeeShareAgreement)
			},
			expectErr: false,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()
			expectedTakerFeeShareAgreements := tc.expectedTakerFeeShareAgreements()

			err := s.App.PoolManagerKeeper.SetTakerFeeShareAgreementsMapCached(s.Ctx)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				cachedTakerFeeShareAgreement, _, _ := s.App.PoolManagerKeeper.GetCachedMaps()
				s.Require().Equal(expectedTakerFeeShareAgreements, cachedTakerFeeShareAgreement, "cachedTakerFeeShareAgreement = %v, want %v", cachedTakerFeeShareAgreement, expectedTakerFeeShareAgreements)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSetTakerFeeShareAgreementForDenom() {
	tests := map[string]struct {
		setupFunc                      func()
		denomToRequest                 string
		expectedTakerFeeShareAgreement types.TakerFeeShareAgreement
		expectedFound                  bool
	}{
		"set one taker fee share agreement, get same taker fee share agreement": {
			setupFunc: func() {
				numTakerFeeShareAgreements := 1
				for i, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					if i >= numTakerFeeShareAgreements {
						break
					}
					err := s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
					s.Require().NoError(err)
				}
			},
			denomToRequest:                 "uosmo",
			expectedTakerFeeShareAgreement: DefaultTakerFeeShareAgreements[0],
			expectedFound:                  true,
		},
		"set three taker fee share agreements, get one of the three taker fee share agreements": {
			setupFunc: func() {
				for _, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			denomToRequest:                 "stake",
			expectedTakerFeeShareAgreement: DefaultTakerFeeShareAgreements[1],
			expectedFound:                  true,
		},
		"set three taker fee share agreements, attempt to get taker fee share agreement for denom that does not exist": {
			setupFunc: func() {
				for _, takerFeeShareAgreement := range DefaultTakerFeeShareAgreements {
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			denomToRequest: "denomNotAdded",
			expectedFound:  false,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			tc.setupFunc()

			takerFeeShareAgreements, found := s.App.PoolManagerKeeper.GetTakerFeeShareAgreementFromDenom(s.Ctx, tc.denomToRequest)
			if tc.expectedFound {
				s.Require().True(found)
				s.Require().Equal(tc.expectedTakerFeeShareAgreement, takerFeeShareAgreements)
			} else {
				s.Require().False(found)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetTakerFeeShareDenomsToAccruedValue() {
	tests := map[string]struct {
		setupFunc     func()
		tierDenom     string
		takerFeeDenom string
		expectedValue osmomath.Int
		expectedError error
	}{
		"tier denom accrued denom value, retrieve denom value": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "stake", osmomath.NewInt(100))
				s.Require().NoError(err)
			},
			tierDenom:     "uosmo",
			takerFeeDenom: "stake",
			expectedValue: osmomath.NewInt(100),
		},
		"tier denom did not accrue value, nothing to retrieve, so not found": {
			setupFunc:     func() {},
			tierDenom:     "uosmo",
			takerFeeDenom: "stake",
			expectedError: fmt.Errorf("no accrued value found for tierDenom uosmo and takerFeeDenom stake"),
		},
		"tier denom accrued denom value, retrieve different denom value, so not found": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "stake", osmomath.NewInt(100))
				s.Require().NoError(err)
			},
			tierDenom:     "uosmo",
			takerFeeDenom: "nonExistentDenom",
			expectedError: fmt.Errorf("no accrued value found for tierDenom uosmo and takerFeeDenom nonExistentDenom"),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()

			value, err := s.App.PoolManagerKeeper.GetTakerFeeShareDenomsToAccruedValue(s.Ctx, tc.tierDenom, tc.takerFeeDenom)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError.Error(), err.Error())
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedValue, value)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSetTakerFeeShareDenomsToAccruedValue() {
	tests := map[string]struct {
		setupFunc     func()
		tierDenom     string
		takerFeeDenom string
		accruedValue  osmomath.Int
		expectErr     bool
	}{
		"set value that was previously non existent": {
			setupFunc:     func() {},
			tierDenom:     "uosmo",
			takerFeeDenom: "stake",
			accruedValue:  osmomath.NewInt(100),
			expectErr:     false,
		},
		"set value lower than what it previously was, should override": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "stake", osmomath.NewInt(200))
				s.Require().NoError(err)
			},
			tierDenom:     "uosmo",
			takerFeeDenom: "stake",
			accruedValue:  osmomath.NewInt(100),
			expectErr:     false,
		},
		"set value greater than what it previously was, should override": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "stake", osmomath.NewInt(200))
				s.Require().NoError(err)
			},
			tierDenom:     "uosmo",
			takerFeeDenom: "stake",
			accruedValue:  osmomath.NewInt(300),
			expectErr:     false,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()

			err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, tc.tierDenom, tc.takerFeeDenom, tc.accruedValue)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				value, err := s.App.PoolManagerKeeper.GetTakerFeeShareDenomsToAccruedValue(s.Ctx, tc.tierDenom, tc.takerFeeDenom)
				s.Require().NoError(err)
				s.Require().Equal(tc.accruedValue, value, "GetTakerFeeShareDenomsToAccruedValue() = %v, want %v", value, tc.accruedValue)
			}
		})
	}
}

func (s *KeeperTestSuite) TestIncreaseTakerFeeShareDenomsToAccruedValue() {
	tests := map[string]struct {
		setupFunc     func()
		tierDenom     string
		takerFeeDenom string
		additiveValue osmomath.Int
		expectedValue osmomath.Int
		expectErr     bool
	}{
		"increase value that was previously non existent": {
			setupFunc:     func() {},
			tierDenom:     "uosmo",
			takerFeeDenom: "stake",
			additiveValue: osmomath.NewInt(100),
			expectedValue: osmomath.NewInt(100),
			expectErr:     false,
		},
		"increase value that was previously set": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "stake", osmomath.NewInt(200))
				s.Require().NoError(err)
			},
			tierDenom:     "uosmo",
			takerFeeDenom: "stake",
			additiveValue: osmomath.NewInt(100),
			expectedValue: osmomath.NewInt(300),
			expectErr:     false,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()

			err := s.App.PoolManagerKeeper.IncreaseTakerFeeShareDenomsToAccruedValue(s.Ctx, tc.tierDenom, tc.takerFeeDenom, tc.additiveValue)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				value, err := s.App.PoolManagerKeeper.GetTakerFeeShareDenomsToAccruedValue(s.Ctx, tc.tierDenom, tc.takerFeeDenom)
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedValue, value, "GetTakerFeeShareDenomsToAccruedValue() = %v, want %v", value, tc.expectedValue)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetAllTakerFeeShareAccumulators() {
	tests := map[string]struct {
		setupFunc            func()
		expectedAccumulators []types.TakerFeeSkimAccumulator
	}{
		"no accumulators": {
			setupFunc:            func() {},
			expectedAccumulators: []types.TakerFeeSkimAccumulator{},
		},
		"single accumulator": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "stake", osmomath.NewInt(100))
				s.Require().NoError(err)
			},
			expectedAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            "uosmo",
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin("stake", osmomath.NewInt(100))),
				},
			},
		},
		"multiple accumulators": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "stake", osmomath.NewInt(100))
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "atom", osmomath.NewInt(200))
				s.Require().NoError(err)
			},
			expectedAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            "uosmo",
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin("stake", osmomath.NewInt(100)), sdk.NewCoin("atom", osmomath.NewInt(200))),
				},
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()

			accumulators := s.App.PoolManagerKeeper.GetAllTakerFeeShareAccumulators(s.Ctx)
			s.Require().Equal(tc.expectedAccumulators, accumulators, "GetAllTakerFeeShareAccumulators() = %v, want %v", accumulators, tc.expectedAccumulators)
		})
	}
}

func (s *KeeperTestSuite) TestDeleteAllTakerFeeShareAccumulatorsForTierDenom() {
	tests := map[string]struct {
		setupFunc                   func()
		tierDenom                   string
		expectedNumAccumAfterDelete uint64
	}{
		"delete non-existent accumulators": {
			setupFunc:                   func() {},
			tierDenom:                   "uosmo",
			expectedNumAccumAfterDelete: 0,
		},
		"delete existing accumulators": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "stake", osmomath.NewInt(100))
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "atom", osmomath.NewInt(200))
				s.Require().NoError(err)
			},
			tierDenom:                   "uosmo",
			expectedNumAccumAfterDelete: 0,
		},
		"delete existing accumulators for non existent tier denom": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "stake", osmomath.NewInt(100))
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "atom", osmomath.NewInt(200))
				s.Require().NoError(err)
			},
			tierDenom:                   "nonExistentTierDenom",
			expectedNumAccumAfterDelete: 1,
		},
		"delete existing accumulators for tier denom with accumulators": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "stake", osmomath.NewInt(100))
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "uosmo", "atom", osmomath.NewInt(200))
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "nBTC", "stake", osmomath.NewInt(100))
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, "nBTC", "atom", osmomath.NewInt(200))
				s.Require().NoError(err)
			},
			tierDenom:                   "nBTC",
			expectedNumAccumAfterDelete: 1,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()

			s.App.PoolManagerKeeper.DeleteAllTakerFeeShareAccumulatorsForTierDenom(s.Ctx, tc.tierDenom)

			accumulators := s.App.PoolManagerKeeper.GetAllTakerFeeShareAccumulators(s.Ctx)
			s.Require().Len(accumulators, int(tc.expectedNumAccumAfterDelete))
		})
	}
}

func (s *KeeperTestSuite) TestSetRegisteredAlloyedPool() {
	const (
		GammPool = iota
		ConcentratedPool
		CWPool
		AlloyedPool
		NonExistentPool
	)

	tests := map[string]struct {
		poolType                        int
		poolId                          uint64
		preSetFunc                      func(ctx sdk.Context)
		postSetFunc                     func(ctx sdk.Context)
		expectedTakerFeeShareAgreements []types.TakerFeeShareAgreement
		expectedError                   error
	}{
		"set non-existent pool": {
			poolType:      NonExistentPool,
			preSetFunc:    func(ctx sdk.Context) {},
			postSetFunc:   func(ctx sdk.Context) {},
			expectedError: types.FailedToFindRouteError{PoolId: 100},
		},
		"set gamm pool": {
			poolType:      GammPool,
			preSetFunc:    func(ctx sdk.Context) {},
			postSetFunc:   func(ctx sdk.Context) {},
			expectedError: fmt.Errorf("pool with id %d is not a CosmWasmPool", 2),
		},
		"set concentrated pool": {
			poolType:      ConcentratedPool,
			preSetFunc:    func(ctx sdk.Context) {},
			postSetFunc:   func(ctx sdk.Context) {},
			expectedError: fmt.Errorf("pool with id %d is not a CosmWasmPool", 1),
		},
		"set cw pool": {
			poolType:      CWPool,
			preSetFunc:    func(ctx sdk.Context) {},
			postSetFunc:   func(ctx sdk.Context) {},
			expectedError: fmt.Errorf("third part of alloyedDenom should be 'alloyed'"),
		},
		"set alloyed pool, with one taker fee share agreement set before alloyed pool is registered and one set after": {
			poolType: AlloyedPool,
			preSetFunc: func(ctx sdk.Context) {
				takerFeeShareAgreement := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(ctx, takerFeeShareAgreement)
			},
			postSetFunc: func(ctx sdk.Context) {
				takerFeeShareAgreement := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomB,
					SkimPercent: osmomath.MustNewDecFromStr("0.02"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(ctx, takerFeeShareAgreement)
			},
			expectedTakerFeeShareAgreements: []types.TakerFeeShareAgreement{
				{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5")),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				},
				{
					Denom:       apptesting.DefaultTransmuterDenomB,
					SkimPercent: osmomath.MustNewDecFromStr("0.02").Mul(osmomath.MustNewDecFromStr("0.5")),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				},
			},
		},
		"set alloyed pool, with both taker fee share agreements set before alloyed pool is registered": {
			poolType: AlloyedPool,
			preSetFunc: func(ctx sdk.Context) {
				takerFeeShareAgreement := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(ctx, takerFeeShareAgreement)
				takerFeeShareAgreement = types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomB,
					SkimPercent: osmomath.MustNewDecFromStr("0.02"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(ctx, takerFeeShareAgreement)
			},
			postSetFunc: func(ctx sdk.Context) {},
			expectedTakerFeeShareAgreements: []types.TakerFeeShareAgreement{
				{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5")),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				},
				{
					Denom:       apptesting.DefaultTransmuterDenomB,
					SkimPercent: osmomath.MustNewDecFromStr("0.02").Mul(osmomath.MustNewDecFromStr("0.5")),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				},
			},
		},
		"set alloyed pool, with both taker fee share agreements set after alloyed pool is registered": {
			poolType:   AlloyedPool,
			preSetFunc: func(ctx sdk.Context) {},
			postSetFunc: func(ctx sdk.Context) {
				takerFeeShareAgreement := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(ctx, takerFeeShareAgreement)
				takerFeeShareAgreement = types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomB,
					SkimPercent: osmomath.MustNewDecFromStr("0.02"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(ctx, takerFeeShareAgreement)
			},
			expectedTakerFeeShareAgreements: []types.TakerFeeShareAgreement{
				{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5")),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				},
				{
					Denom:       apptesting.DefaultTransmuterDenomB,
					SkimPercent: osmomath.MustNewDecFromStr("0.02").Mul(osmomath.MustNewDecFromStr("0.5")),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				},
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			poolInfos := s.PrepareAllSupportedPools()
			if tc.poolType == GammPool {
				tc.poolId = poolInfos.BalancerPoolID
			} else if tc.poolType == ConcentratedPool {
				tc.poolId = poolInfos.ConcentratedPoolID
			} else if tc.poolType == CWPool {
				tc.poolId = poolInfos.CosmWasmPoolID
			} else if tc.poolType == AlloyedPool {
				tc.poolId = poolInfos.AlloyedPoolID
			} else {
				tc.poolId = 100
			}

			tc.preSetFunc(s.Ctx)
			err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, tc.poolId)
			tc.postSetFunc(s.Ctx)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError.Error(), err.Error())
			} else {
				s.Require().NoError(err)
				alloyedDenom, shareState, err := s.App.PoolManagerKeeper.GetRegisteredAlloyedPoolFromPoolId(s.Ctx, tc.poolId)
				s.Require().NoError(err)

				alloyedPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, tc.poolId)
				s.Require().NoError(err)
				expectedAlloyedDenom := fmt.Sprintf("factory/%s/alloyed/testdenom", alloyedPool.GetAddress().String())
				s.Require().Equal(expectedAlloyedDenom, alloyedDenom)
				expectedShareState := types.AlloyContractTakerFeeShareState{
					ContractAddress:         alloyedPool.GetAddress().String(),
					TakerFeeShareAgreements: tc.expectedTakerFeeShareAgreements,
				}
				s.Require().Equal(expectedShareState, shareState)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetRegisteredAlloyedPoolFromDenom() {
	tests := map[string]struct {
		setupFunc                       func()
		getDenomFromPool                bool
		denom                           string
		expectFound                     bool
		expectedTakerFeeShareAgreements []types.TakerFeeShareAgreement
	}{
		"get non-existent pool": {
			setupFunc:   func() {},
			denom:       "nonExistent",
			expectFound: false,
		},
		"setup alloyed pool, get the alloyed pool denom, no taker fee share agreements": {
			setupFunc:                       func() {},
			getDenomFromPool:                true,
			expectFound:                     true,
			expectedTakerFeeShareAgreements: []types.TakerFeeShareAgreement(nil),
		},
		"setup alloyed pool, get the alloyed pool denom, with one taker fee share agreement": {
			setupFunc: func() {
				takerFeeShareAgreement := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
			},
			getDenomFromPool: true,
			expectFound:      true,
			expectedTakerFeeShareAgreements: []types.TakerFeeShareAgreement{
				{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5")),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				},
			},
		},
		"setup alloyed pool, get the alloyed pool denom, with multiple taker fee share agreements": {
			setupFunc: func() {
				takerFeeShareAgreement := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				takerFeeShareAgreement = types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomB,
					SkimPercent: osmomath.MustNewDecFromStr("0.02"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
			},
			getDenomFromPool: true,
			expectFound:      true,
			expectedTakerFeeShareAgreements: []types.TakerFeeShareAgreement{
				{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5")),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				},
				{
					Denom:       apptesting.DefaultTransmuterDenomB,
					SkimPercent: osmomath.MustNewDecFromStr("0.02").Mul(osmomath.MustNewDecFromStr("0.5")),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				},
			},
		},
		"setup alloyed pool, get the alloyed pool denom, with unrelated taker fee share agreement": {
			setupFunc: func() {
				takerFeeShareAgreement := types.TakerFeeShareAgreement{
					Denom:       "unrelatedDenom",
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
			},
			getDenomFromPool:                true,
			expectFound:                     true,
			expectedTakerFeeShareAgreements: []types.TakerFeeShareAgreement(nil),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()

			poolInfos := s.PrepareAllSupportedPools()
			alloyedPoolID := poolInfos.AlloyedPoolID
			alloyedPool, err := s.App.CosmwasmPoolKeeper.GetPool(s.Ctx, alloyedPoolID)

			if tc.getDenomFromPool {
				s.Require().NoError(err)
				tc.denom = fmt.Sprintf("factory/%s/alloyed/testdenom", alloyedPool.GetAddress().String())
			}

			s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)

			shareState, found := s.App.PoolManagerKeeper.GetRegisteredAlloyedPoolFromDenom(s.Ctx, tc.denom)
			s.Require().Equal(tc.expectFound, found)
			if tc.expectFound {
				expectedShareState := types.AlloyContractTakerFeeShareState{
					ContractAddress:         alloyedPool.GetAddress().String(),
					TakerFeeShareAgreements: tc.expectedTakerFeeShareAgreements,
				}
				s.Require().Equal(expectedShareState, shareState, "GetRegisteredAlloyedPoolFromDenom() = %v, want %v", shareState, expectedShareState)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetRegisteredAlloyedPoolFromPoolId() {
	tests := map[string]struct {
		setupFunc                       func()
		poolId                          uint64
		expectedError                   error
		expectedTakerFeeShareAgreements []types.TakerFeeShareAgreement
	}{
		"get non-existent pool": {
			setupFunc:     func() {},
			poolId:        1,
			expectedError: fmt.Errorf("no registered alloyed pool found for poolId %d", 1),
		},
		"get existing alloyed pool": {
			setupFunc: func() {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
			},
			poolId:                          5,
			expectedTakerFeeShareAgreements: []types.TakerFeeShareAgreement(nil),
		},
		"get existing alloyed pool with one taker fee share agreement": {
			setupFunc: func() {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)

				takerFeeShareAgreement := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
			},
			poolId: 5,
			expectedTakerFeeShareAgreements: []types.TakerFeeShareAgreement{
				{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5")),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				},
			},
		},
		"get existing non-alloyed pool": {
			setupFunc: func() {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
			},
			poolId:        1,
			expectedError: fmt.Errorf("no registered alloyed pool found for poolId %d", 1),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()

			_, shareState, err := s.App.PoolManagerKeeper.GetRegisteredAlloyedPoolFromPoolId(s.Ctx, tc.poolId)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError.Error(), err.Error())
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedTakerFeeShareAgreements, shareState.TakerFeeShareAgreements, "GetRegisteredAlloyedPoolFromPoolId() = %v, want %v", shareState.TakerFeeShareAgreements, tc.expectedTakerFeeShareAgreements)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetAllRegisteredAlloyedPools() {
	tests := map[string]struct {
		setupFunc                       func()
		expectedTakerFeeShareAgreements [][]types.TakerFeeShareAgreement
	}{
		"no registered pools": {
			setupFunc:                       func() {},
			expectedTakerFeeShareAgreements: [][]types.TakerFeeShareAgreement(nil),
		},
		"single registered pool, no taker fee share agreements": {
			setupFunc: func() {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
			},
			expectedTakerFeeShareAgreements: [][]types.TakerFeeShareAgreement{
				[]types.TakerFeeShareAgreement(nil),
			},
		},
		"multiple registered pools, no taker fee share agreements": {
			setupFunc: func() {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3(s.TestAccs[0], []string{"testA", "testB"}, "osmosis", "x/cosmwasmpool/bytecode")
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
			},
			expectedTakerFeeShareAgreements: [][]types.TakerFeeShareAgreement{
				[]types.TakerFeeShareAgreement(nil),
				[]types.TakerFeeShareAgreement(nil),
			},
		},
		"single registered pool, with one taker fee share agreement": {
			setupFunc: func() {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)

				takerFeeShareAgreement := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
			},
			expectedTakerFeeShareAgreements: [][]types.TakerFeeShareAgreement{
				{
					{
						Denom:       apptesting.DefaultTransmuterDenomA,
						SkimPercent: osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5")),
						SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
					},
				},
			},
		},
		"multiple registered pools, with multiple taker fee share agreements": {
			setupFunc: func() {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)

				takerFeeShareAgreement := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)

				takerFeeShareAgreement = types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomB,
					SkimPercent: osmomath.MustNewDecFromStr("0.02"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)

				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3(s.TestAccs[0], []string{"testA", "testB"}, "osmosis", "x/cosmwasmpool/bytecode")
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)

				takerFeeShareAgreement = types.TakerFeeShareAgreement{
					Denom:       "testA",
					SkimPercent: osmomath.MustNewDecFromStr("0.03"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
			},
			expectedTakerFeeShareAgreements: [][]types.TakerFeeShareAgreement{
				{
					{
						Denom:       apptesting.DefaultTransmuterDenomA,
						SkimPercent: osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5")),
						SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
					},
					{
						Denom:       apptesting.DefaultTransmuterDenomB,
						SkimPercent: osmomath.MustNewDecFromStr("0.02").Mul(osmomath.MustNewDecFromStr("0.5")),
						SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
					},
				},
				{
					{
						Denom:       "testA",
						SkimPercent: osmomath.MustNewDecFromStr("0.03").Mul(osmomath.MustNewDecFromStr("0.5")),
						SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
					},
				},
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			tc.setupFunc()

			shareStates, err := s.App.PoolManagerKeeper.GetAllRegisteredAlloyedPools(s.Ctx)
			s.Require().NoError(err)

			for i, shareState := range shareStates {
				s.Require().Equal(tc.expectedTakerFeeShareAgreements[i], shareState.TakerFeeShareAgreements)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetAllRegisteredAlloyedPoolsMap() {
	tests := map[string]struct {
		setupFunc func() map[string]types.AlloyContractTakerFeeShareState
	}{
		"no registered pools": {
			setupFunc: func() map[string]types.AlloyContractTakerFeeShareState {
				return map[string]types.AlloyContractTakerFeeShareState{}
			},
		},
		"single registered pool, no taker fee share agreements": {
			setupFunc: func() map[string]types.AlloyContractTakerFeeShareState {
				poolInfos := s.PrepareAllSupportedPools()
				pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfos.AlloyedPoolID)
				denom := fmt.Sprintf("factory/%s/alloyed/testdenom", pool.GetAddress().String())
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				return map[string]types.AlloyContractTakerFeeShareState{
					denom: {
						ContractAddress:         pool.GetAddress().String(),
						TakerFeeShareAgreements: []types.TakerFeeShareAgreement(nil),
					},
				}
			},
		},
		"multiple registered pools, no taker fee share agreements": {
			setupFunc: func() map[string]types.AlloyContractTakerFeeShareState {
				poolInfos := s.PrepareAllSupportedPools()
				pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3(s.TestAccs[0], []string{"testA", "testB"}, "osmosis", "x/cosmwasmpool/bytecode")
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				denomA := fmt.Sprintf("factory/%s/alloyed/testdenom", pool.GetAddress().String())
				denomB := fmt.Sprintf("factory/%s/alloyed/testdenom", cwPool.GetAddress().String())
				return map[string]types.AlloyContractTakerFeeShareState{
					denomA: {
						ContractAddress:         pool.GetAddress().String(),
						TakerFeeShareAgreements: []types.TakerFeeShareAgreement(nil),
					},
					denomB: {
						ContractAddress:         cwPool.GetAddress().String(),
						TakerFeeShareAgreements: []types.TakerFeeShareAgreement(nil),
					},
				}
			},
		},
		"single registered pool, with one taker fee share agreement": {
			setupFunc: func() map[string]types.AlloyContractTakerFeeShareState {
				poolInfos := s.PrepareAllSupportedPools()
				pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				denom := fmt.Sprintf("factory/%s/alloyed/testdenom", pool.GetAddress().String())
				takerFeeShareAgreement := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				return map[string]types.AlloyContractTakerFeeShareState{
					denom: {
						ContractAddress: pool.GetAddress().String(),
						TakerFeeShareAgreements: []types.TakerFeeShareAgreement{{
							Denom:       apptesting.DefaultTransmuterDenomA,
							SkimPercent: osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5")),
							SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
						}},
					},
				}
			},
		},
		"multiple registered pools, with multiple taker fee share agreements": {
			setupFunc: func() map[string]types.AlloyContractTakerFeeShareState {
				poolInfos := s.PrepareAllSupportedPools()
				pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				denomA := fmt.Sprintf("factory/%s/alloyed/testdenom", pool.GetAddress().String())
				takerFeeShareAgreementA := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreementA)
				takerFeeShareAgreementB := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomB,
					SkimPercent: osmomath.MustNewDecFromStr("0.02"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreementB)
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3(s.TestAccs[0], []string{"testA", "testB"}, "osmosis", "x/cosmwasmpool/bytecode")
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				denomB := fmt.Sprintf("factory/%s/alloyed/testdenom", cwPool.GetAddress().String())
				takerFeeShareAgreementTestA := types.TakerFeeShareAgreement{
					Denom:       "testA",
					SkimPercent: osmomath.MustNewDecFromStr("0.03"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreementTestA)
				return map[string]types.AlloyContractTakerFeeShareState{
					denomA: {
						ContractAddress: pool.GetAddress().String(),
						TakerFeeShareAgreements: []types.TakerFeeShareAgreement{
							{
								Denom:       apptesting.DefaultTransmuterDenomA,
								SkimPercent: osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5")),
								SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
							},
							{
								Denom:       apptesting.DefaultTransmuterDenomB,
								SkimPercent: osmomath.MustNewDecFromStr("0.02").Mul(osmomath.MustNewDecFromStr("0.5")),
								SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
							},
						},
					},
					denomB: {
						ContractAddress: cwPool.GetAddress().String(),
						TakerFeeShareAgreements: []types.TakerFeeShareAgreement{
							{
								Denom:       "testA",
								SkimPercent: osmomath.MustNewDecFromStr("0.03").Mul(osmomath.MustNewDecFromStr("0.5")),
								SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
							},
						},
					},
				}
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			expectedTakerFeeShareAgreementsMap := tc.setupFunc()

			shareStatesMap, err := s.App.PoolManagerKeeper.GetAllRegisteredAlloyedPoolsMap(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(expectedTakerFeeShareAgreementsMap, shareStatesMap)
		})
	}
}

func (s *KeeperTestSuite) TestSetAllRegisteredAlloyedPoolsCached() {
	tests := map[string]struct {
		setupFunc func() map[string]types.AlloyContractTakerFeeShareState
	}{
		"no registered pools": {
			setupFunc: func() map[string]types.AlloyContractTakerFeeShareState {
				return map[string]types.AlloyContractTakerFeeShareState{}
			},
		},
		"single registered pool, no taker fee share agreements": {
			setupFunc: func() map[string]types.AlloyContractTakerFeeShareState {
				poolInfos := s.PrepareAllSupportedPools()
				pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfos.AlloyedPoolID)
				denom := fmt.Sprintf("factory/%s/alloyed/testdenom", pool.GetAddress().String())
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				return map[string]types.AlloyContractTakerFeeShareState{
					denom: {
						ContractAddress:         pool.GetAddress().String(),
						TakerFeeShareAgreements: []types.TakerFeeShareAgreement(nil),
					},
				}
			},
		},
		"multiple registered pools, no taker fee share agreements": {
			setupFunc: func() map[string]types.AlloyContractTakerFeeShareState {
				poolInfos := s.PrepareAllSupportedPools()
				pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3(s.TestAccs[0], []string{"testA", "testB"}, "osmosis", "x/cosmwasmpool/bytecode")
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				denomA := fmt.Sprintf("factory/%s/alloyed/testdenom", pool.GetAddress().String())
				denomB := fmt.Sprintf("factory/%s/alloyed/testdenom", cwPool.GetAddress().String())
				return map[string]types.AlloyContractTakerFeeShareState{
					denomA: {
						ContractAddress:         pool.GetAddress().String(),
						TakerFeeShareAgreements: []types.TakerFeeShareAgreement(nil),
					},
					denomB: {
						ContractAddress:         cwPool.GetAddress().String(),
						TakerFeeShareAgreements: []types.TakerFeeShareAgreement(nil),
					},
				}
			},
		},
		"single registered pool, with one taker fee share agreement": {
			setupFunc: func() map[string]types.AlloyContractTakerFeeShareState {
				poolInfos := s.PrepareAllSupportedPools()
				pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				denom := fmt.Sprintf("factory/%s/alloyed/testdenom", pool.GetAddress().String())
				takerFeeShareAgreement := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				return map[string]types.AlloyContractTakerFeeShareState{
					denom: {
						ContractAddress: pool.GetAddress().String(),
						TakerFeeShareAgreements: []types.TakerFeeShareAgreement{{
							Denom:       apptesting.DefaultTransmuterDenomA,
							SkimPercent: osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5")),
							SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
						}},
					},
				}
			},
		},
		"multiple registered pools, with multiple taker fee share agreements": {
			setupFunc: func() map[string]types.AlloyContractTakerFeeShareState {
				poolInfos := s.PrepareAllSupportedPools()
				pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				denomA := fmt.Sprintf("factory/%s/alloyed/testdenom", pool.GetAddress().String())
				takerFeeShareAgreementA := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomA,
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreementA)
				takerFeeShareAgreementB := types.TakerFeeShareAgreement{
					Denom:       apptesting.DefaultTransmuterDenomB,
					SkimPercent: osmomath.MustNewDecFromStr("0.02"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreementB)
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3(s.TestAccs[0], []string{"testA", "testB"}, "osmosis", "x/cosmwasmpool/bytecode")
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				denomB := fmt.Sprintf("factory/%s/alloyed/testdenom", cwPool.GetAddress().String())
				takerFeeShareAgreementTestA := types.TakerFeeShareAgreement{
					Denom:       "testA",
					SkimPercent: osmomath.MustNewDecFromStr("0.03"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreementTestA)
				return map[string]types.AlloyContractTakerFeeShareState{
					denomA: {
						ContractAddress: pool.GetAddress().String(),
						TakerFeeShareAgreements: []types.TakerFeeShareAgreement{
							{
								Denom:       apptesting.DefaultTransmuterDenomA,
								SkimPercent: osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5")),
								SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
							},
							{
								Denom:       apptesting.DefaultTransmuterDenomB,
								SkimPercent: osmomath.MustNewDecFromStr("0.02").Mul(osmomath.MustNewDecFromStr("0.5")),
								SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
							},
						},
					},
					denomB: {
						ContractAddress: cwPool.GetAddress().String(),
						TakerFeeShareAgreements: []types.TakerFeeShareAgreement{
							{
								Denom:       "testA",
								SkimPercent: osmomath.MustNewDecFromStr("0.03").Mul(osmomath.MustNewDecFromStr("0.5")),
								SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
							},
						},
					},
				}
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			expectedTakerFeeShareAgreementsMap := tc.setupFunc()

			// Call the function to test
			err := s.App.PoolManagerKeeper.SetAllRegisteredAlloyedPoolsCached(s.Ctx)
			s.Require().NoError(err)

			// Check that the cache was correctly set
			_, cachedRegisteredAlloyPoolToState, _ := s.App.PoolManagerKeeper.GetCachedMaps()
			s.Require().Equal(expectedTakerFeeShareAgreementsMap, cachedRegisteredAlloyPoolToState)
		})
	}
}

func (s *KeeperTestSuite) TestGetAllRegisteredAlloyedPoolsIdMap() {
	tests := map[string]struct {
		setupFunc func() map[uint64]bool
	}{
		"single registered pool": {
			setupFunc: func() map[uint64]bool {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				return map[uint64]bool{
					poolInfos.AlloyedPoolID: true,
				}
			},
		},
		"multiple registered pools": {
			setupFunc: func() map[uint64]bool {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3(s.TestAccs[0], []string{"testA", "testB"}, "osmosis", "x/cosmwasmpool/bytecode")
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				return map[uint64]bool{
					poolInfos.AlloyedPoolID: true,
					cwPool.GetId():          true,
				}
			},
		},
		"no registered pools": {
			setupFunc: func() map[uint64]bool {
				return map[uint64]bool{}
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			expectedRegisteredAlloyedPoolsIdMap := tc.setupFunc()

			registeredAlloyedPoolsIdMap, err := s.App.PoolManagerKeeper.GetAllRegisteredAlloyedPoolsIdMap(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(expectedRegisteredAlloyedPoolsIdMap, registeredAlloyedPoolsIdMap)
		})
	}
}

func (s *KeeperTestSuite) TestSetAllRegisteredAlloyedPoolsIdCached() {
	tests := map[string]struct {
		setupFunc func() map[uint64]bool
	}{
		"no registered pools": {
			setupFunc: func() map[uint64]bool {
				return map[uint64]bool{}
			},
		},
		"single registered pool": {
			setupFunc: func() map[uint64]bool {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				return map[uint64]bool{
					poolInfos.AlloyedPoolID: true,
				}
			},
		},
		"multiple registered pools": {
			setupFunc: func() map[uint64]bool {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3(s.TestAccs[0], []string{"testA", "testB"}, "osmosis", "x/cosmwasmpool/bytecode")
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				return map[uint64]bool{
					poolInfos.AlloyedPoolID: true,
					cwPool.GetId():          true,
				}
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			expectedRegisteredAlloyedPoolsIdMap := tc.setupFunc()

			// Call the function to test
			err := s.App.PoolManagerKeeper.SetAllRegisteredAlloyedPoolsIdCached(s.Ctx)
			s.Require().NoError(err)

			// Check that the cache was correctly set
			_, _, cachedRegisteredAlloyedPoolId := s.App.PoolManagerKeeper.GetCachedMaps()
			s.Require().Equal(expectedRegisteredAlloyedPoolsIdMap, cachedRegisteredAlloyedPoolId)
		})
	}
}

func (s *KeeperTestSuite) TestQueryAndCheckAlloyedDenom() {
	tests := map[string]struct {
		setupFunc     func() sdk.AccAddress
		expectedError error
	}{
		"valid alloyed denom": {
			setupFunc: func() sdk.AccAddress {
				poolInfos := s.PrepareAllSupportedPools()
				poolId := poolInfos.AlloyedPoolID
				pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolId)
				s.Require().NoError(err)
				return pool.GetAddress()
			},
		},
		"invalid alloyed denom": {
			setupFunc: func() sdk.AccAddress {
				poolInfos := s.PrepareAllSupportedPools()
				poolId := poolInfos.CosmWasmPoolID
				pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolId)
				s.Require().NoError(err)
				return pool.GetAddress()
			},
			expectedError: fmt.Errorf("third part of alloyedDenom should be 'alloyed'"),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			contractAddr := tc.setupFunc()

			_, err := s.App.PoolManagerKeeper.QueryAndCheckAlloyedDenom(s.Ctx, contractAddr)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError.Error(), err.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSnapshotTakerFeeShareAlloyComposition() {
	tests := map[string]struct {
		setupFunc     func() cosmwasmpooltypes.CosmWasmExtension
		expectedError error
	}{
		"alloyed pool exists, composed of no taker fee share denoms": {
			setupFunc: func() cosmwasmpooltypes.CosmWasmExtension {
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3(s.TestAccs[0], []string{"testA", "testB", "testC"}, "osmosis", "x/cosmwasmpool/bytecode")
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				return cwPool
			},
			expectedError: nil,
		},
		"alloyed pool exists, composed of one taker fee share denom": {
			setupFunc: func() cosmwasmpooltypes.CosmWasmExtension {
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3(s.TestAccs[0], []string{"testA", "testB", "testC"}, "osmosis", "x/cosmwasmpool/bytecode")
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
					Denom:       "testA",
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				})
				return cwPool
			},
			expectedError: nil,
		},
		"alloyed pool exists, composed of two taker fee share denoms, differing ratios": {
			setupFunc: func() cosmwasmpooltypes.CosmWasmExtension {
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3CustomRatio(s.TestAccs[0], []string{"testA", "testB", "testC"}, []uint16{1, 3, 6}, "osmosis", "x/cosmwasmpool/bytecode")
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
					Denom:       "testA",
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				})
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
					Denom:       "testB",
					SkimPercent: osmomath.MustNewDecFromStr("0.02"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				})
				return cwPool
			},
			expectedError: nil,
		},
		"alloyed pool exists, composed of two taker fee share denoms, differing ratios, first asset has no liquidity": {
			setupFunc: func() cosmwasmpooltypes.CosmWasmExtension {
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3CustomRatio(s.TestAccs[0], []string{"testA", "testB", "testC"}, []uint16{0, 3, 6}, "osmosis", "x/cosmwasmpool/bytecode")
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
					Denom:       "testA",
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				})
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
					Denom:       "testB",
					SkimPercent: osmomath.MustNewDecFromStr("0.02"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				})
				return cwPool
			},
			expectedError: nil,
		},
		"error: alloyed pool has no liquidity": {
			setupFunc: func() cosmwasmpooltypes.CosmWasmExtension {
				return s.PrepareCustomTransmuterPoolCustomProjectV3CustomRatio(s.TestAccs[0], []string{"testA", "testB", "testC"}, nil, "osmosis", "x/cosmwasmpool/bytecode")
			},
			expectedError: fmt.Errorf("totalAlloyedLiquidity is zero"),
		},
		// TODO: Diff scaling factors for assets
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			cwPool := tc.setupFunc()

			expectedComposition, err := s.App.PoolManagerKeeper.SnapshotTakerFeeShareAlloyComposition(s.Ctx, cwPool.GetAddress())
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError.Error(), err.Error())
			} else {
				s.Require().NoError(err)
				_, alloyedComp, err := s.App.PoolManagerKeeper.GetRegisteredAlloyedPoolFromPoolId(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				s.Require().Equal(expectedComposition, alloyedComp.TakerFeeShareAgreements)
			}
		})
	}
}

func (s *KeeperTestSuite) TestRecalculateAndSetTakerFeeShareAlloyComposition() {
	tests := map[string]struct {
		setupFunc                  func() uint64
		expectedUpdatedSkimPercent []osmomath.Dec
		expectedError              error
	}{
		"1:1 to 2:1": {
			setupFunc: func() uint64 {
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3CustomRatio(s.TestAccs[0], []string{"testA", "testB"}, []uint16{1, 1}, "osmosis", "x/cosmwasmpool/bytecode")
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
					Denom:       "testA",
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				})
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
					Denom:       "testB",
					SkimPercent: osmomath.MustNewDecFromStr("0.02"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				})
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)

				testACoins := sdk.NewCoins(sdk.NewCoin("testA", osmomath.NewInt(1000000000)))
				s.FundAcc(s.TestAccs[0], testACoins)
				s.JoinTransmuterPool(s.TestAccs[0], cwPool.GetId(), testACoins)
				return cwPool.GetId()
			},
			expectedUpdatedSkimPercent: []osmomath.Dec{
				osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.66666666666666666")),
				osmomath.MustNewDecFromStr("0.02").Mul(osmomath.MustNewDecFromStr("0.33333333333333333")),
			},
			expectedError: nil,
		},
		"1:1:1 to 3:2:1": {
			setupFunc: func() uint64 {
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3CustomRatio(s.TestAccs[0], []string{"testA", "testB", "testC"}, []uint16{1, 1, 1}, "osmosis", "x/cosmwasmpool/bytecode")
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
					Denom:       "testA",
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				})
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
					Denom:       "testB",
					SkimPercent: osmomath.MustNewDecFromStr("0.02"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				})
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
					Denom:       "testC",
					SkimPercent: osmomath.MustNewDecFromStr("0.03"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc3",
				})
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)

				// Change the ratio to 3:2:1 by adding more of testA and testB
				testACoins := sdk.NewCoins(sdk.NewCoin("testA", osmomath.NewInt(2000000000)))
				testBCoins := sdk.NewCoins(sdk.NewCoin("testB", osmomath.NewInt(1000000000)))
				s.FundAcc(s.TestAccs[0], testACoins.Add(testBCoins...))
				s.JoinTransmuterPool(s.TestAccs[0], cwPool.GetId(), testACoins.Add(testBCoins...))

				return cwPool.GetId()
			},
			expectedUpdatedSkimPercent: []osmomath.Dec{
				osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5")),
				osmomath.MustNewDecFromStr("0.02").Mul(osmomath.MustNewDecFromStr("0.33333333333333333")),
				osmomath.MustNewDecFromStr("0.03").Mul(osmomath.MustNewDecFromStr("0.16666666666666666")),
			},
			expectedError: nil,
		},
		"1:1:1 to 4:2:1": {
			setupFunc: func() uint64 {
				cwPool := s.PrepareCustomTransmuterPoolCustomProjectV3CustomRatio(s.TestAccs[0], []string{"testA", "testB", "testC"}, []uint16{1, 1, 1}, "osmosis", "x/cosmwasmpool/bytecode")
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
					Denom:       "testA",
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				})
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
					Denom:       "testB",
					SkimPercent: osmomath.MustNewDecFromStr("0.02"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
				})
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, types.TakerFeeShareAgreement{
					Denom:       "testC",
					SkimPercent: osmomath.MustNewDecFromStr("0.03"),
					SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc3",
				})
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)

				// Change the ratio to 4:2:1 by adding more of testA and testB
				testACoins := sdk.NewCoins(sdk.NewCoin("testA", osmomath.NewInt(3000000000)))
				testBCoins := sdk.NewCoins(sdk.NewCoin("testB", osmomath.NewInt(1000000000)))
				s.FundAcc(s.TestAccs[0], testACoins.Add(testBCoins...))
				s.JoinTransmuterPool(s.TestAccs[0], cwPool.GetId(), testACoins.Add(testBCoins...))

				return cwPool.GetId()
			},
			expectedUpdatedSkimPercent: []osmomath.Dec{
				osmomath.MustNewDecFromStr("0.01").Mul(osmomath.MustNewDecFromStr("0.5714285714285714")),
				osmomath.MustNewDecFromStr("0.02").Mul(osmomath.MustNewDecFromStr("0.2857142857142857")),
				osmomath.MustNewDecFromStr("0.03").Mul(osmomath.MustNewDecFromStr("0.14285714285714285")),
			},
			expectedError: nil,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			poolId := tc.setupFunc()

			err := s.App.PoolManagerKeeper.RecalculateAndSetTakerFeeShareAlloyComposition(s.Ctx, poolId)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError.Error(), err.Error())
			} else {
				_, shareStates, err := s.App.PoolManagerKeeper.GetRegisteredAlloyedPoolFromPoolId(s.Ctx, poolId)
				for i, shareState := range shareStates.TakerFeeShareAgreements {
					s.Require().Equal(tc.expectedUpdatedSkimPercent[i], shareState.SkimPercent)
				}
				s.Require().NoError(err)
			}
		})
	}
}

func reverseSlice(input []types.TakerFeeShareAgreement) []types.TakerFeeShareAgreement {
	if len(input) == 0 {
		return input
	}
	return append(reverseSlice(input[1:]), input[0])
}
