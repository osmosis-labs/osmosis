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

// SetUpRoutes sets up the searcher routes for testing
func (suite *KeeperTestSuite) SetUpSearcherRoutes() {
	tokens := []string{"Akash", "Juno", "Ethereum", "Bitcoin", "Canto"}

	for _, token := range tokens {
		// create routes with atom
		searcherRoutes := CreateSeacherRoutes(5, "atom", token)
		suite.App.AppKeepers.ProtoRevKeeper.SetSearcherRoutes(suite.Ctx, "atom", token, &searcherRoutes)
	}

	for _, token := range tokens {
		// create routes with osmo
		searcherRoutes := CreateSeacherRoutes(5, "osmo", token)
		suite.App.AppKeepers.ProtoRevKeeper.SetSearcherRoutes(suite.Ctx, "osmo", token, &searcherRoutes)
	}
}

// CreateRoute creates SearchRoutes object for testing
func CreateSeacherRoutes(numberPools uint64, tokenA, tokenB string) types.SearcherRoutes {
	routes := make([]*types.Route, numberPools)
	for i := uint64(0); i < numberPools; i++ {
		routes[i] = &types.Route{
			Pools: []uint64{i, i + 1, i + 2},
		}
	}

	return types.NewSearcherRoutes(routes, tokenA, tokenB)
}
