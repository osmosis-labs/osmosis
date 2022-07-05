package keeper_test

import "github.com/stretchr/testify/suite"

func (suite *KeeperTestSuite) TestGaugeReferencesManagement() {
	key1 := []byte{0x11}
	key2 := []byte{0x12}

	suite.SetupTest()

	// set two gauge references to key 1 and three gauge references to key 2
	_ = suite.App.IncentivesKeeper.AddGaugeRefByKey(suite.Ctx, key1, 1)
	_ = suite.App.IncentivesKeeper.AddGaugeRefByKey(suite.Ctx, key2, 1)
	_ = suite.App.IncentivesKeeper.AddGaugeRefByKey(suite.Ctx, key1, 2)
	_ = suite.App.IncentivesKeeper.AddGaugeRefByKey(suite.Ctx, key2, 2)
	_ = suite.App.IncentivesKeeper.AddGaugeRefByKey(suite.Ctx, key2, 3)

	// ensure key1 only has 2 entires
	gaugeRefs1 := suite.App.IncentivesKeeper.GetGaugeRefs(suite.Ctx, key1)
	suite.Require().Equal(len(gaugeRefs1), 2)

	// ensure key2 only has 3 entries
	gaugeRefs2 := suite.App.IncentivesKeeper.GetGaugeRefs(suite.Ctx, key2)
	suite.Require().Equal(len(gaugeRefs2), 3)

	// remove gauge 1 from key2, resulting in a reduction from 3 to 2 entries
	err := suite.App.IncentivesKeeper.DeleteGaugeRefByKey(suite.Ctx, key2, 1)
	suite.Require().NoError(err)

	// ensure key2 now only has 2 entires
	gaugeRefs3 := suite.App.IncentivesKeeper.GetGaugeRefs(suite.Ctx, key2)
	suite.Require().Equal(len(gaugeRefs3), 2)
}

var _ = suite.TestingSuite(nil)
