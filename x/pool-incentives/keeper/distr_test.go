package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v19/x/pool-incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
					Weight:  sdk.NewInt(100),
				},
				{
					GaugeId: 2,
					Weight:  sdk.NewInt(200),
				},
				{
					GaugeId: 3,
					Weight:  sdk.NewInt(300),
				},
			},
			mintedCoins: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(15000)),
			expectedGaugesBalances: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(2500))),
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(4999))),
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(7500))),
			},
			expectedCommunityPool: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(0)),
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
					Weight:  sdk.NewInt(700),
				},
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(100),
				},
				{
					GaugeId: 2,
					Weight:  sdk.NewInt(200),
				},
			},
			mintedCoins: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(30000)),
			expectedGaugesBalances: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0))),
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(3000))),
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(6000))),
			},
			expectedCommunityPool: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(21000)),
		},
		// With minting 30000 stake to module, after AllocateAsset we get:
		// 	expectedCommunityPool = 30000 (Cause there are no gauges, all rewards are transferred to the community pool)
		{
			name:                   "community pool distribution when no distribution records are set",
			testingDistrRecord:     []types.DistrRecord{},
			mintedCoins:            sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(30000)),
			expectedGaugesBalances: []sdk.Coins{},
			expectedCommunityPool:  sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(30000)),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Setup()
			keeper := s.App.PoolIncentivesKeeper
			s.FundModuleAcc(types.ModuleName, sdk.NewCoins(test.mintedCoins))
			s.PrepareBalancerPool()

			// LockableDurations should be 1, 3, 7 hours from the default genesis state.
			lockableDurations := keeper.GetLockableDurations(s.Ctx)
			s.Equal(3, len(lockableDurations))

			for i, duration := range lockableDurations {
				s.Equal(duration, types.DefaultGenesisState().GetLockableDurations()[i])
			}

			feePoolOrigin := s.App.DistrKeeper.GetFeePool(s.Ctx)

			// Create record
			err := keeper.ReplaceDistrRecords(s.Ctx, test.testingDistrRecord...)
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

			feePoolNew := s.App.DistrKeeper.GetFeePool(s.Ctx)
			s.Require().Equal(feePoolOrigin.CommunityPool.Add(test.expectedCommunityPool), feePoolNew.CommunityPool)
		})
	}
}

func (s *KeeperTestSuite) TestReplaceDistrRecords() {
	tests := []struct {
		name               string
		testingDistrRecord []types.DistrRecord
		isPoolPrepared     bool
		expectErr          bool
		expectTotalWeight  sdk.Int
	}{
		{
			name: "Not existent gauge.",
			testingDistrRecord: []types.DistrRecord{{
				GaugeId: 1,
				Weight:  sdk.NewInt(100),
			}},
			isPoolPrepared: false,
			expectErr:      true,
		},
		{
			name: "Adding two of the same gauge id at once should error",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(200),
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
					Weight:  sdk.NewInt(200),
				},
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(250),
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
					Weight:  sdk.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(100),
				},
			},
			isPoolPrepared:    true,
			expectErr:         false,
			expectTotalWeight: sdk.NewInt(200),
		},
		{
			name: "With different weights",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  sdk.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(200),
				},
			},
			isPoolPrepared:    true,
			expectErr:         false,
			expectTotalWeight: sdk.NewInt(300),
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

func (s *KeeperTestSuite) TestUpdateDistrRecords() {
	tests := []struct {
		name               string
		testingDistrRecord []types.DistrRecord
		isPoolPrepared     bool
		expectErr          bool
		expectTotalWeight  sdk.Int
	}{
		{
			name: "Not existent gauge.",
			testingDistrRecord: []types.DistrRecord{{
				GaugeId: 1,
				Weight:  sdk.NewInt(100),
			}},
			isPoolPrepared: false,
			expectErr:      true,
		},
		{
			name: "Adding two of the same gauge id at once should error",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(100),
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
					Weight:  sdk.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(200),
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
					Weight:  sdk.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(100),
				},
			},
			isPoolPrepared:    true,
			expectErr:         false,
			expectTotalWeight: sdk.NewInt(200),
		},
		{
			name: "With different weights",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  sdk.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(200),
				},
			},
			isPoolPrepared:    true,
			expectErr:         false,
			expectTotalWeight: sdk.NewInt(300),
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
