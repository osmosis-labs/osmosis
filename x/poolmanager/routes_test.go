package poolmanager_test

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

// TestGetPoolModule tests that the correct pool module is returned for a given pool id.
// Additionally, validates that the expected errors are produced when expected.
func (s *KeeperTestSuite) TestDenomPairRoute() {
	tests := map[string]struct {
		poolId            uint64
		preCreatePoolType types.PoolType
		routesOverwrite   map[types.PoolType]types.PoolModuleI

		expectedModule reflect.Type
		expectError    error
	}{
		"valid balancer pool": {
			preCreatePoolType: types.Balancer,
			poolId:            1,
			expectedModule:    gammKeeperType,
		},
	}

	for name, _ := range tests {
		s.Run(name, func() {
			s.SetupTest()
			poolmanagerKeeper := s.App.PoolManagerKeeper

			s.PrepareConcentratedPool()
			pool := s.PrepareConcentratedPoolWithCoins("eth", "stake")
			pool1 := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("adam", "eth")
			pool2 := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("adam", "eth")
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(2000000000000000000)), sdk.NewCoin("adam", sdk.NewInt(2000000000000000000)))...)
			s.PrepareBalancerPool()

			s.CreateFullRangePosition(pool, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(1000000000000000000)), sdk.NewCoin("stake", sdk.NewInt(1000000000000000000))))
			s.CreateFullRangePosition(pool1, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(3000000000000000000)), sdk.NewCoin("adam", sdk.NewInt(3000000000000000000))))
			s.CreateFullRangePosition(pool2, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(2000000000000000000)), sdk.NewCoin("adam", sdk.NewInt(2000000000000000000))))

			test, err := poolmanagerKeeper.GetPoolLiquidityOfDenom(s.Ctx, 3, "eth")
			s.Require().NoError(err)

			fmt.Println("eth spply 4", test)

			test, err = poolmanagerKeeper.GetPoolLiquidityOfDenom(s.Ctx, 4, "adam")
			s.Require().NoError(err)

			fmt.Println("adam spply 5", test)

			test, err = poolmanagerKeeper.GetPoolLiquidityOfDenom(s.Ctx, 2, "stake")
			s.Require().NoError(err)

			fmt.Println("adam stake 2", test)

			//poolmanagerKeeper.SetDenomPairRoutes(s.Ctx)

			routes, err := poolmanagerKeeper.GetDenomPairRoute(s.Ctx, "adam", "stake")
			s.Require().NoError(err)

			fmt.Println(routes)
		})
	}
}
