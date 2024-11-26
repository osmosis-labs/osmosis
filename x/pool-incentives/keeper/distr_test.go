package keeper_test

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/coinutil"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var defaultCoins = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(10000)))

func (s *KeeperTestSuite) TestAllocateAsset() {
	tests := []struct {
		name                   string
		testingDistrRecord     []types.DistrRecord
		mintedCoins            sdk.Coin
		expectedGaugesBalances []sdk.Coins
		expectedCommunityPool  sdk.DecCoin
	}{
		// With minting 15000 stake to module, after AllocateAsset we get:
		// expectedCommunityPool = 0 (All reward will be transferred to the gauges)
		// 	expectedGaugesBalances in order:
		//    gaue1_balance = 15000 * 100/(100+200+300) = 2500
		//    gaue2_balance = 15000 * 200/(100+200+300) = 5000 (using the formula in the function gives the exact result 4999,9999999999995000. But TruncateInt return 4999. Is this the issue?)
		//    gaue3_balance = 15000 * 300/(100+200+300) = 7500
		{
			name: "Allocated to the gauges proportionally",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: 2,
					Weight:  osmomath.NewInt(200),
				},
				{
					GaugeId: 3,
					Weight:  osmomath.NewInt(300),
				},
			},
			mintedCoins: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(15000)),
			expectedGaugesBalances: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(2500))),
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(4999))),
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(7500))),
			},
			expectedCommunityPool: sdk.NewDecCoin(sdk.DefaultBondDenom, osmomath.NewInt(0)),
		},

		// With minting 30000 stake to module, after AllocateAsset we get:
		// 	expectedCommunityPool = 30000 * 700/(700+200+100) = 21000 stake (Cause gaugeId=0 the reward will be transferred to the community pool)
		// 	expectedGaugesBalances in order:
		//    gaue1_balance = 30000 * 100/(700+200+100) = 3000
		//    gaue2_balance = 30000 * 200/(700+200+100) = 6000
		{
			name: "Community pool distribution when gaugeId is zero",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  osmomath.NewInt(700),
				},
				{
					GaugeId: 1,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: 2,
					Weight:  osmomath.NewInt(200),
				},
			},
			mintedCoins: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(30000)),
			expectedGaugesBalances: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(0))),
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(3000))),
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(6000))),
			},
			expectedCommunityPool: sdk.NewDecCoin(sdk.DefaultBondDenom, osmomath.NewInt(21000)),
		},
		// With minting 30000 stake to module, after AllocateAsset we get:
		// 	expectedCommunityPool = 30000 (Cause there are no gauges, all rewards are transferred to the community pool)
		{
			name:                   "community pool distribution when no distribution records are set",
			testingDistrRecord:     []types.DistrRecord{},
			mintedCoins:            sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(30000)),
			expectedGaugesBalances: []sdk.Coins{},
			expectedCommunityPool:  sdk.NewDecCoin(sdk.DefaultBondDenom, osmomath.NewInt(30000)),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Setup()

			// Since this test creates or adds to a gauge, we need to ensure a route exists in protorev hot routes.
			// The pool doesn't need to actually exist for this test, so we can just ensure the denom pair has some entry.
			s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, sdk.DefaultBondDenom, 9999)

			keeper := s.App.PoolIncentivesKeeper
			s.FundModuleAcc(types.ModuleName, sdk.NewCoins(test.mintedCoins))
			s.PrepareBalancerPool()

			// LockableDurations should be 1, 3, 7 hours from the default genesis state.
			lockableDurations := keeper.GetLockableDurations(s.Ctx)
			s.Equal(3, len(lockableDurations))

			for i, duration := range lockableDurations {
				s.Equal(duration, types.DefaultGenesisState().GetLockableDurations()[i])
			}

			feePoolOrigin, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)

			// Create record
			err = keeper.ReplaceDistrRecords(s.Ctx, test.testingDistrRecord...)
			s.Require().NoError(err)

			err = keeper.AllocateAsset(s.Ctx)
			s.Require().NoError(err)

			for i := 0; i < len(test.testingDistrRecord); i++ {
				if test.testingDistrRecord[i].GaugeId == 0 {
					continue
				}
				gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, test.testingDistrRecord[i].GaugeId)
				s.Require().NoError(err)
				s.Require().Equal(test.expectedGaugesBalances[i], gauge.Coins)
			}

			feePoolNew, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
			s.Require().Equal(feePoolOrigin.CommunityPool.Add(test.expectedCommunityPool), feePoolNew.CommunityPool)
		})
	}
}

