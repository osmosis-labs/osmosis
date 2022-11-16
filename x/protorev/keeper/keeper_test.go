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
	suite.App.AppKeepers.ProtoRevKeeper.SetOsmoPool(suite.Ctx, "Juon", 2)
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
func (suite *KeeperTestSuite) SetUpRoutes() {
	tokens := []string{"Akash", "Juno", "Ethereum", "Bitcoin", "Canto"}

	// create routes with atom
	for _, token := range tokens {
		route := CreateRoute(token, 5)
		suite.App.AppKeepers.ProtoRevKeeper.SetRoute(suite.Ctx, types.AtomDenomination, token, &route)
	}

	// create routes with osmo
	for _, token := range tokens {
		route := CreateRoute(token, 5)
		suite.App.AppKeepers.ProtoRevKeeper.SetRoute(suite.Ctx, types.OsmosisDenomination, token, &route)
	}

}

// CreateRoute creates a route for testing
func CreateRoute(arbDenom string, numberPools uint64) types.Route {
	pools := make([]uint64, numberPools)

	var pool uint64
	for pool = 0; pool < numberPools; pool++ {
		pools[pool] = pool
	}

	return types.Route{
		ArbDenom: arbDenom,
		Pools:    pools,
	}
}
