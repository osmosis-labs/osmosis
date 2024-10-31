package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

var _ = suite.TestingSuite(nil)

func (s *KeeperTestSuite) TestGaugeReferencesManagement() {
	key1 := []byte{0x11}
	key2 := []byte{0x12}

	s.SetupTest()

	// set two gauge references to key 1 and three gauge references to key 2
	_ = s.App.IncentivesKeeper.AddGaugeRefByKey(s.Ctx, key1, 1)
	_ = s.App.IncentivesKeeper.AddGaugeRefByKey(s.Ctx, key2, 1)
	_ = s.App.IncentivesKeeper.AddGaugeRefByKey(s.Ctx, key1, 2)
	_ = s.App.IncentivesKeeper.AddGaugeRefByKey(s.Ctx, key2, 2)
	_ = s.App.IncentivesKeeper.AddGaugeRefByKey(s.Ctx, key2, 3)

	// ensure key1 only has 2 entries
	gaugeRefs1 := s.App.IncentivesKeeper.GetGaugeRefs(s.Ctx, key1)
	s.Require().Equal(len(gaugeRefs1), 2)

	// ensure key2 only has 3 entries
	gaugeRefs2 := s.App.IncentivesKeeper.GetGaugeRefs(s.Ctx, key2)
	s.Require().Equal(len(gaugeRefs2), 3)

	// remove gauge 1 from key2, resulting in a reduction from 3 to 2 entries
	err := s.App.IncentivesKeeper.DeleteGaugeRefByKey(s.Ctx, key2, 1)
	s.Require().NoError(err)

	// ensure key2 now only has 2 entries
	gaugeRefs3 := s.App.IncentivesKeeper.GetGaugeRefs(s.Ctx, key2)
	s.Require().Equal(len(gaugeRefs3), 2)
}

func (s *KeeperTestSuite) TestGetGroupByGaugeID() {
	// TODO: Re-enable this once gauge creation refactor is complete in https://github.com/osmosis-labs/osmosis/issues/6404
	s.T().Skip()

	tests := map[string]struct {
		groupGaugeId   uint64
		expectedRecord types.Group
	}{
		"Valid record": {
			groupGaugeId: uint64(5),
			expectedRecord: types.Group{
				GroupGaugeId: uint64(5),
				InternalGaugeInfo: types.InternalGaugeInfo{
					TotalWeight: osmomath.NewInt(150),
					GaugeRecords: []types.InternalGaugeRecord{
						{
							GaugeId:          2,
							CurrentWeight:    osmomath.NewInt(50),
							CumulativeWeight: osmomath.NewInt(50),
						},
						{
							GaugeId:          3,
							CurrentWeight:    osmomath.NewInt(50),
							CumulativeWeight: osmomath.NewInt(50),
						},
						{
							GaugeId:          4,
							CurrentWeight:    osmomath.NewInt(50),
							CumulativeWeight: osmomath.NewInt(50),
						},
					},
				},
				SplittingPolicy: types.ByVolume,
			},
		},

		"InValid record": {
			groupGaugeId:   uint64(6),
			expectedRecord: types.Group{},
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.FundAcc(s.TestAccs[1], sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100_000_000)))) // 1,000 osmo
			clPool := s.PrepareConcentratedPool()                                                                     // gaugeid = 1

			// create 3 internal Gauge
			var internalGauges []uint64
			for i := 0; i <= 2; i++ {
				internalGauge := s.CreateNoLockExternalGauges(clPool.GetId(), sdk.NewCoins(), s.TestAccs[1], uint64(1)) // gauge id = 2,3,4
				internalGauges = append(internalGauges, internalGauge)
			}

			_, err := s.App.IncentivesKeeper.CreateGroup(s.Ctx, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100_000_000))), 1, s.TestAccs[1], internalGauges) // gauge id = 5
			s.Require().NoError(err)

			record, err := s.App.IncentivesKeeper.GetGroupByGaugeID(s.Ctx, test.groupGaugeId)
			s.Require().NoError(err)

			s.Require().Equal(test.expectedRecord, record)
		})
	}
}

func (s *KeeperTestSuite) TestGetAllGroupsWithGauge() {
	groupPools := s.PrepareAllSupportedPools()
	groupPoolIds := []uint64{groupPools.ConcentratedPoolID, groupPools.BalancerPoolID, groupPools.StableSwapPoolID}

	s.overwriteVolumes(groupPoolIds, []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount, defaultVolumeAmount})
	expectedStartTime := s.Ctx.BlockTime().UTC()
	_, err := s.App.IncentivesKeeper.CreateGroup(s.Ctx, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100_000_000))), 1, s.TestAccs[0], groupPoolIds)
	s.Require().NoError(err)

	// Call GetAllGroupsWithGauge
	groupsWithGauge, err := s.App.IncentivesKeeper.GetAllGroupsWithGauge(s.Ctx)
	s.Require().NoError(err)

	// Check the length of the returned slice
	s.Require().Equal(1, len(groupsWithGauge))

	// Check the content of the returned slice
	expectedGroupsWithGauge := types.GroupsWithGauge{
		Group: types.Group{
			GroupGaugeId: uint64(8),
			InternalGaugeInfo: types.InternalGaugeInfo{
				TotalWeight: osmomath.NewInt(900),
				GaugeRecords: []types.InternalGaugeRecord{
					// Concentrated Pool (1)
					{
						GaugeId:          1,
						CurrentWeight:    osmomath.NewInt(300),
						CumulativeWeight: osmomath.NewInt(300),
					},
					// Balancer Pool (2-4)
					{
						GaugeId:          4,
						CurrentWeight:    osmomath.NewInt(300),
						CumulativeWeight: osmomath.NewInt(300),
					},
					// Stable Pool (5-7)
					{
						GaugeId:          7,
						CurrentWeight:    osmomath.NewInt(300),
						CumulativeWeight: osmomath.NewInt(300),
					},
				},
			},
			SplittingPolicy: types.ByVolume,
		},
		Gauge: types.Gauge{
			Id:                uint64(8),
			DistributeTo:      lockuptypes.QueryCondition{LockQueryType: lockuptypes.ByGroup},
			Coins:             sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100_000_000))),
			StartTime:         expectedStartTime,
			NumEpochsPaidOver: 1,
		},
	}
	s.Require().Equal(expectedGroupsWithGauge.String(), groupsWithGauge[0].String())
}
