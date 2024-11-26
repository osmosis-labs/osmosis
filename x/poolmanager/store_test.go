package poolmanager_test

import (
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	v3 "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/cosmwasm/msg/v3"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

var (
	denomA           = apptesting.DefaultTransmuterDenomA
	denomB           = apptesting.DefaultTransmuterDenomB
	denomC           = apptesting.DefaultTransmuterDenomC
	OSMO             = "uosmo"
	ATOM             = "atom"
	secondaryDenomA  = "testA"
	secondaryDenomB  = "testB"
	nonExistentDenom = "nonExistentDenom"

	oneHundred   = osmomath.NewInt(100)
	twoHundred   = osmomath.NewInt(200)
	threeHundred = osmomath.NewInt(300)

	defaultTakerFeeShareAgreements = []types.TakerFeeShareAgreement{
		{
			Denom:       denomA,
			SkimPercent: osmomath.MustNewDecFromStr("0.01"),
			SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
		},
		{
			Denom:       denomB,
			SkimPercent: osmomath.MustNewDecFromStr("0.02"),
			SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
		},
		{
			Denom:       denomC,
			SkimPercent: osmomath.MustNewDecFromStr("0.03"),
			SkimAddress: "osmo1jermpr9yust7cyhfjme3cr08kt6n8jv6p35l39",
		},
	}

	secondaryTakerFeeShareAgreements = []types.TakerFeeShareAgreement{
		{
			Denom:       secondaryDenomA,
			SkimPercent: osmomath.MustNewDecFromStr("0.03"),
			SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
		},
		{
			Denom:       secondaryDenomB,
			SkimPercent: osmomath.MustNewDecFromStr("0.04"),
			SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc3",
		},
	}
)

func (s *KeeperTestSuite) TestGetAllTakerFeeShareAgreementsMap() {
	tests := map[string]struct {
		setupFunc                         func()
		expectedTakerFeeShareAgreementMap func() map[string]types.TakerFeeShareAgreement
	}{
		"single taker fee share agreement": {
			setupFunc: func() {
				numTakerFeeShareAgreements := 1
				for i, takerFeeShareAgreement := range defaultTakerFeeShareAgreements {
					if i >= numTakerFeeShareAgreements {
						break
					}
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			expectedTakerFeeShareAgreementMap: func() map[string]types.TakerFeeShareAgreement {
				numTakerFeeShareAgreements := 1
				expectedTakerFeeShareAgreementMap := make(map[string]types.TakerFeeShareAgreement)
				for i, takerFeeShareAgreement := range defaultTakerFeeShareAgreements {
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
				for _, takerFeeShareAgreement := range defaultTakerFeeShareAgreements {
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			expectedTakerFeeShareAgreementMap: func() map[string]types.TakerFeeShareAgreement {
				expectedTakerFeeShareAgreementMap := make(map[string]types.TakerFeeShareAgreement)
				for _, takerFeeShareAgreement := range defaultTakerFeeShareAgreements {
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
				for i, takerFeeShareAgreement := range defaultTakerFeeShareAgreements {
					if i >= numTakerFeeShareAgreements {
						break
					}
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			expectedTakerFeeShareAgreements: func() []types.TakerFeeShareAgreement {
				numTakerFeeShareAgreements := 1
				var expectedTakerFeeShareAgreements []types.TakerFeeShareAgreement
				for i, takerFeeShareAgreement := range defaultTakerFeeShareAgreements {
					if i >= numTakerFeeShareAgreements {
						break
					}
					expectedTakerFeeShareAgreements = append(expectedTakerFeeShareAgreements, takerFeeShareAgreement)
				}
				return expectedTakerFeeShareAgreements
			},
		},
		"multiple taker fee share agreements": {
			setupFunc: func() {
				for _, takerFeeShareAgreement := range defaultTakerFeeShareAgreements {
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			expectedTakerFeeShareAgreements: func() []types.TakerFeeShareAgreement {
				return defaultTakerFeeShareAgreements
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

			takerFeeShareAgreements, err := s.App.PoolManagerKeeper.GetAllTakerFeesShareAgreements(s.Ctx)
			s.Require().NoError(err)
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
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:1])
			},
			expectedTakerFeeShareAgreements: func() map[string]types.TakerFeeShareAgreement {
				return createExpectedTakerFeeShareAgreementsMap(defaultTakerFeeShareAgreements[:1])
			},
			expectErr: false,
		},
		"multiple taker fee share agreements": {
			setupFunc: func() {
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements)
			},
			expectedTakerFeeShareAgreements: func() map[string]types.TakerFeeShareAgreement {
				return createExpectedTakerFeeShareAgreementsMap(defaultTakerFeeShareAgreements)
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
				cachedTakerFeeShareAgreementMap, _ := s.App.PoolManagerKeeper.GetCacheTrackers()
				s.Require().Equal(expectedTakerFeeShareAgreements, cachedTakerFeeShareAgreementMap, "cachedTakerFeeShareAgreementMap = %v, want %v", cachedTakerFeeShareAgreementMap, expectedTakerFeeShareAgreements)
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
				for i, takerFeeShareAgreement := range defaultTakerFeeShareAgreements {
					if i >= numTakerFeeShareAgreements {
						break
					}
					err := s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
					s.Require().NoError(err)
				}
			},
			denomToRequest:                 denomA,
			expectedTakerFeeShareAgreement: defaultTakerFeeShareAgreements[0],
			expectedFound:                  true,
		},
		"set three taker fee share agreements, get one of the three taker fee share agreements": {
			setupFunc: func() {
				for _, takerFeeShareAgreement := range defaultTakerFeeShareAgreements {
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			denomToRequest:                 denomB,
			expectedTakerFeeShareAgreement: defaultTakerFeeShareAgreements[1],
			expectedFound:                  true,
		},
		"set three taker fee share agreements, attempt to get taker fee share agreement for denom that does not exist": {
			setupFunc: func() {
				for _, takerFeeShareAgreement := range defaultTakerFeeShareAgreements {
					s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, takerFeeShareAgreement)
				}
			},
			denomToRequest: nonExistentDenom,
			expectedFound:  false,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			tc.setupFunc()

			takerFeeShareAgreements, found := s.App.PoolManagerKeeper.GetTakerFeeShareAgreementFromDenom(tc.denomToRequest)
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
		setupFunc            func()
		takerFeeShareDenom   string
		takerFeeChargedDenom string
		expectedValue        osmomath.Int
		expectedError        error
	}{
		"tier denom accrued denom value, retrieve denom value": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, denomB, oneHundred)
				s.Require().NoError(err)
			},
			takerFeeShareDenom:   denomA,
			takerFeeChargedDenom: denomB,
			expectedValue:        oneHundred,
		},
		"tier denom did not accrue value, nothing to retrieve, so not found": {
			setupFunc:            func() {},
			takerFeeShareDenom:   denomA,
			takerFeeChargedDenom: denomB,
			expectedError:        types.NoAccruedValueError{TakerFeeShareDenom: denomA, TakerFeeChargedDenom: denomB},
		},
		"tier denom accrued denom value, retrieve different denom value, so not found": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, denomB, oneHundred)
				s.Require().NoError(err)
			},
			takerFeeShareDenom:   denomA,
			takerFeeChargedDenom: nonExistentDenom,
			expectedError:        types.NoAccruedValueError{TakerFeeShareDenom: denomA, TakerFeeChargedDenom: nonExistentDenom},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()

			value, err := s.App.PoolManagerKeeper.GetTakerFeeShareDenomsToAccruedValue(s.Ctx, tc.takerFeeShareDenom, tc.takerFeeChargedDenom)
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
			tierDenom:     denomA,
			takerFeeDenom: denomB,
			accruedValue:  oneHundred,
			expectErr:     false,
		},
		"set value lower than what it previously was, should override": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, denomB, twoHundred)
				s.Require().NoError(err)
			},
			tierDenom:     denomA,
			takerFeeDenom: denomB,
			accruedValue:  oneHundred,
			expectErr:     false,
		},
		"set value greater than what it previously was, should override": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, denomB, twoHundred)
				s.Require().NoError(err)
			},
			tierDenom:     denomA,
			takerFeeDenom: denomB,
			accruedValue:  threeHundred,
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
		setupFunc          func()
		takerFeeShareDenom string
		takerFeeDenom      string
		additiveValue      osmomath.Int
		expectedValue      osmomath.Int
		expectErr          bool
	}{
		"increase value that was previously non existent": {
			setupFunc:          func() {},
			takerFeeShareDenom: denomA,
			takerFeeDenom:      denomB,
			additiveValue:      oneHundred,
			expectedValue:      oneHundred,
			expectErr:          false,
		},
		"increase value that was previously set": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, denomB, twoHundred)
				s.Require().NoError(err)
			},
			takerFeeShareDenom: denomA,
			takerFeeDenom:      denomB,
			additiveValue:      oneHundred,
			expectedValue:      threeHundred,
			expectErr:          false,
		},
		"increase value with zero additive value": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, denomB, twoHundred)
				s.Require().NoError(err)
			},
			takerFeeShareDenom: denomA,
			takerFeeDenom:      denomB,
			additiveValue:      osmomath.ZeroInt(),
			expectedValue:      twoHundred,
			expectErr:          false,
		},
		"increase value with very large additive value": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, denomB, twoHundred)
				s.Require().NoError(err)
			},
			takerFeeShareDenom: denomA,
			takerFeeDenom:      denomB,
			additiveValue:      osmomath.NewIntFromUint64(math.MaxUint64),
			expectedValue:      osmomath.NewIntFromUint64(math.MaxUint64).Add(twoHundred),
			expectErr:          false,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()

			err := s.App.PoolManagerKeeper.IncreaseTakerFeeShareDenomsToAccruedValue(s.Ctx, tc.takerFeeShareDenom, tc.takerFeeDenom, tc.additiveValue)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				value, err := s.App.PoolManagerKeeper.GetTakerFeeShareDenomsToAccruedValue(s.Ctx, tc.takerFeeShareDenom, tc.takerFeeDenom)
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
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, OSMO, oneHundred)
				s.Require().NoError(err)
			},
			expectedAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            denomA,
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin(OSMO, oneHundred)),
				},
			},
		},
		"multiple accumulators": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, OSMO, oneHundred)
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, ATOM, twoHundred)
				s.Require().NoError(err)
			},
			expectedAccumulators: []types.TakerFeeSkimAccumulator{
				{
					Denom:            denomA,
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin(OSMO, oneHundred), sdk.NewCoin(ATOM, twoHundred)),
				},
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()

			accumulators, err := s.App.PoolManagerKeeper.GetAllTakerFeeShareAccumulators(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedAccumulators, accumulators, "GetAllTakerFeeShareAccumulators() = %v, want %v", accumulators, tc.expectedAccumulators)
		})
	}
}

