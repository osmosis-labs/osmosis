package v13_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"

	v13 "github.com/osmosis-labs/osmosis/v12/app/upgrades/v13"
	gamm "github.com/osmosis-labs/osmosis/v12/x/gamm/keeper"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.Setup()
}

func TestUpgradeSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) TestMigrateNextPoolIdAndCreatePool() {
	suite.SetupTest() // reset

	const (
		expectedNextPoolId uint64 = 3
	)

	var (
		gammKeeperType = reflect.TypeOf(&gamm.Keeper{})
	)

	ctx := suite.Ctx
	gammKeeper := suite.App.GAMMKeeper
	swaprouterKeeper := suite.App.SwapRouterKeeper

	// Set next pool id to given constant, because creating pools doesn't
	// increment id on current version
	gammKeeper.SetNextPoolId(ctx, expectedNextPoolId)
	nextPoolId := gammKeeper.GetNextPoolId(ctx)
	suite.Require().Equal(expectedNextPoolId, nextPoolId)

	// system under test.
	v13.MigrateNextPoolId(ctx, gammKeeper, swaprouterKeeper)

	// validate swaprouter's next pool id.
	actualNextPoolId := swaprouterKeeper.GetNextPoolId(ctx)
	suite.Require().Equal(expectedNextPoolId, actualNextPoolId)

	// validate gamm pool count.
	actualGammPoolCount := gammKeeper.GetPoolCount(ctx)
	suite.Require().Equal(expectedNextPoolId-1, actualGammPoolCount)

	// create a pool after migration.
	actualCreatedPoolId := suite.PrepareBalancerPool()
	suite.Require().Equal(expectedNextPoolId, actualCreatedPoolId)

	// validate that module route mapping has been created for each pool id.
	for poolId := uint64(1); poolId < expectedNextPoolId; poolId++ {
		swapModule, err := swaprouterKeeper.GetSwapModule(ctx, poolId)
		suite.Require().NoError(err)

		suite.Require().Equal(gammKeeperType, reflect.TypeOf(swapModule))
	}
}