// Validates that group gauges can be allocated minted tokens from pool incentives as expected
// The test creates 2 groups, creates distribution records for them, calls AllocateAsset and then
// checks that the group gauges have the expected amount of tokens.
func (s *KeeperTestSuite) TestAllocateAsset_GroupGauge() {
	var (
		weightGroupOne = osmomath.NewInt(100)
		weightGroupTwo = osmomath.NewInt(200)
		totalWeight    = weightGroupTwo.Add(weightGroupOne)

		poolIncentiveDistribution = defaultCoins.Add(defaultCoins...)
	)

	s.Setup()

	// Since this test creates or adds to a gauge, we need to ensure a route exists in protorev hot routes.
	// The pool doesn't need to actually exist for this test, so we can just ensure the denom pair has some entry.
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, sdk.DefaultBondDenom, 9999)

	poolInfo := s.PrepareAllSupportedPools()

	poolIDs := []uint64{poolInfo.BalancerPoolID, poolInfo.ConcentratedPoolID, poolInfo.StableSwapPoolID}

	// Fund fee and initial coins for each group.
	groupCreationFee := s.App.IncentivesKeeper.GetParams(s.Ctx).GroupCreationFee
	s.FundAcc(s.TestAccs[1], groupCreationFee.Add(groupCreationFee...).Add(defaultCoins...).Add(defaultCoins...))

	// Setup initial volume for each pool.
	for _, poolID := range poolIDs {
		s.App.PoolManagerKeeper.SetVolume(s.Ctx, poolID, defaultCoins)
	}

	groupGaugeIDOne, err := s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins, 0, s.TestAccs[1], poolIDs)
	s.Require().NoError(err)

	groupGaugeIDTwo, err := s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins, 0, s.TestAccs[1], poolIDs)
	s.Require().NoError(err)

	err = s.App.PoolIncentivesKeeper.ReplaceDistrRecords(s.Ctx, types.DistrRecord{
		GaugeId: groupGaugeIDOne,
		Weight:  weightGroupOne,
	}, types.DistrRecord{
		GaugeId: groupGaugeIDTwo,
		Weight:  weightGroupTwo,
	})
	s.Require().NoError(err)

	// Fund pool incentives module account
	s.FundModuleAcc(types.ModuleName, poolIncentiveDistribution)

	// Allocate pool incentive distribution
	err = s.App.PoolIncentivesKeeper.AllocateAsset(s.Ctx)
	s.Require().NoError(err)

	// Get first Group Gauge and ensure that full amount is received.
	groupGaugeOne, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, groupGaugeIDOne)
	s.Require().NoError(err)

	// expected to contain initial default coins + 100 / 300 of the distribution
	expectedDistributionGroupGaugeOne := coinutil.MulDec(poolIncentiveDistribution, weightGroupOne.ToLegacyDec().Quo(totalWeight.ToLegacyDec()))
	s.Require().Equal(defaultCoins.Add(expectedDistributionGroupGaugeOne...), groupGaugeOne.Coins)

	// Get second Group Gauge and ensure that full amount is received.
	groupGaugeTwo, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, groupGaugeIDTwo)
	s.Require().NoError(err)

	// expected to contain initial default coins + 200 / 300 of the distribution
	expectedDistributionGroupGaugeTwo := coinutil.MulDec(poolIncentiveDistribution, weightGroupTwo.ToLegacyDec().Quo(totalWeight.ToLegacyDec()))
	s.Require().Equal(defaultCoins.Add(expectedDistributionGroupGaugeTwo...), groupGaugeTwo.Coins)
}