func (s *KeeperTestSuite) TestDeleteAllTakerFeeShareAccumulatorsForTakerFeeShareDenom() {
	tests := map[string]struct {
		setupFunc                      func()
		takerFeeShareDenom             string
		expectedNumOfAccumsAfterDelete uint64
	}{
		"delete non-existent accumulators": {
			setupFunc:                      func() {},
			takerFeeShareDenom:             denomA,
			expectedNumOfAccumsAfterDelete: 0,
		},
		"delete existing accumulators": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, OSMO, oneHundred)
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, ATOM, twoHundred)
				s.Require().NoError(err)
			},
			takerFeeShareDenom:             denomA,
			expectedNumOfAccumsAfterDelete: 0,
		},
		"delete existing accumulators for non existent tier denom": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, OSMO, oneHundred)
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, ATOM, twoHundred)
				s.Require().NoError(err)
			},
			takerFeeShareDenom:             "nonExistentTierDenom",
			expectedNumOfAccumsAfterDelete: 1,
		},
		"delete existing accumulators for tier denom with accumulators": {
			setupFunc: func() {
				err := s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, OSMO, oneHundred)
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, ATOM, twoHundred)
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomC, OSMO, oneHundred)
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomC, ATOM, twoHundred)
				s.Require().NoError(err)
			},
			takerFeeShareDenom:             denomC,
			expectedNumOfAccumsAfterDelete: 1,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()

			s.App.PoolManagerKeeper.DeleteAllTakerFeeShareAccumulatorsForTakerFeeShareDenom(s.Ctx, tc.takerFeeShareDenom)

			accumulators, err := s.App.PoolManagerKeeper.GetAllTakerFeeShareAccumulators(s.Ctx)
			s.Require().NoError(err)
			s.Require().Len(accumulators, int(tc.expectedNumOfAccumsAfterDelete))
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
			expectedError: types.NotCosmWasmPoolError{PoolId: 2},
		},
		"set concentrated pool": {
			poolType:      ConcentratedPool,
			preSetFunc:    func(ctx sdk.Context) {},
			postSetFunc:   func(ctx sdk.Context) {},
			expectedError: types.NotCosmWasmPoolError{PoolId: 1},
		},
		"set cw pool": {
			poolType:      CWPool,
			preSetFunc:    func(ctx sdk.Context) {},
			postSetFunc:   func(ctx sdk.Context) {},
			expectedError: types.InvalidAlloyedDenomPartError{PartIndex: 2, Expected: "alloyed", Actual: "transmuter"},
		},
		"set alloyed pool, with one taker fee share agreement set before alloyed pool is registered and one set after": {
			poolType: AlloyedPool,
			preSetFunc: func(ctx sdk.Context) {
				setTakerFeeShareAgreements(ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:1])
			},
			postSetFunc: func(ctx sdk.Context) {
				setTakerFeeShareAgreements(ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[1:2])
			},
			expectedTakerFeeShareAgreements: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{osmomath.MustNewDecFromStr("0.5"), osmomath.MustNewDecFromStr("0.5")}),
		},
		"set alloyed pool, with both taker fee share agreements set before alloyed pool is registered": {
			poolType: AlloyedPool,
			preSetFunc: func(ctx sdk.Context) {
				setTakerFeeShareAgreements(ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
			},
			postSetFunc:                     func(ctx sdk.Context) {},
			expectedTakerFeeShareAgreements: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{osmomath.MustNewDecFromStr("0.5"), osmomath.MustNewDecFromStr("0.5")}),
		},
		"set alloyed pool, with both taker fee share agreements set after alloyed pool is registered": {
			poolType:   AlloyedPool,
			preSetFunc: func(ctx sdk.Context) {},
			postSetFunc: func(ctx sdk.Context) {
				setTakerFeeShareAgreements(ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
			},
			expectedTakerFeeShareAgreements: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{osmomath.MustNewDecFromStr("0.5"), osmomath.MustNewDecFromStr("0.5")}),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			poolInfos := s.PrepareAllSupportedPools()
			switch tc.poolType {
			case GammPool:
				tc.poolId = poolInfos.BalancerPoolID
			case ConcentratedPool:
				tc.poolId = poolInfos.ConcentratedPoolID
			case CWPool:
				tc.poolId = poolInfos.CosmWasmPoolID
			case AlloyedPool:
				tc.poolId = poolInfos.AlloyedPoolID
			default:
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
				shareState, err := s.App.PoolManagerKeeper.GetRegisteredAlloyedPoolFromPoolId(s.Ctx, tc.poolId)
				s.Require().NoError(err)
				alloyedDenom, err := s.App.PoolManagerKeeper.GetAlloyedDenomFromPoolId(s.Ctx, tc.poolId)
				s.Require().NoError(err)

				alloyedPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, tc.poolId)
				s.Require().NoError(err)
				expectedAlloyedDenom := createAlloyedDenom(alloyedPool.GetAddress().String())
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
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:1])
			},
			getDenomFromPool: true,
			expectFound:      true,
			expectedTakerFeeShareAgreements: modifySkimPercent(defaultTakerFeeShareAgreements[:1], []osmomath.Dec{
				osmomath.MustNewDecFromStr("0.5"),
			}),
		},
		"setup alloyed pool, get the alloyed pool denom, with multiple taker fee share agreements": {
			setupFunc: func() {
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
			},
			getDenomFromPool: true,
			expectFound:      true,
			expectedTakerFeeShareAgreements: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{
				osmomath.MustNewDecFromStr("0.5"),
				osmomath.MustNewDecFromStr("0.5"),
			}),
		},
		"setup alloyed pool, get the alloyed pool denom, with unrelated taker fee share agreement": {
			setupFunc: func() {
				unrelatedAgreement := types.TakerFeeShareAgreement{
					Denom:       "unrelatedDenom",
					SkimPercent: osmomath.MustNewDecFromStr("0.01"),
					SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
				}
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, unrelatedAgreement)
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
				tc.denom = createAlloyedDenom(alloyedPool.GetAddress().String())
			}

			s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)

			shareState, found := s.App.PoolManagerKeeper.GetRegisteredAlloyedPoolFromDenom(tc.denom)
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
			expectedError: types.NoRegisteredAlloyedPoolError{PoolId: 1},
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
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:1])
			},
			poolId: 5,
			expectedTakerFeeShareAgreements: modifySkimPercent(defaultTakerFeeShareAgreements[:1], []osmomath.Dec{
				osmomath.MustNewDecFromStr("0.5"),
			}),
		},
		"get existing non-alloyed pool": {
			setupFunc: func() {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
			},
			poolId:        1,
			expectedError: types.NoRegisteredAlloyedPoolError{PoolId: 1},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			tc.setupFunc()

			shareState, err := s.App.PoolManagerKeeper.GetRegisteredAlloyedPoolFromPoolId(s.Ctx, tc.poolId)
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
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{secondaryDenomA, secondaryDenomB}, []uint16{1, 1})
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
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:1])
			},
			expectedTakerFeeShareAgreements: [][]types.TakerFeeShareAgreement{
				modifySkimPercent(defaultTakerFeeShareAgreements[:1], []osmomath.Dec{
					osmomath.MustNewDecFromStr("0.5"),
				}),
			},
		},
		"multiple registered pools, with multiple taker fee share agreements": {
			setupFunc: func() {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{secondaryDenomA, secondaryDenomB}, []uint16{1, 1})
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, secondaryTakerFeeShareAgreements[:1])
			},
			expectedTakerFeeShareAgreements: [][]types.TakerFeeShareAgreement{
				modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{
					osmomath.MustNewDecFromStr("0.5"),
					osmomath.MustNewDecFromStr("0.5"),
				}),
				modifySkimPercent(secondaryTakerFeeShareAgreements[:1], []osmomath.Dec{
					osmomath.MustNewDecFromStr("0.5"),
				}),
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
				denom := createAlloyedDenom(pool.GetAddress().String())
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
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{secondaryDenomA, secondaryDenomB}, []uint16{1, 1})
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				denomA := createAlloyedDenom(pool.GetAddress().String())
				denomB := createAlloyedDenom(cwPool.GetAddress().String())
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
				denom := createAlloyedDenom(pool.GetAddress().String())
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:1])
				return map[string]types.AlloyContractTakerFeeShareState{
					denom: {
						ContractAddress: pool.GetAddress().String(),
						TakerFeeShareAgreements: modifySkimPercent(defaultTakerFeeShareAgreements[:1], []osmomath.Dec{
							osmomath.MustNewDecFromStr("0.5"),
						}),
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
				denomA := createAlloyedDenom(pool.GetAddress().String())
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{secondaryDenomA, secondaryDenomB}, []uint16{1, 1})
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				denomB := createAlloyedDenom(cwPool.GetAddress().String())
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, secondaryTakerFeeShareAgreements[:1])
				return map[string]types.AlloyContractTakerFeeShareState{
					denomA: {
						ContractAddress: pool.GetAddress().String(),
						TakerFeeShareAgreements: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{
							osmomath.MustNewDecFromStr("0.5"),
							osmomath.MustNewDecFromStr("0.5"),
						}),
					},
					denomB: {
						ContractAddress: cwPool.GetAddress().String(),
						TakerFeeShareAgreements: modifySkimPercent(secondaryTakerFeeShareAgreements[:1], []osmomath.Dec{
							osmomath.MustNewDecFromStr("0.5"),
						}),
					},
				}
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			expectedTakerFeeShareAgreementsMap := tc.setupFunc()

			shareStatesMap, err := s.App.PoolManagerKeeper.GetAllRegisteredAlloyedPoolsByDenomMap(s.Ctx)
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
				s.Require().NoError(err)
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				denom := createAlloyedDenom(pool.GetAddress().String())
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
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{secondaryDenomA, secondaryDenomB}, []uint16{1, 1})
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				denomA := createAlloyedDenom(pool.GetAddress().String())
				denomB := createAlloyedDenom(cwPool.GetAddress().String())
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
				denom := createAlloyedDenom(pool.GetAddress().String())
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:1])
				return map[string]types.AlloyContractTakerFeeShareState{
					denom: {
						ContractAddress: pool.GetAddress().String(),
						TakerFeeShareAgreements: modifySkimPercent(defaultTakerFeeShareAgreements[:1], []osmomath.Dec{
							osmomath.MustNewDecFromStr("0.5"),
						}),
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
				denomA := createAlloyedDenom(pool.GetAddress().String())
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{secondaryDenomA, secondaryDenomB}, []uint16{1, 1})
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				denomB := createAlloyedDenom(cwPool.GetAddress().String())
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, secondaryTakerFeeShareAgreements[:1])
				return map[string]types.AlloyContractTakerFeeShareState{
					denomA: {
						ContractAddress: pool.GetAddress().String(),
						TakerFeeShareAgreements: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{
							osmomath.MustNewDecFromStr("0.5"),
							osmomath.MustNewDecFromStr("0.5"),
						}),
					},
					denomB: {
						ContractAddress: cwPool.GetAddress().String(),
						TakerFeeShareAgreements: modifySkimPercent(secondaryTakerFeeShareAgreements[:1], []osmomath.Dec{
							osmomath.MustNewDecFromStr("0.5"),
						}),
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
			err := s.App.PoolManagerKeeper.SetAllRegisteredAlloyedPoolsByDenomCached(s.Ctx)
			s.Require().NoError(err)

			// Check that the cache was correctly set
			_, cachedRegisteredAlloyPoolByAlloyDenomMap := s.App.PoolManagerKeeper.GetCacheTrackers()
			s.Require().Equal(expectedTakerFeeShareAgreementsMap, cachedRegisteredAlloyPoolByAlloyDenomMap)
		})
	}
}

