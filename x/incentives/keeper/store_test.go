package keeper_test

import "github.com/stretchr/testify/suite"

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
