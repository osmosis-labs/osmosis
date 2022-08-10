package keeper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"

	minttypes "github.com/osmosis-labs/osmosis/v10/x/mint/types"
	"github.com/osmosis-labs/osmosis/v10/x/pool-incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
)

func (suite *KeeperTestSuite) TestAllocateAsset() {
	tests := []struct {
		name                   string
		testingDistrRecord     []types.DistrRecord
		mintedCoins            sdk.Coin
		expectedGaugesBalances []sdk.Coins
		expectedFeeCollector   sdk.Coin
		expectedCommunityPool  sdk.DecCoin
	}{
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
			mintedCoins: sdk.NewCoin("stake", sdk.NewInt(50000)),
			expectedGaugesBalances: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(2500))),
				sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(4999))),
				sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(7500))),
			},
			expectedFeeCollector:  sdk.NewCoin("stake", sdk.NewInt(20000)),
			expectedCommunityPool: sdk.NewDecCoin("stake", sdk.NewInt(5000)),
		},
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
			mintedCoins: sdk.NewCoin("stake", sdk.NewInt(100000)),
			expectedGaugesBalances: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(0))),
				sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(3000))),
				sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(6000))),
			},
			expectedFeeCollector:  sdk.NewCoin("stake", sdk.NewInt(40000)),
			expectedCommunityPool: sdk.NewDecCoin("stake", sdk.NewInt(31000)),
		},
		{
			name:                   "community pool distribution when no distribution records are set",
			testingDistrRecord:                   []types.DistrRecord{},
			mintedCoins:            sdk.NewCoin("stake", sdk.NewInt(100000)),
			expectedGaugesBalances: []sdk.Coins{},
			expectedFeeCollector:   sdk.NewCoin("stake", sdk.NewInt(40000)),
			expectedCommunityPool:  sdk.NewDecCoin("stake", sdk.NewInt(40000)),
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.Setup()
			keeper := suite.App.PoolIncentivesKeeper
			mintKeeper := suite.App.MintKeeper
			params := suite.App.MintKeeper.GetParams(suite.Ctx)
			params.WeightedDeveloperRewardsReceivers = []minttypes.WeightedAddress{
				{
					Address: sdk.AccAddress([]byte("addr1---------------")).String(),
					Weight:  sdk.NewDec(1),
				},
			}
			suite.App.MintKeeper.SetParams(suite.Ctx, params)

			suite.PrepareBalancerPool()

			// LockableDurations should be 1, 3, 7 hours from the default genesis state.
			lockableDurations := keeper.GetLockableDurations(suite.Ctx)
			suite.Equal(3, len(lockableDurations))

			for i, duration := range lockableDurations {
				suite.Equal(duration, types.DefaultGenesisState().GetLockableDurations()[i])
			}

			feePoolOrigin := suite.App.DistrKeeper.GetFeePool(suite.Ctx)

			// Create record
			err := keeper.ReplaceDistrRecords(suite.Ctx, test.testingDistrRecord...)
			suite.Require().NoError(err)

			err = mintKeeper.MintCoins(suite.Ctx, sdk.NewCoins(test.mintedCoins))
			suite.Require().NoError(err)

			err = mintKeeper.DistributeMintedCoin(suite.Ctx, test.mintedCoins) // this calls AllocateAsset via hook
			suite.Require().NoError(err)
			distribution.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{}, *suite.App.DistrKeeper)

			suite.Require().Equal(test.expectedFeeCollector, suite.App.BankKeeper.GetBalance(suite.Ctx, suite.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake"))
			for i := 0; i < len(test.testingDistrRecord); i++ {
				if test.testingDistrRecord[i].GaugeId == 0 {
					continue
				}
				gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, test.testingDistrRecord[i].GaugeId)
				suite.Require().NoError(err)
				suite.Require().Equal(test.expectedGaugesBalances[i], gauge.Coins)
			}

			feePoolNew := suite.App.DistrKeeper.GetFeePool(suite.Ctx)
			suite.Require().Equal(feePoolOrigin.CommunityPool.Add(test.expectedCommunityPool), feePoolNew.CommunityPool)
		})
	}
}

func (suite *KeeperTestSuite) TestReplaceDistrRecords() {
	tests := []struct {
		name           string
		testingDistrRecord           []types.DistrRecord
		isPoolPrepared bool
		expectErr      bool
		expectTotalWeight sdk.Int
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
			isPoolPrepared: true,
			expectErr:      false,
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
			isPoolPrepared: true,
			expectErr:      false,
			expectTotalWeight: sdk.NewInt(300),
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()
			keeper := suite.App.PoolIncentivesKeeper

			if test.isPoolPrepared {
				suite.PrepareBalancerPool()
			}

			err := keeper.ReplaceDistrRecords(suite.Ctx, test.testingDistrRecord...)
			if test.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)

				distrInfo := keeper.GetDistrInfo(suite.Ctx)
				suite.Require().Equal(len(test.testingDistrRecord), len(distrInfo.Records))
				for i, record := range test.testingDistrRecord {
					suite.Require().Equal(record.Weight, distrInfo.Records[i].Weight)
				}
				suite.Require().Equal(test.expectTotalWeight, distrInfo.TotalWeight)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUpdateDistrRecords() {
	tests := []struct {
		name           string
		testingDistrRecord           []types.DistrRecord
		isPoolPrepared bool
		expectErr      bool
		expectTotalWeight sdk.Int
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
			isPoolPrepared: true,
			expectErr:      false,
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
			isPoolPrepared: true,
			expectErr:      false,
			expectTotalWeight: sdk.NewInt(300),
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()
			keeper := suite.App.PoolIncentivesKeeper

			if test.isPoolPrepared {
				suite.PrepareBalancerPool()
			}

			err := keeper.UpdateDistrRecords(suite.Ctx, test.testingDistrRecord...)
			if test.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)

				distrInfo := keeper.GetDistrInfo(suite.Ctx)
				suite.Require().Equal(len(test.testingDistrRecord), len(distrInfo.Records))
				for i, record := range test.testingDistrRecord {
					suite.Require().Equal(record.Weight, distrInfo.Records[i].Weight)
				}
				suite.Require().Equal(test.expectTotalWeight, distrInfo.TotalWeight)
			}
		})
	}
}
