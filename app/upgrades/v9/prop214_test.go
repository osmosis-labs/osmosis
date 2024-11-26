package v9_test

import (
	"testing"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/x/upgrade"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	v9 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v9"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))
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