func (s *KeeperTestSuite) TestReplaceDistrRecords() {
	tests := []struct {
		name               string
		testingDistrRecord []types.DistrRecord
		isPoolPrepared     bool
		expectErr          bool
		expectTotalWeight  osmomath.Int
	}{
		{
			name: "Not existent gauge.",
			testingDistrRecord: []types.DistrRecord{{
				GaugeId: 1,
				Weight:  osmomath.NewInt(100),
			}},
			isPoolPrepared: false,
			expectErr:      true,
		},
		{
			name: "Adding two of the same gauge id at once should error",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  osmomath.NewInt(200),
				},
			},
			isPoolPrepared: true,
			expectErr:      true,
		},
		{
			name: "Adding unsort gauges at once should error",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 2,
					Weight:  osmomath.NewInt(200),
				},
				{
					GaugeId: 1,
					Weight:  osmomath.NewInt(250),
				},
			},
			isPoolPrepared: true,
			expectErr:      true,
		},
		{
			name: "Normal case with same weights",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  osmomath.NewInt(100),
				},
			},
			isPoolPrepared:    true,
			expectErr:         false,
			expectTotalWeight: osmomath.NewInt(200),
		},
		{
			name: "With different weights",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  osmomath.NewInt(200),
				},
			},
			isPoolPrepared:    true,
			expectErr:         false,
			expectTotalWeight: osmomath.NewInt(300),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			keeper := s.App.PoolIncentivesKeeper

			if test.isPoolPrepared {
				s.PrepareBalancerPool()
			}

			err := keeper.ReplaceDistrRecords(s.Ctx, test.testingDistrRecord...)
			if test.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				distrInfo := keeper.GetDistrInfo(s.Ctx)
				s.Require().Equal(len(test.testingDistrRecord), len(distrInfo.Records))
				for i, record := range test.testingDistrRecord {
					s.Require().Equal(record.Weight, distrInfo.Records[i].Weight)
				}
				s.Require().Equal(test.expectTotalWeight, distrInfo.TotalWeight)
			}
		})
	}
}

