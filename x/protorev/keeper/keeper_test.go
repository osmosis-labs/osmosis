package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/protorev/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}

// SetUpOsmoPools sets up the Osmo pools for testing
func (suite *KeeperTestSuite) SetUpOsmoPools() {
	suite.App.AppKeepers.ProtoRevKeeper.SetOsmoPool(suite.Ctx, "Akash", 1)
	suite.App.AppKeepers.ProtoRevKeeper.SetOsmoPool(suite.Ctx, "Juno", 2)
	suite.App.AppKeepers.ProtoRevKeeper.SetOsmoPool(suite.Ctx, "Ethereum", 3)
	suite.App.AppKeepers.ProtoRevKeeper.SetOsmoPool(suite.Ctx, "Bitcoin", 4)
	suite.App.AppKeepers.ProtoRevKeeper.SetOsmoPool(suite.Ctx, "Canto", 5)
}

// SetUpAtomPools sets up the Atom pools for testing
func (suite *KeeperTestSuite) SetUpAtomPools() {
	suite.App.AppKeepers.ProtoRevKeeper.SetAtomPool(suite.Ctx, "Akash", 6)
	suite.App.AppKeepers.ProtoRevKeeper.SetAtomPool(suite.Ctx, "Juno", 7)
	suite.App.AppKeepers.ProtoRevKeeper.SetAtomPool(suite.Ctx, "Ethereum", 8)
	suite.App.AppKeepers.ProtoRevKeeper.SetAtomPool(suite.Ctx, "Bitcoin", 9)
	suite.App.AppKeepers.ProtoRevKeeper.SetAtomPool(suite.Ctx, "Canto", 10)
}

// SetUpRoutes sets up the routes for testing
func (suite *KeeperTestSuite) SetUpSearcherRoutes() {
	var index uint64
	for ; index <= 5; index++ {
		// create routes with atom
		searcherRoutes := CreateSeacherRoutes("atom", 5)
		suite.App.AppKeepers.ProtoRevKeeper.SetSearcherRoute(suite.Ctx, index, &searcherRoutes)

		// create routes with osmo
		searcherRoutes = CreateSeacherRoutes("osmo", 5)
		suite.App.AppKeepers.ProtoRevKeeper.SetSearcherRoute(suite.Ctx, index*2, &searcherRoutes)
	}
}

// CreateRoute creates SearchRoutes object for testing
func CreateSeacherRoutes(arbDenom string, numberRoutes uint64) types.SearcherRoutes {
	routes := make([]*types.Route, numberRoutes)
	for i := uint64(0); i < numberRoutes; i++ {
		routes[i] = &types.Route{
			Pools: []uint64{i, i + 1, i + 2, i + 3, i + 4},
		}
	}

	return types.NewSearcherRoutes(arbDenom, routes)
}