func (s *KeeperTestSuite) TestGetAllRegisteredAlloyedPoolsIdMap() {
	tests := map[string]struct {
		setupFunc func() []uint64
	}{
		"single registered pool": {
			setupFunc: func() []uint64 {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				return []uint64{poolInfos.AlloyedPoolID}
			},
		},
		"multiple registered pools": {
			setupFunc: func() []uint64 {
				poolInfos := s.PrepareAllSupportedPools()
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, poolInfos.AlloyedPoolID)
				s.Require().NoError(err)
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{secondaryDenomA, secondaryDenomB}, []uint16{1, 1})
				err = s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				return []uint64{poolInfos.AlloyedPoolID, cwPool.GetId()}
			},
		},
		"no registered pools": {
			setupFunc: func() []uint64 {
				return []uint64{}
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			expectedRegisteredAlloyedPoolsIdMap := tc.setupFunc()

			registeredAlloyedPoolIdsArray, err := s.App.PoolManagerKeeper.GetAllRegisteredAlloyedPoolsIdArray(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(expectedRegisteredAlloyedPoolsIdMap, registeredAlloyedPoolIdsArray)
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
			expectedError: types.InvalidAlloyedDenomPartError{PartIndex: 2, Expected: "alloyed", Actual: "transmuter"},
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
		setupFunc           func() cosmwasmpooltypes.CosmWasmExtension
		expectedComposition []types.TakerFeeShareAgreement
		expectedError       error
	}{
		"alloyed pool exists, composed of no taker fee share denoms": {
			setupFunc: func() cosmwasmpooltypes.CosmWasmExtension {
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{denomA, denomB, denomC}, []uint16{1, 1, 1})
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				return cwPool
			},
			expectedComposition: []types.TakerFeeShareAgreement(nil),
			expectedError:       nil,
		},
		"alloyed pool exists, composed of one taker fee share denom": {
			setupFunc: func() cosmwasmpooltypes.CosmWasmExtension {
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{denomA, denomB, denomC}, []uint16{1, 1, 1})
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:1])
				return cwPool
			},
			expectedComposition: modifySkimPercent(defaultTakerFeeShareAgreements[:1], []osmomath.Dec{
				osmomath.MustNewDecFromStr("0.3333333333333333"),
			}),
			expectedError: nil,
		},
		"alloyed pool exists, composed of two taker fee share denoms, differing ratios": {
			setupFunc: func() cosmwasmpooltypes.CosmWasmExtension {
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{denomA, denomB, denomC}, []uint16{1, 3, 6})
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
				return cwPool
			},
			expectedComposition: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{
				osmomath.MustNewDecFromStr("0.1"),
				osmomath.MustNewDecFromStr("0.3"),
			}),
			expectedError: nil,
		},
		"alloyed pool exists, composed of two taker fee share denoms, differing ratios, first asset has no liquidity": {
			setupFunc: func() cosmwasmpooltypes.CosmWasmExtension {
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{denomA, denomB, denomC}, []uint16{0, 3, 6})
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
				return cwPool
			},
			expectedComposition: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{
				osmomath.ZeroDec(),
				osmomath.MustNewDecFromStr("0.33333333333333333"),
			}),
			expectedError: nil,
		},
		"error: alloyed pool has no liquidity": {
			setupFunc: func() cosmwasmpooltypes.CosmWasmExtension {
				return s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{denomA, denomB, denomC}, nil)
			},
			expectedComposition: []types.TakerFeeShareAgreement{},
			expectedError:       types.ErrTotalAlloyedLiquidityIsZero,
		},
		"alloyed pool with normalization factors 1 and 1000000000000": {
			setupFunc: func() cosmwasmpooltypes.CosmWasmExtension {
				cwPool := s.PrepareCustomTransmuterPoolV3WithNormalization(s.TestAccs[0], []string{denomA, denomB}, []string{"1", "1000000000000"}, []uint16{1, 1})
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
				return cwPool
			},
			expectedComposition: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{
				osmomath.MustNewDecFromStr("100000000000000").Quo(osmomath.MustNewDecFromStr("100000000000100")),
				osmomath.MustNewDecFromStr("100").Quo(osmomath.MustNewDecFromStr("100000000000100")),
			}),
			expectedError: nil,
		},
		"alloyed pool with normalization factors 1, 1000000000000, and 1": {
			setupFunc: func() cosmwasmpooltypes.CosmWasmExtension {
				cwPool := s.PrepareCustomTransmuterPoolV3WithNormalization(s.TestAccs[0], []string{denomA, denomB, denomC}, []string{"1", "1000000000000", "1"}, []uint16{1, 1, 1})
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, append(defaultTakerFeeShareAgreements[:2], types.TakerFeeShareAgreement{
					Denom:       denomC,
					SkimPercent: osmomath.MustNewDecFromStr("0.03"),
					SkimAddress: "osmo1k5t7xrevz5fhvs5zg5jtpnht2mzv539008uc3",
				}))
				return cwPool
			},
			expectedComposition: modifySkimPercent(append(defaultTakerFeeShareAgreements[:2], types.TakerFeeShareAgreement{
				Denom:       denomC,
				SkimPercent: osmomath.MustNewDecFromStr("0.03"),
				SkimAddress: "osmo1k5t7xrevz5fhvs5zg5jtpnht2mzv539008uc3",
			}), []osmomath.Dec{
				osmomath.MustNewDecFromStr("100000000000000").Quo(osmomath.MustNewDecFromStr("200000000000100")),
				osmomath.MustNewDecFromStr("100").Quo(osmomath.MustNewDecFromStr("200000000000100")),
				osmomath.MustNewDecFromStr("100000000000000").Quo(osmomath.MustNewDecFromStr("200000000000100")),
			}),
			expectedError: nil,
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			cwPool := tc.setupFunc()
			actualComposition, err := s.App.PoolManagerKeeper.SnapshotTakerFeeShareAlloyComposition(s.Ctx, cwPool.GetAddress())
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError.Error(), err.Error())
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedComposition, actualComposition)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCreateNormalizationFactorsMap() {
	tests := map[string]struct {
		assetConfigs []v3.AssetConfig
		expected     map[string]osmomath.Dec
		expectedErr  error
	}{
		"successful creation": {
			assetConfigs: []v3.AssetConfig{
				{Denom: "denomA", NormalizationFactor: "1"},
				{Denom: "denomB", NormalizationFactor: "1000"},
			},
			expected: map[string]osmomath.Dec{
				"denomA": osmomath.OneDec(),
				"denomB": osmomath.MustNewDecFromStr("1000"),
			},
			expectedErr: nil,
		},
		"error in normalization factor": {
			assetConfigs: []v3.AssetConfig{
				{Denom: "denomA", NormalizationFactor: "invalid"},
			},
			expected:    nil,
			expectedErr: fmt.Errorf("failed to set decimal string with base 10: invalid000000000000000000"),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			actual, err := s.App.PoolManagerKeeper.CreateNormalizationFactorsMap(tc.assetConfigs)
			if tc.expectedErr != nil {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectedErr.Error())
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expected, actual)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCalculateTakerFeeShareAgreements() {
	tests := map[string]struct {
		totalPoolLiquidity   []sdk.Coin
		normalizationFactors map[string]osmomath.Dec
		setupFunc            func()
		expected             []types.TakerFeeShareAgreement
		expectedErr          error
	}{
		"successful calculation": {
			totalPoolLiquidity: []sdk.Coin{
				{Denom: denomA, Amount: oneHundred},
				{Denom: denomB, Amount: twoHundred},
			},
			normalizationFactors: map[string]osmomath.Dec{
				denomA: osmomath.OneDec(),
				denomB: osmomath.MustNewDecFromStr("10"),
			},
			setupFunc: func() {
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
			},
			expected: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{
				osmomath.MustNewDecFromStr("1000").Quo(osmomath.MustNewDecFromStr("1200")),
				osmomath.MustNewDecFromStr("200").Quo(osmomath.MustNewDecFromStr("1200")),
			}),
			expectedErr: nil,
		},
		"error: total alloyed liquidity is zero": {
			totalPoolLiquidity: []sdk.Coin{
				{Denom: denomA, Amount: osmomath.ZeroInt()},
				{Denom: denomB, Amount: osmomath.ZeroInt()},
			},
			normalizationFactors: map[string]osmomath.Dec{
				denomA: osmomath.OneDec(),
				denomB: osmomath.OneDec(),
			},
			setupFunc: func() {
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
			},
			expected:    nil,
			expectedErr: types.ErrTotalAlloyedLiquidityIsZero,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			tc.setupFunc()
			actual, err := s.App.PoolManagerKeeper.CalculateTakerFeeShareAgreements(s.Ctx, tc.totalPoolLiquidity, tc.normalizationFactors)
			if tc.expectedErr != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedErr.Error(), err.Error())
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expected, actual)
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
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{denomA, denomB}, []uint16{1, 1})
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)

				testACoins := sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(1000000000)))
				s.FundAcc(s.TestAccs[0], testACoins)
				s.JoinTransmuterPool(s.TestAccs[0], cwPool.GetId(), testACoins)
				return cwPool.GetId()
			},
			expectedUpdatedSkimPercent: []osmomath.Dec{
				defaultTakerFeeShareAgreements[0].SkimPercent.Mul(osmomath.MustNewDecFromStr("0.66666666666666666")),
				defaultTakerFeeShareAgreements[1].SkimPercent.Mul(osmomath.MustNewDecFromStr("0.33333333333333333")),
			},
			expectedError: nil,
		},
		"1:1:1 to 3:2:1": {
			setupFunc: func() uint64 {
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{denomA, denomB, denomC}, []uint16{1, 1, 1})
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:3])
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)

				// Change the ratio to 3:2:1 by adding more of denomA and denomB
				testACoins := sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(2000000000)))
				testBCoins := sdk.NewCoins(sdk.NewCoin(denomB, osmomath.NewInt(1000000000)))
				s.FundAcc(s.TestAccs[0], testACoins.Add(testBCoins...))
				s.JoinTransmuterPool(s.TestAccs[0], cwPool.GetId(), testACoins.Add(testBCoins...))

				return cwPool.GetId()
			},
			expectedUpdatedSkimPercent: []osmomath.Dec{
				defaultTakerFeeShareAgreements[0].SkimPercent.Mul(osmomath.MustNewDecFromStr("0.5")),
				defaultTakerFeeShareAgreements[1].SkimPercent.Mul(osmomath.MustNewDecFromStr("0.33333333333333333")),
				defaultTakerFeeShareAgreements[2].SkimPercent.Mul(osmomath.MustNewDecFromStr("0.16666666666666666")),
			},
			expectedError: nil,
		},
		"1:1:1 to 4:2:1": {
			setupFunc: func() uint64 {
				cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{denomA, denomB, denomC}, []uint16{1, 1, 1})
				setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:3])
				err := s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
				s.Require().NoError(err)

				// Change the ratio to 4:2:1 by adding more of denomA and denomB
				testACoins := sdk.NewCoins(sdk.NewCoin(denomA, osmomath.NewInt(3000000000)))
				testBCoins := sdk.NewCoins(sdk.NewCoin(denomB, osmomath.NewInt(1000000000)))
				s.FundAcc(s.TestAccs[0], testACoins.Add(testBCoins...))
				s.JoinTransmuterPool(s.TestAccs[0], cwPool.GetId(), testACoins.Add(testBCoins...))

				return cwPool.GetId()
			},
			expectedUpdatedSkimPercent: []osmomath.Dec{
				defaultTakerFeeShareAgreements[0].SkimPercent.Mul(osmomath.MustNewDecFromStr("0.5714285714285714")),
				defaultTakerFeeShareAgreements[1].SkimPercent.Mul(osmomath.MustNewDecFromStr("0.2857142857142857")),
				defaultTakerFeeShareAgreements[2].SkimPercent.Mul(osmomath.MustNewDecFromStr("0.14285714285714285")),
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
				shareStates, err := s.App.PoolManagerKeeper.GetRegisteredAlloyedPoolFromPoolId(s.Ctx, poolId)
				s.Require().NoError(err)
				for i, shareState := range shareStates.TakerFeeShareAgreements {
					s.Require().Equal(tc.expectedUpdatedSkimPercent[i], shareState.SkimPercent)
				}
			}
		})
	}
}