func (s *KeeperTestSuite) TestReplaceDistrRecords_GroupGauge() {
	s.SetupTest()
	keeper := s.App.PoolIncentivesKeeper

	poolGroup := s.PrepareAllSupportedPools()

	// Create perpetual group
	perpetualGroupGauge, err := s.App.IncentivesKeeper.CreateGroupAsIncentivesModuleAcc(s.Ctx, 0, []uint64{poolGroup.BalancerPoolID, poolGroup.ConcentratedPoolID, poolGroup.StableSwapPoolID})
	s.Require().NoError(err)

	// Create non-perpetual group
	nonPerpetualGroupGauge, err := s.App.IncentivesKeeper.CreateGroupAsIncentivesModuleAcc(s.Ctx, 1, []uint64{poolGroup.BalancerPoolID, poolGroup.ConcentratedPoolID, poolGroup.StableSwapPoolID})
	s.Require().NoError(err)

	// Initial state to use to ensure replace actually removes previous records.
	initialDistrRecords := []types.DistrRecord{
		{
			GaugeId: 0,
			Weight:  osmomath.NewInt(500),
		},
		{
			GaugeId: 1,
			Weight:  osmomath.NewInt(600),
		},
		{
			GaugeId: 2,
			Weight:  osmomath.NewInt(700),
		},
	}

	tests := []struct {
		name               string
		testingDistrRecord []types.DistrRecord
		expectedErr        error
		expectTotalWeight  osmomath.Int
	}{
		{
			name: "Perpetual group",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: perpetualGroupGauge,
					Weight:  osmomath.NewInt(100),
				},
			},
			expectTotalWeight: osmomath.NewInt(200),
		},
		{
			name: "Perpetual group, new weights",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: perpetualGroupGauge,
					Weight:  osmomath.NewInt(200),
				},
			},
			expectTotalWeight: osmomath.NewInt(300),
		},
		{
			name: "Error: non perpetual group",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: nonPerpetualGroupGauge,
					Weight:  osmomath.NewInt(200),
				},
			},
			expectedErr: errorsmod.Wrapf(types.ErrDistrRecordRegisteredGauge,
				"Gauge ID #%d is not perpetual.",
				nonPerpetualGroupGauge),
		},
		{
			name: "Error: perpetual and non perpetual group",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: perpetualGroupGauge,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: nonPerpetualGroupGauge,
					Weight:  osmomath.NewInt(100),
				},
			},
			expectedErr: errorsmod.Wrapf(types.ErrDistrRecordRegisteredGauge,
				"Gauge ID #%d is not perpetual.",
				nonPerpetualGroupGauge),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Set up distribution records prior to replace, to ensure replace deletes them.
			distrInfo := keeper.GetDistrInfo(s.Ctx)
			totalWeight := osmomath.NewInt(0)
			for _, record := range initialDistrRecords {
				totalWeight = totalWeight.Add(record.Weight)
			}
			distrInfo.Records = initialDistrRecords
			distrInfo.TotalWeight = totalWeight
			keeper.SetDistrInfo(s.Ctx, distrInfo)

			// System under test.
			err = keeper.ReplaceDistrRecords(s.Ctx, test.testingDistrRecord...)
			if test.expectedErr != nil {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				distrInfo := keeper.GetDistrInfo(s.Ctx)
				s.Require().Equal(len(test.testingDistrRecord), len(distrInfo.Records))
				for i, record := range test.testingDistrRecord {
					s.Require().Equal(record.Weight, distrInfo.Records[i].Weight)
				}
				s.Require().Equal(test.expectTotalWeight, distrInfo.TotalWeight)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUpdateDistrRecords() {
	tests := []struct {
		name               string
		testingDistrRecord []types.DistrRecord
		isPoolPrepared     bool
		expectErr          bool
		expectTotalWeight  osmomath.Int
	}{
		{
			name: "Not existent gauge.",
			testingDistrRecord: []types.DistrRecord{{
				GaugeId: 1,
				Weight:  osmomath.NewInt(100),
			}},
			isPoolPrepared: false,
			expectErr:      true,
		},
		{
			name: "Adding two of the same gauge id at once should error",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  osmomath.NewInt(100),
				},
			},
			isPoolPrepared: true,
			expectErr:      true,
		},
		{
			name: "Adding unsort gauges at once should error",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 2,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  osmomath.NewInt(200),
				},
			},
			isPoolPrepared: true,
			expectErr:      true,
		},
		{
			name: "Normal case with same weights",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  osmomath.NewInt(100),
				},
			},
			isPoolPrepared:    true,
			expectErr:         false,
			expectTotalWeight: osmomath.NewInt(200),
		},
		{
			name: "With different weights",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  osmomath.NewInt(200),
				},
			},
			isPoolPrepared:    true,
			expectErr:         false,
			expectTotalWeight: osmomath.NewInt(300),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			keeper := s.App.PoolIncentivesKeeper

			if test.isPoolPrepared {
				s.PrepareBalancerPool()
			}

			err := keeper.UpdateDistrRecords(s.Ctx, test.testingDistrRecord...)
			if test.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				distrInfo := keeper.GetDistrInfo(s.Ctx)
				s.Require().Equal(len(test.testingDistrRecord), len(distrInfo.Records))
				for i, record := range test.testingDistrRecord {
					s.Require().Equal(record.Weight, distrInfo.Records[i].Weight)
				}
				s.Require().Equal(test.expectTotalWeight, distrInfo.TotalWeight)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUpdateDistrRecords_GroupGauge() {
	s.SetupTest()
	keeper := s.App.PoolIncentivesKeeper

	poolGroup := s.PrepareAllSupportedPools()

	// Create perpetual group
	perpetualGroupGauge, err := s.App.IncentivesKeeper.CreateGroupAsIncentivesModuleAcc(s.Ctx, 0, []uint64{poolGroup.BalancerPoolID, poolGroup.ConcentratedPoolID, poolGroup.StableSwapPoolID})
	s.Require().NoError(err)

	// Create non-perpetual group
	nonPerpetualGroupGauge, err := s.App.IncentivesKeeper.CreateGroupAsIncentivesModuleAcc(s.Ctx, 1, []uint64{poolGroup.BalancerPoolID, poolGroup.ConcentratedPoolID, poolGroup.StableSwapPoolID})
	s.Require().NoError(err)

	tests := []struct {
		name                     string
		testingDistrRecord       []types.DistrRecord
		testingDistrRecordUpdate []types.DistrRecord
		expectedFinalDistrRecord []types.DistrRecord
		expectedErr              error
		expectTotalWeight        osmomath.Int
		expectedFinalTotalWeight osmomath.Int
	}{
		{
			name: "Perpetual group",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: perpetualGroupGauge,
					Weight:  osmomath.NewInt(100),
				},
			},
			expectTotalWeight: osmomath.NewInt(200),
			testingDistrRecordUpdate: []types.DistrRecord{
				{
					GaugeId: perpetualGroupGauge,
					Weight:  osmomath.NewInt(200),
				},
			},
			expectedFinalDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: perpetualGroupGauge,
					Weight:  osmomath.NewInt(200),
				},
			},
			expectedFinalTotalWeight: osmomath.NewInt(300),
		},
		{
			name: "Error: non perpetual group",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: nonPerpetualGroupGauge,
					Weight:  osmomath.NewInt(200),
				},
			},
			testingDistrRecordUpdate: []types.DistrRecord{
				{
					GaugeId: nonPerpetualGroupGauge,
					Weight:  osmomath.NewInt(200),
				},
			},
			expectedErr: errorsmod.Wrapf(types.ErrDistrRecordRegisteredGauge,
				"Gauge ID #%d is not perpetual.",
				nonPerpetualGroupGauge),
		},
		{
			name: "Error: perpetual and non perpetual group",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: perpetualGroupGauge,
					Weight:  osmomath.NewInt(100),
				},
				{
					GaugeId: nonPerpetualGroupGauge,
					Weight:  osmomath.NewInt(100),
				},
			},
			testingDistrRecordUpdate: []types.DistrRecord{
				{
					GaugeId: nonPerpetualGroupGauge,
					Weight:  osmomath.NewInt(200),
				},
			},
			expectedErr: errorsmod.Wrapf(types.ErrDistrRecordRegisteredGauge,
				"Gauge ID #%d is not perpetual.",
				nonPerpetualGroupGauge),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := keeper.UpdateDistrRecords(s.Ctx, test.testingDistrRecord...)
			if test.expectedErr != nil {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				distrInfo := keeper.GetDistrInfo(s.Ctx)
				s.Require().Equal(len(test.testingDistrRecord), len(distrInfo.Records))
				for i, record := range test.testingDistrRecord {
					s.Require().Equal(record.Weight, distrInfo.Records[i].Weight)
				}
				s.Require().Equal(test.expectTotalWeight, distrInfo.TotalWeight)
			}

			err = keeper.UpdateDistrRecords(s.Ctx, test.testingDistrRecordUpdate...)
			if test.expectedErr != nil {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				distrInfo := keeper.GetDistrInfo(s.Ctx)
				s.Require().Equal(len(test.expectedFinalDistrRecord), len(distrInfo.Records))
				for i, record := range test.expectedFinalDistrRecord {
					s.Require().Equal(record.Weight, distrInfo.Records[i].Weight)
				}
				s.Require().Equal(test.expectedFinalTotalWeight, distrInfo.TotalWeight)
			}
		})
	}
}
