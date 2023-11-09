package usecase_test

import (
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

// This test validates that the pools in the route are converted into a new serializable
// type for clients with the following list of fields that are returned in each pool:
// - ID
// - Type
// - Balances
// - Spread Factor
// - Token Out Denom
// - Taker Fee
func (s *RouterTestSuite) TestPrepareResultPools() {
	s.Setup()

	balancerPoolID := s.PrepareBalancerPool()

	balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, balancerPoolID)
	s.Require().NoError(err)

	testcases := map[string]struct {
		route domain.Route

		expectedPools []domain.RoutablePool
	}{
		"empty route": {
			route: emptyRoute.DeepCopy(),

			expectedPools: []domain.RoutablePool{},
		},
		"single balancer pool in route": {
			route: withRoutePools(
				emptyRoute,
				[]domain.RoutablePool{
					withChainPoolModel(withTokenOutDenom(defaultPool, denomOne), balancerPool),
				},
			),

			expectedPools: []domain.RoutablePool{
				&usecase.RoutableResultPoolImpl{
					ID:            balancerPoolID,
					Type:          poolmanagertypes.Balancer,
					Balances:      defaultPoolBalances,
					SpreadFactor:  defaultSpreadFactor,
					TokenOutDenom: denomOne,
					TakerFee:      defaultTakerFee,
				},
			},
		},

		// TODO:
		// add tests with more pool types as well as multiple pools in the route
		// https://app.clickup.com/t/86a1cfwag
	}

	for name, tc := range testcases {
		tc := tc
		s.Run(name, func() {

			resultPools := tc.route.PrepareResultPools()

			s.validateRoutePools(tc.expectedPools, resultPools)
			s.validateRoutePools(tc.expectedPools, tc.route.GetPools())
		})
	}
}

// validateRoutePools validates that the expected pools are equal to the actual pools.
// Specifically, validates the following fields:
// - ID
// - Type
// - Balances
// - Spread Factor
// - Token Out Denom
// - Taker Fee
func (s *RouterTestSuite) validateRoutePools(expectedPools []domain.RoutablePool, actualPools []domain.RoutablePool) {

	s.Require().Equal(len(expectedPools), len(actualPools))

	for i, expectedPool := range expectedPools {
		actualPool := actualPools[i]

		expectedResultPool, ok := expectedPool.(*usecase.RoutableResultPoolImpl)
		s.Require().True(ok)

		// Cast to result pool
		actualResultPool, ok := actualPool.(*usecase.RoutableResultPoolImpl)
		s.Require().True(ok)

		s.Require().Equal(expectedResultPool.ID, actualResultPool.ID)
		s.Require().Equal(expectedResultPool.Type, actualResultPool.Type)
		s.Require().Equal(expectedResultPool.Balances.String(), actualResultPool.Balances.String())
		s.Require().Equal(expectedResultPool.SpreadFactor.String(), actualResultPool.SpreadFactor.String())
		s.Require().Equal(expectedResultPool.TokenOutDenom, actualResultPool.TokenOutDenom)
		s.Require().Equal(expectedResultPool.TakerFee.String(), actualResultPool.TakerFee.String())
	}
}
