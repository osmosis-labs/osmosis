package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
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

	// ensure key1 only has 2 entires
	gaugeRefs1 := s.App.IncentivesKeeper.GetGaugeRefs(s.Ctx, key1)
	s.Require().Equal(len(gaugeRefs1), 2)

	// ensure key2 only has 3 entries
	gaugeRefs2 := s.App.IncentivesKeeper.GetGaugeRefs(s.Ctx, key2)
	s.Require().Equal(len(gaugeRefs2), 3)

	// remove gauge 1 from key2, resulting in a reduction from 3 to 2 entries
	err := s.App.IncentivesKeeper.DeleteGaugeRefByKey(s.Ctx, key2, 1)
	s.Require().NoError(err)

	// ensure key2 now only has 2 entires
	gaugeRefs3 := s.App.IncentivesKeeper.GetGaugeRefs(s.Ctx, key2)
	s.Require().Equal(len(gaugeRefs3), 2)
}

func (s *KeeperTestSuite) TestGetGroupGaugeById() {
	tests := map[string]struct {
		groupGaugeId   uint64
		expectedRecord types.GroupGauge
	}{
		"Valid record": {
			groupGaugeId: uint64(5),
			expectedRecord: types.GroupGauge{
				GroupGaugeId:    uint64(5),
				InternalIds:     []uint64{2, 3, 4},
				SplittingPolicy: types.Evenly,
			},
		},

		"InValid record": {
			groupGaugeId:   uint64(6),
			expectedRecord: types.GroupGauge{},
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.FundAcc(s.TestAccs[1], sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(100_000_000)))) // 1,000 osmo
			clPool := s.PrepareConcentratedPool()                                                      // gaugeid = 1

			// create 3 internal Gauge
			var internalGauges []uint64
			for i := 0; i <= 2; i++ {
				internalGauge := s.CreateNoLockExternalGauges(clPool.GetId(), sdk.NewCoins(), s.TestAccs[1], uint64(1)) // gauge id = 2,3,4
				internalGauges = append(internalGauges, internalGauge)
			}

			_, err := s.App.IncentivesKeeper.CreateGroupGauge(s.Ctx, sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(100_000_000))), 1, s.TestAccs[1], internalGauges, lockuptypes.ByGroup, types.Evenly) // gauge id = 5
			s.Require().NoError(err)

			record, err := s.App.IncentivesKeeper.GetGroupGaugeById(s.Ctx, test.groupGaugeId)
			s.Require().NoError(err)

			s.Require().Equal(test.expectedRecord, record)
		})
	}
}
