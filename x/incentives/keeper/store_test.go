package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func (suite *KeeperTestSuite) TestGaugeReferencesManagement() {
	key1 := []byte{0x11}
	key2 := []byte{0x12}

	suite.SetupTest()
	_ = suite.App.IncentivesKeeper.AddGaugeRefByKey(suite.Ctx, key1, 1)
	_ = suite.App.IncentivesKeeper.AddGaugeRefByKey(suite.Ctx, key2, 1)
	_ = suite.App.IncentivesKeeper.AddGaugeRefByKey(suite.Ctx, key1, 2)
	_ = suite.App.IncentivesKeeper.AddGaugeRefByKey(suite.Ctx, key2, 2)
	_ = suite.App.IncentivesKeeper.AddGaugeRefByKey(suite.Ctx, key2, 3)

	gaugeRefs1 := suite.App.IncentivesKeeper.GetGaugeRefs(suite.Ctx, key1)
	suite.Require().Equal(len(gaugeRefs1), 2)
	gaugeRefs2 := suite.App.IncentivesKeeper.GetGaugeRefs(suite.Ctx, key2)
	suite.Require().Equal(len(gaugeRefs2), 3)

	err := suite.App.IncentivesKeeper.DeleteGaugeRefByKey(suite.Ctx, key2, 1)
	suite.Require().NoError(err)

	gaugeRefs3 := suite.App.IncentivesKeeper.GetGaugeRefs(suite.Ctx, key2)
	suite.Require().Equal(len(gaugeRefs3), 2)
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
