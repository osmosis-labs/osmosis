package v9_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v20/app/apptesting"
	v9 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v9"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestProp214() {
	poolId := s.PrepareBalancerPool()
	v9.ExecuteProp214(s.Ctx, s.App.GAMMKeeper)

	_, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
	s.Require().NoError(err)

	// Kept as comments for recordkeeping. Since SetPool is now private, the changes being tested for can no longer be made:
	// 		spreadFactor := pool.GetSpreadFactor(s.Ctx)
	//  	expectedSpreadFactor := osmomath.MustNewDecFromStr("0.002")
	//
	//  	s.Require().Equal(expectedSpreadFactor, spreadFactor)
}