func setTakerFeeShareAgreements(ctx sdk.Context, keeper *poolmanager.Keeper, agreements []types.TakerFeeShareAgreement) {
	for _, agreement := range agreements {
		err := keeper.SetTakerFeeShareAgreementForDenom(ctx, agreement)
		if err != nil {
			panic(err)
		}
	}
}

func modifySkimPercent(agreements []types.TakerFeeShareAgreement, factors []osmomath.Dec) []types.TakerFeeShareAgreement {
	if len(agreements) != len(factors) {
		panic("length of agreements and factors must match")
	}
	modified := make([]types.TakerFeeShareAgreement, len(agreements))
	for i, agreement := range agreements {
		modified[i] = agreement
		if factors[i].IsZero() {
			modified[i].SkimPercent = osmomath.ZeroDec()
		} else {
			modified[i].SkimPercent = agreement.SkimPercent.Mul(factors[i])
		}
	}
	return modified
}

func createExpectedTakerFeeShareAgreementsMap(agreements []types.TakerFeeShareAgreement) map[string]types.TakerFeeShareAgreement {
	expectedTakerFeeShareAgreements := make(map[string]types.TakerFeeShareAgreement)
	for _, agreement := range agreements {
		expectedTakerFeeShareAgreements[agreement.Denom] = agreement
	}
	return expectedTakerFeeShareAgreements
}

func createAlloyedDenom(address string) string {
	return fmt.Sprintf("factory/%s/alloyed/%s", address, apptesting.DefaultAlloyedSubDenom)
}
